package honu

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/honu/rpc"
	"github.com/bbengfort/x/stats"
	"google.golang.org/grpc"
)

const timeout = 10 * time.Second

// Client wraps information about throughput to the storage server, each
// client works with a single key and maintains information about the version
// of each key as it generates work.
type Client struct {
	key     string           // the key the client accesses
	version *Version         // current version of the key (must be increasing)
	addr    string           // the address of the server
	conn    *grpc.ClientConn // the connection to the server
	rpc     pb.StorageClient // the transport to make requests on
	metrics *stats.Benchmark // client-side latency benchmarks
}

// Connect creates the connection and rpc client to the server
func (c *Client) Connect(addr string) error {
	var err error

	c.addr = addr
	c.conn, err = grpc.Dial(
		c.addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout),
	)

	if err != nil {
		warn(err.Error())
		return fmt.Errorf("could not connect to %s: %s", c.addr, err)
	}

	debug("connected to storage server at %s", c.addr)
	c.rpc = pb.NewStorageClient(c.conn)
	return nil
}

// Close the connection to the server
func (c *Client) Close() error {
	if !c.IsConnected() {
		return errors.New("client is not connected, cannot close")
	}

	c.rpc = nil // nilify the rpc client to the server

	// close the connection
	if err := c.conn.Close(); err != nil {
		warn(err.Error())
		return err
	}

	info("connection to server at %s closed", c.addr)
	c.conn = nil
	return nil
}

// IsConnected verifies if the client is connected
func (c *Client) IsConnected() bool {
	if c.conn != nil && c.rpc != nil {
		return true
	}

	return false
}

// Get composes a Get Request and returns the value and version.
func (c *Client) Get(key string) ([]byte, string, error) {
	if !c.IsConnected() {
		return nil, "", errors.New("not connected, cannot make a request")
	}

	req := &pb.GetRequest{
		Key: key,
	}

	debug("send get %s", req.Key)
	reply, err := c.rpc.GetValue(context.Background(), req)

	if err != nil {
		warn(err.Error())
		return nil, "", err
	}

	if !reply.Success {
		warn(reply.Error)
		return nil, "", errors.New(reply.Error)
	}

	return reply.Value, reply.Version, nil
}

// Put composes a Put request and returns the version created.
func (c *Client) Put(key string, value []byte) (string, error) {
	if !c.IsConnected() {
		return "", errors.New("not connected, cannot make a request")
	}

	req := &pb.PutRequest{
		Key:   key,
		Value: value,
	}

	debug("send put %d bytes to %s", len(value), req.Key)
	reply, err := c.rpc.PutValue(context.Background(), req)

	if err != nil {
		warn(err.Error())
		return "", err
	}

	if !reply.Success {
		warn(reply.Error)
		return "", errors.New(reply.Error)
	}

	return reply.Version, nil
}
