package honu

import (
	"fmt"
	"net"
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
func NewServer(sequential bool) *Server {
	server := new(Server)
	server.store = NewStore(sequential)

	return server
}

// Server responds to Get and Put requests, modifying an in-memory store
// in a thread-safe fashion (because the store is surrounded by locks).
type Server struct {
	sync.Mutex
	store    Store     // The in-memory key/value store
	started  time.Time // The time the first message was received
	finished time.Time // The time of the last message to be received
	reads    uint64    // The number of reads to the server
	writes   uint64    // The number of writes to the server
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

	// Create the TCP channel to receive connections
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err.Error())
	}

	// Create the gRPC handler for RPC messages
	srv := grpc.NewServer()
	pb.RegisterStorageServer(srv, s)

	// Capture interrupt and shutdown gracefully
	go signalHandler(s.Shutdown)

	info("honu storage server listening on %s", addr)
	srv.Serve(lis)

	return nil
}

// Shutdown the Huno server, printing metrics.
func (s *Server) Shutdown() error {
	info(s.Metrics())
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
		debug("get key %s returns version %d", reply.Key, reply.Version)
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
		debug("put key %s to version %d", reply.Key, reply.Version)
	}

	return reply, nil
}

//===========================================================================
// Server metrics
//===========================================================================

// Metrics returns a string representation of the throughput of the server.
func (s *Server) Metrics() string {
	s.Lock()
	defer s.Unlock()

	var throughput float64
	duration := s.finished.Sub(s.started)
	accesses := s.reads + s.writes
	if accesses > 0 && duration > 0 {
		throughput = float64(accesses) / duration.Seconds()
	}

	return fmt.Sprintf(
		"%d accesses (%d reads, %d writes) in %s -- %0.4f accesses/second",
		accesses, s.reads, s.writes, duration, throughput,
	)
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
