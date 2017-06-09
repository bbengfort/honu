package honu

import (
	"math/rand"
	"time"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/honu/rpc"
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
	// TODO: update probabilities and message counts
	peer := s.peers[rand.Intn(len(s.peers))]

	// TODO: do better at ignoring self-connections
	if peer == s.addr {
		return
	}

	// Create a connection to the client
	conn, err := grpc.Dial(
		peer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout),
	)

	if err != nil {
		warn(err.Error())
		return
	}

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
	rep, err := client.Pull(context.Background(), req)
	if err != nil {
		warn(err.Error())
		return
	}

	// Handle the pull response
	if !rep.Success {
		debug("no synchronization occurred")
		return
	}

	var items uint64

	for key, pbentry := range rep.Entries {
		entry := new(Entry)
		entry.frompb(pbentry)
		if s.store.PutEntry(key, entry) {
			items++
		}
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

		go func() {
			client.Push(context.Background(), push)
		}()
	}

	// Log anti-entropy success
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
				vers = &pb.Version{0, 0}
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

// Push handles incoming push reqeuests, accepting any entries in the request
// that are later than the current view. It returns success if any
// synchronization occurs, otherwise false for a late push.
func (s *Server) Push(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {
	reply := &pb.PushReply{Success: false}

	for key, pbent := range in.Entries {
		entry := new(Entry)
		entry.frompb(pbent)

		if s.store.PutEntry(key, entry) {
			reply.Success = true
		}
	}

	return reply, nil
}
