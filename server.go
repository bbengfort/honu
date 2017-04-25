package honu

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/honu/rpc"
	"golang.org/x/net/context"
)

// NewServer creates and initializes a server.
func NewServer() *Server {
	server := new(Server)
	server.store = NewStore()

	return server
}

// Server responds to Get and Put requests, modifying an in-memory store
// in a thread-safe fashion (because the store is surrounded by locks).
type Server struct {
	store *Store // The in-memory key/value store
}

// Run the Honu server.
func (s *Server) Run(addr string) error {
	if addr == "" {
		addr = ":3264"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err.Error())
	}

	srv := grpc.NewServer()
	pb.RegisterStorageServer(srv, s)

	info("honu storage server listening on %s", addr)
	srv.Serve(lis)

	return nil
}

// GetValue implements the RPC for a get request from a client.
func (s *Server) GetValue(ctx context.Context, in *pb.GetRequest) (*pb.GetReply, error) {
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
		info("get key %s returns version %d", reply.Key, reply.Version)
	}

	return reply, nil
}

// PutValue implements the RPC for a put request from a client.
func (s *Server) PutValue(ctx context.Context, in *pb.PutRequest) (*pb.PutReply, error) {
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
		info("put key %s to version %d", reply.Key, reply.Version)
	}

	return reply, nil
}
