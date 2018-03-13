package honu

import (
	"fmt"
	"time"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/honu/rpc"
	"github.com/bbengfort/x/stats"
	"golang.org/x/net/context"
)

// AntiEntropy performs a pairwise, bilateral syncrhonization with a random
// remote peer, first sending our version vector, then sending any required
// versions to the remote host.
//
// NOTE: the view specified is the view at the start of anti-entropy.
func (s *Server) AntiEntropy() {
	// Schedule the next anti-entropy session
	defer time.AfterFunc(s.delay, s.AntiEntropy)

	// Select a random peer for pairwise anti-entropy
	reward := 0.0
	arm := s.bandit.Select()
	peer := s.peers[arm]

	// Ensure we update the reward for the bandit when we are done.
	defer func() { s.bandit.Update(arm, reward) }()

	// TODO: do better at ignoring self-connections
	if peer == s.addr {
		// Penalize self selection by a lot
		reward = -1.0
		s.syncs[peer].Misses++
		return
	}

	// Create a connection to the client
	conn, err := grpc.Dial(
		peer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout),
	)

	if err != nil {
		s.syncs[peer].Misses++
		warn(err.Error())
		return
	}
	defer conn.Close()

	// Create a gossip client
	client := pb.NewGossipClient(conn)
	debug("connected to anti-entropy peer at %s", peer)

	// Get the current version vector for every object
	vector := s.store.View()

	// Create the pull request
	req := &pb.PullRequest{
		Versions: make(map[string]*pb.Version),
	}

	for key, version := range vector {
		req.Versions[key] = version.topb()
	}

	// Send the pull request
	pullStart := time.Now()
	rep, err := client.Pull(context.Background(), req)
	if err != nil {
		s.syncs[peer].Misses++
		warn(err.Error())
		return
	}
	pullLatency := time.Since(pullStart)
	s.syncs[peer].Update(pullLatency, "pull")

	// Handle the pull response
	if !rep.Success {
		s.syncs[peer].Misses++
		debug("no synchronization occurred")
		return
	}

	reward += 0.25 // add reward for a successful pull request

	// add reward for low latency pull requests
	if pullLatency < 5*time.Millisecond {
		reward += 0.20 // highest reward for local latencies
	} else if pullLatency <= 100*time.Millisecond {
		reward += 0.10 // reward for close by links that don't globe span.
	}

	s.syncs[peer].Pulls++
	var items uint64

	for key, pbentry := range rep.Entries {
		entry := new(Entry)
		entry.frompb(pbentry)
		if s.store.PutEntry(key, entry) {
			items++

			// Track visibility if requested
			if s.visibility != nil && entry.TrackVisibility {
				s.visibility.Log(key, entry.Version.String())
			}
		}
	}

	if items > 1 {
		// add reward for multi-items
		reward += 0.05
	}

	// Send the push request (bilateral)
	// Can be fire and forget if needed
	if len(rep.Pull.Versions) > 0 {
		debug("pushing %d versions back to %s", len(rep.Pull.Versions), peer)
		push := &pb.PushRequest{
			Entries: make(map[string]*pb.Entry),
		}

		for key := range rep.Pull.Versions {
			entry := s.store.GetEntry(key)
			push.Entries[key] = entry.topb()
			items++
		}

		reward += 0.25 // add reward for a push request

		if len(rep.Pull.Versions) > 1 {
			// add reward for multi-items
			reward += 0.05
		}

		s.syncs[peer].Pushes++
		pushStart := time.Now()
		client.Push(context.Background(), push)
		pushLatency := time.Since(pushStart)
		s.syncs[peer].Update(pushLatency, "push")

		// add reward for low latency pull requests
		if pushLatency < 5*time.Millisecond {
			reward += 0.20 // highest reward for local latencies
		} else if pushLatency <= 100*time.Millisecond {
			reward += 0.10 // reward for close by links that don't globe span.
		}
	}

	// Log anti-entropy success and metrics
	s.syncs[peer].Syncs++
	s.syncs[peer].Versions += items
	info("synchronized %d items to %s", items, peer)
}

//===========================================================================
// Server Gossip RPC methods
//===========================================================================

