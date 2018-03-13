package honu

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	pb "github.com/bbengfort/honu/rpc"
	"github.com/bbengfort/x/stats"
)

// Benchmark runs clients sending continuous accesses to the remote server.
type Benchmark struct {
	workers int
	clients []*Client
	extra   map[string]interface{}
	metrics *stats.Benchmark
}

// NewBenchmark creates the data structure and clients.
func NewBenchmark(workers int, prefix string, visibility bool, extra map[string]interface{}) (*Benchmark, error) {
	b := new(Benchmark)
	b.workers = workers
	b.clients = make([]*Client, 0, workers)

	b.extra = extra
	b.extra["workers"] = workers
	b.extra["prefix"] = prefix
	b.extra["version"] = PackageVersion
	b.extra["timestamp"] = time.Now().Format(time.RFC3339)

	for i := 0; i < workers; i++ {
		client := new(Client)
		client.visibility = visibility

		// Generate a key with specified prefix
		if len(prefix) > 1 {
			client.key = prefix
		} else {
			client.key = generateKey(prefix)
		}

		b.clients = append(b.clients, client)
	}

	return b, nil
}

// Run the benchmark with the specified duration
func (b *Benchmark) Run(addr, outpath string, duration, delay, rate time.Duration) error {
	if delay > 0 {
		status("delaying benchmark for %s", delay)
		time.Sleep(delay)
	}

	status("starting throughput benchmark with %d clients for %s", b.workers, duration)

	group := new(errgroup.Group)
	for _, client := range b.clients {
		c := client
		group.Go(func() error { return c.Run(addr, duration, rate) })
	}

	if err := group.Wait(); err != nil {
		return err
	}

	if err := b.Results(outpath); err != nil {
		return err
	}

	status("benchmark complete: %s", b)
	return nil
}

// Results writes the results to disk.
func (b *Benchmark) Results(path string) error {
	latencies := make([]float64, 0)
	b.metrics = new(stats.Benchmark)
	for _, client := range b.clients {
		b.metrics.Append(client.metrics)
	}

	// Write the results to disk
	if path != "" {
		data := b.metrics.Serialize()
		for key, val := range b.extra {
			data[key] = val
		}

		if len(latencies) > 0 {
			data["latencies"] = latencies
		}

		return appendJSON(path, data)
	}
	return nil
}

// String returns the metrics results
func (b *Benchmark) String() string {
	return fmt.Sprintf(
		"%d accesses (%d timeouts): %0.3f accesses/second",
		b.metrics.N(), b.metrics.Timeouts(), b.metrics.Throughput(),
	)
}

//===========================================================================
// Client Run function
//===========================================================================

// Access sends a put request, measuring latency.
func (c *Client) Access(done chan<- bool, echan chan<- error, rate time.Duration) {
	defer func() {
		// If we're rate limited, then wait a bit before we return
		if rate > 0 {
			time.Sleep(rate)
		}
		done <- true
	}()

	// Create a string value for the object
	val := fmt.Sprintf(
		"Put msg %d to object %s",
		c.metrics.N()+1, c.key,
	)

	// Create the Put request
	req := &pb.PutRequest{
		Key:             c.key,
		Value:           []byte(val),
		TrackVisibility: c.visibility,
	}

	// Send the request
	start := time.Now()
	rep, err := c.rpc.PutValue(context.Background(), req)
	if err != nil {
		echan <- err
		return
	}

	// Compute the latency ASAP on successful message
	if rep.Success {
		delta := time.Since(start)
		c.metrics.Update(delta)
		return
	}

	// Otherwise return the error
	if rep.Error != "" {
		caution("could not Put '%s': %s", c.key, rep.Error)
	} else {
		caution("could not Put '%s': unknown error occurred", c.key)
	}

}

// Run a continuous access client for the specified duration.
func (c *Client) Run(addr string, duration, rate time.Duration) error {
	c.metrics = new(stats.Benchmark)

	if err := c.Connect(addr); err != nil {
		return err
	}

	timer := time.NewTimer(duration)
	echan := make(chan error, 1)
	done := make(chan bool, 1)

	// Kick off the first access
	go c.Access(done, echan, rate)

	// Keep accessing until time runs out or something goes down
	for {
		select {
		case <-timer.C:
			// Done!
			return nil
		case err := <-echan:
			return err
		case <-done:
			go c.Access(done, echan, rate)
		}
	}
}
