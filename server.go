package honu

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/honu/rpc"
	"golang.org/x/net/context"
)

//===========================================================================
// Honu server implementation
//===========================================================================

// DefaultAddr that the honu server listens on.
const DefaultAddr = ":3264"

// NewServer creates and initializes a server.
func NewServer(pid uint64, sequential bool) *Server {
	server := new(Server)
	server.store = NewStore(pid, sequential)

	// Save the server type for analytics
	// TODO: refactor to use reflect to check the name of the struct.
	if sequential {
		server.stype = "sequential"
	} else {
		server.stype = "linearizable"
	}

	return server
}

// Server responds to Get and Put requests, modifying an in-memory store
// in a thread-safe fashion (because the store is surrounded by locks).
type Server struct {
	sync.Mutex
	store      Store             // The in-memory key/value store
	addr       string            // The IP address of the local server
	peers      []string          // IP addresses of replica peers
	delay      time.Duration     // The anti-entropy delay
	stype      string            // The type of storage being used
	started    time.Time         // The time the first message was received
	finished   time.Time         // The time of the last message to be received
	reads      uint64            // The number of reads to the server
	writes     uint64            // The number of writes to the server
	syncs      Syncs             // Per-peer metrics of anti-entropy synchronizations
	bandit     BanditStrategy    // Peer selection bandit strategy
	stats      string            // Path to write metrics to
	history    string            // Path to write version history to
	visibility *VisibilityLogger // Track the visibility of writes
}

//===========================================================================
// Run the server
//===========================================================================

// Run the Honu server.
func (s *Server) Run(addr string) error {
	// Use the default address to run on if one isn't specified.
	if addr == "" {
		addr = DefaultAddr
	}

	// Store the addr on the server
	s.addr = addr

	// Create the TCP channel to receive connections
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err.Error())
	}

	// Create the gRPC handler for RPC messages
	srv := grpc.NewServer()
	pb.RegisterStorageServer(srv, s)
	pb.RegisterGossipServer(srv, s)

	// Capture interrupt and shutdown gracefully
	go signalHandler(s.Shutdown)

	status("honu storage server listening on %s", addr)
	srv.Serve(lis)

	return nil
}

// Uptime sets a fixed amount of time to keep the server up for, shutting it
// down when the duration has passed and exiting gracefully.
func (s *Server) Uptime(d time.Duration) {
	time.AfterFunc(d, func() {
		defer os.Exit(0)

		info("shutting down server after %s uptime", d)
		s.Shutdown()
	})
}

// Visibility opens the visibility logger at the specified path.
func (s *Server) Visibility(path string) (err error) {
	s.visibility, err = NewVisibilityLogger(path)
	return err
}

// Measure the Honu server activity on shutdown. Pass in the paths to write
// stats and history to on shutdown. If empty strings, they will be ignored.
func (s *Server) Measure(stats, history string) {
	s.stats = stats
	s.history = history
}

// Replicate the Honu server using anti-entropy.
func (s *Server) Replicate(peers []string, delay time.Duration, strategy string, epsilon float64) error {
	// Store the peers and delay on the server
	s.peers = peers
	s.delay = delay

	// Create the peer selection strategy
	strategy = strings.ToLower(strategy)
	switch strategy {
	case "uniform":
		s.bandit = new(Uniform)
	case "epsilon":
		s.bandit = &EpsilonGreedy{Epsilon: epsilon}
	case "annealing":
		s.bandit = new(AnnealingEpsilonGreedy)
	default:
		return fmt.Errorf("no peer selection bandit strategy named  %s", strategy)
	}

	// Initialize the bandit with the number of cases
	s.bandit.Init(len(s.peers))

	// Create the sync stats objects for each peer
	s.syncs = make(map[string]*SyncStats)
	for _, peer := range peers {
		s.syncs[peer] = new(SyncStats)
	}

	// Schedule the anti-entropy delay
	time.AfterFunc(s.delay, s.AntiEntropy)

	// Give notice and return no error
	info("replicating to %d peers with anti-entropy interval %s", len(peers), delay)
	return nil
}