// Pull handles incoming push requests, comparing the object version with the
// current view of the server and returning a push reply with entries that are
// later than the remote and a pull request where the remote's versions are
// later. This method operates by read locking the entire store.
func (s *Server) Pull(ctx context.Context, in *pb.PullRequest) (*pb.PullReply, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	reply := &pb.PullReply{
		Entries: make(map[string]*pb.Entry),
		Pull: &pb.PullRequest{
			Versions: make(map[string]*pb.Version),
		},
	}

	for key, pbvers := range in.Versions {
		// Get the remote version
		version := new(Version)
		version.frompb(pbvers)

		// Get the latest version and compare with old version
		entry := s.store.GetEntry(key)

		// Compare versions to see which version is later
		// Excluded condition is if the versions are equal.
		if entry == nil || version.Greater(entry.Version) {

			// Remote is greater than our local, request it to be pushed.
			// First create the protobuf version
			var vers *pb.Version
			if entry == nil {
				vers = nil
			} else {
				vers = entry.Version.topb()
			}

			// Update the version scalar
			s.store.Update(key, version)

			// Add version to the response
			reply.Pull.Versions[key] = vers

		} else if entry != nil && entry.Version.Greater(version) {

			// Local is greater than the remote, send it on.
			reply.Entries[key] = entry.topb()

		}

	}

	// Set success on the reply if synchronization has occurred.
	if len(reply.Entries) > 0 || len(reply.Pull.Versions) > 0 {
		reply.Success = true
	}

	return reply, nil
}

// Push handles incoming push requests, accepting any entries in the request
// that are later than the current view. It returns success if any
// synchronization occurs, otherwise false for a late push.
func (s *Server) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {
	reply := &pb.PushReply{Success: false}

	for key, pbent := range in.Entries {
		entry := new(Entry)
		entry.frompb(pbent)

		if s.store.PutEntry(key, entry) {
			reply.Success = true

			// Track visibility if requested
			if s.visibility != nil && entry.TrackVisibility {
				s.visibility.Log(key, entry.Version.String())
			}
		}
	}

	return reply, nil
}

//===========================================================================
// Per-peer metrics for syncrhonization
//===========================================================================

// Syncs is a map of peer hostnames to their synchronization statistics.
type Syncs map[string]*SyncStats

// Serialize the syncs to save to JSON format.
func (s Syncs) Serialize() map[string]interface{} {
	data := make(map[string]interface{})
	for peer, stats := range s {
		data[peer] = stats.Serialize()
	}
	return data
}

// SyncStats represents per-peer pairwise metrics of synchronization.
type SyncStats struct {
	Syncs       uint64 // Total number of anti-entropy sessions between peers
	Pulls       uint64 // Number of successful pull exchanges between peers
	Pushes      uint64 // Number of successful push exchanges between peers
	Misses      uint64 // Number of unsuccessful exchanges between peers
	Versions    uint64 // The total number of object versions exchanged
	PullLatency *stats.Benchmark
	PushLatency *stats.Benchmark
	initialized bool
}

// Init the Syncstats to ensure it's ready for updating.
func (s *SyncStats) Init() {
	s.PullLatency = new(stats.Benchmark)
	s.PushLatency = new(stats.Benchmark)
	s.initialized = true
}

// Update the latency of the given type
func (s *SyncStats) Update(latency time.Duration, method string) error {
	if !s.initialized {
		s.Init()
	}

	switch method {
	case "pull":
		s.PullLatency.Update(latency)
	case "push":
		s.PushLatency.Update(latency)
	default:
		return fmt.Errorf("no method '%s'", method)
	}

	return nil
}

// Serialize the SyncStats to write to disk
func (s *SyncStats) Serialize() map[string]interface{} {
	if !s.initialized {
		s.Init()
	}

	data := make(map[string]interface{})
	data["Syncs"] = s.Syncs
	data["Pulls"] = s.Pulls
	data["Pushes"] = s.Pushes
	data["Misses"] = s.Misses
	data["Versions"] = s.Versions
	data["PullLatency"] = s.PullLatency.Serialize()
	data["PushLatency"] = s.PushLatency.Serialize()
	return data
}
