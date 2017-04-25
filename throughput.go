package honu

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/honu/rpc"
)

// write sends a single Put request to the server to test the write throughput.
func (c *Client) write(done chan<- bool, echan chan<- error, results chan<- *Latency) {

	// Ensure the client is connected
	if !c.IsConnected() {
		echan <- errors.New("client is not connected to storage server")
		return
	}

	// Prepare the write
	c.messages++
	msg := fmt.Sprintf("msg %d at %s", c.messages, time.Now().Format(time.RFC3339Nano))
	req := &pb.PutRequest{
		Key:   c.key,
		Value: []byte(msg),
	}

	// Execute the write to the storage server
	start := time.Now()
	reply, err := c.rpc.PutValue(context.Background(), req)
	if err != nil {
		echan <- err
		return
	}

	// Compute the message latency asap
	latency := time.Since(start)
	c.latency += latency

	// Assert monotonically increasing version numbers
	if reply.Version <= c.version {
		echan <- errors.New("monotonically increasing version error")
		return
	}
	c.version = reply.Version

	// Send the result to be written to disk
	result := &Latency{
		Message: c.messages,
		Start:   start,
		Delay:   latency,
		Bytes:   len(req.Value),
		Key:     req.Key,
		Version: reply.Version,
		Success: reply.Success,
	}

	results <- result
	done <- true
}

// Run the write throughput workload
func (c *Client) Run(key string, duration time.Duration, outpath string) error {
	// Initialize the client
	c.key = key
	c.messages = 0
	c.version = 0
	c.latency = 0

	// Initialize the channels
	timer := time.NewTimer(duration)
	echan := make(chan error, 1)
	done := make(chan bool, 1)
	results, err := Results(outpath)

	if err != nil {
		return err
	}

	// Send the first write
	go c.write(done, echan, results)

	// Continue until the timer is complete
	for {
		select {
		case <-timer.C:
			results <- nil // close the results file
			throughput := float64(c.messages) / c.latency.Seconds()
			msg := fmt.Sprintf("%d messages sent in %s (%0.4f msg/sec)", c.messages, c.latency, throughput)
			fmt.Println(msg)
			return nil
		case err := <-echan:
			results <- nil // close the results file
			return err
		case <-done:
			go c.write(done, echan, results)
		}
	}

}