// Shutdown the Huno server, printing metrics.
func (s *Server) Shutdown() error {
	// Save the version history snapshot
	if s.history != "" {
		if err := s.store.Snapshot(s.history); err != nil {
			warn(err.Error())
		} else {
			info("version history snapshot saved to %s", s.history)
		}

	}

	// Save the results stats to disk for analysis
	if err := s.Metrics(s.stats); err != nil {
		warn(err.Error())
	} else {
		if s.stats != "" {
			info("server stats saved to %s", s.stats)
		}
	}

	return nil
}

//===========================================================================
// Server RPC methods
//===========================================================================

// GetValue implements the RPC for a get request from a client.
func (s *Server) GetValue(ctx context.Context, in *pb.GetRequest) (*pb.GetReply, error) {
	// Keep tracks of metrics with enter and exit
	s.enter("read")
	defer s.exit()

	reply := new(pb.GetReply)
	reply.Key = in.Key

	var err error
	reply.Value, reply.Version, err = s.store.Get(in.Key)
	if err != nil {
		warn(err.Error())
		reply.Success = false
		reply.Error = err.Error()
	} else {
		reply.Success = true
		debug("get key %s returns version %s", reply.Key, reply.Version)
	}

	return reply, nil
}

// PutValue implements the RPC for a put request from a client.
func (s *Server) PutValue(ctx context.Context, in *pb.PutRequest) (*pb.PutReply, error) {
	// Keep tracks of metrics with enter and exit
	s.enter("write")
	defer s.exit()

	reply := new(pb.PutReply)
	reply.Key = in.Key

	var err error
	reply.Version, err = s.store.Put(in.Key, in.Value)
	if err != nil {
		warn(err.Error())
		reply.Success = false
		reply.Error = err.Error()
	} else {
		reply.Success = true
		debug("put key %s to version %s", reply.Key, reply.Version)
	}

	// Track visibility if requested
	if err == nil && s.visibility != nil && in.TrackVisibility {
		s.visibility.Log(in.Key, reply.Version)
	}

	return reply, nil
}

//===========================================================================
// Server metrics
//===========================================================================

// Metrics writes server-side statistics as a JSON line to the specified path
// on disk. This function also logs the overall metrics (usually on shutdown)
// so if the path is an empty string, the metrics can be reported to the log
// without being saved to disk.
func (s *Server) Metrics(path string) error {
	s.Lock()
	defer s.Unlock()

	// Compute the final metrics
	var throughput float64
	duration := s.finished.Sub(s.started)
	accesses := s.reads + s.writes
	if accesses > 0 && duration > 0 {
		throughput = float64(accesses) / duration.Seconds()
	}

	var syncs uint64
	for _, stats := range s.syncs {
		syncs += stats.Syncs
	}

	// Log the metrics
	if accesses > 0 {
		status(
			"%d accesses (%d reads, %d writes) in %s -- %0.4f accesses/second",
			accesses, s.reads, s.writes, duration, throughput,
		)
	}

	status(
		"stored %d items after %d successful synchronizations",
		s.store.Length(), syncs,
	)

	// Compose the metrics to write to the given path.
	if path != "" {
		// Create the JSON data to write to disk
		data := make(map[string]interface{})
		data["started"] = s.started
		data["finished"] = s.finished
		data["timestamp"] = time.Now()
		data["duration"] = duration.Seconds()
		data["reads"] = s.reads
		data["writes"] = s.writes
		data["throughput"] = throughput
		data["store"] = s.stype
		data["nkeys"] = s.store.Length()
		data["syncs"] = s.syncs.Serialize()
		data["bandit"] = s.bandit.Serialize()
		data["peers"] = s.peers
		data["host"] = s.addr

		// Now write that data to disk
		if err := appendJSON(path, data); err != nil {
			return fmt.Errorf("could not append server metrics to %s: %s", path, err)
		}
		status("metrics written to %s", path)
	}

	return nil
}

// enter is called when an RPC method is started, it updates the count of the
// number of messages as well as tracks the start time of the steady state.
func (s *Server) enter(method string) {
	s.Lock()
	defer s.Unlock()

	if s.started.IsZero() {
		s.started = time.Now()
	}

	switch method {
	case "read":
		s.reads++
	case "write":
		s.writes++
	}
}

// exit is called when an RPC method is complete, it updates the end time of
// the steady state to measure the amount of throughput on the server.
func (s *Server) exit() {
	s.Lock()
	defer s.Unlock()

	s.finished = time.Now()
}
