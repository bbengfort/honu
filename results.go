package honu

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

//===========================================================================
// Results Aggregation Helpers
//===========================================================================

// Helper function to append json data as a one line string to the end of a
// results file without deleting the previous contents in it.
func appendJSON(path string, val interface{}) error {
	// Open the file for appending, creating it if necessary
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Marshal the JSON in one line without indents
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// Append a newline to the data
	data = append(data, byte('\n'))

	// Append the data to the file
	_, err = f.Write(data)
	return err
}

//===========================================================================
// Streaming writes of server-side latency
//===========================================================================

// Latency describes the round-trip time of a single message.
type Latency struct {
	Message uint64        // the id of the message sent
	Start   time.Time     // the time the message was sent
	Delay   time.Duration // the amount of time the message took
	Bytes   int           // the number of bytes in the message
	Key     string        // the key that was sent
	Version string        // the version of the key written
	Success bool          // if the write was successful or not
}

// String serializes the latency for writing.
// msgid,key,version,ts,latency,bytes,success
func (l *Latency) String() string {
	return fmt.Sprintf(
		"%d,%s,%s,%s,%d,%d,%t",
		l.Message,
		l.Key,
		l.Version,
		l.Start.Format(time.RFC3339Nano),
		l.Delay.Nanoseconds(),
		l.Bytes,
		l.Success,
	)
}

// ResultWriter specifies the interface for writing results
type ResultWriter interface {
	io.Writer
	io.Closer
	Sync() error
}

// Results instantiates a thread-safe writer to a file at a given path or to
// stdout if no path is specified. Returns a channel to queue results on.
// If aggregate is true then the Results go not to a file, but rather to an
// aggregator that summarizes the latency statistics and appends them to the
// specified file in JSON form.
func Results(path string, aggregate bool) (chan<- *Latency, error) {
	// Make a large buffered results channel
	results := make(chan *Latency, 1000)

	// Create the aggregator if aggregate is true
	if aggregate {
		go Aggregator(results, path)
		return results, nil
	}

	// Otherwise create the stream writer
	var err error
	var stream ResultWriter

	if path != "" {
		stream, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	} else {
		stream = os.Stdout
	}

	// Execute the writer object
	go Writer(results, stream)
	return results, err
}

// Writer listens to a results channel and writes them to the writer.
func Writer(results <-chan *Latency, stream ResultWriter) {
	defer stream.Sync()
	defer stream.Close()

	// Write the header
	hdr := "msg,key,version,timestamp,latency (ns),bytes,success\n"
	if _, err := stream.Write([]byte(hdr)); err != nil {
		return
	}

	// Keep writing as long as we are receiving results
	for {
		result := <-results
		if result == nil {
			return
		}

		if _, err := stream.Write([]byte(result.String() + "\n")); err != nil {
			return
		}
	}
}

// Aggregator listens to a results channel and aggregates the results to disk
// when it sees the nil result come through the stream.
func Aggregator(results <-chan *Latency, path string) {

	// Keep track of latency statistics
	var messages uint64
	var delay float64
	var delaysq float64
	var bytes uint64
	var successes uint64
	var failures uint64
	versions := make(map[string]string)

	// Defer saving the files to disk on close
	defer func() {
		// Create the aggregated metrics to write to disk
		data := make(map[string]interface{})
		data["messages"] = messages

		// Compute the delay distribution
		n := float64(messages)
		delaydist := make(map[string]float64)
		delaydist["total"] = delay
		delaydist["mean"] = delay / n
		delaydist["stddev"] = math.Sqrt((n*delaysq - delay*delay) / (n * (n - 1)))
		data["latency"] = delaydist

		// Compute the throughput
		data["throughput"] = n / (delay / 1e9)

		// Add the total number of bytes and ratio of success to failure
		data["bytes"] = bytes
		data["successes"] = successes
		data["failures"] = failures

		// Store the latest seen version for each key
		data["versions"] = versions

		// Write the data to disk
		if err := appendJSON(path, data); err != nil {
			warn(err.Error())
		}
	}()

	// Handle streaming aggregation
	for {
		result := <-results
		if result == nil {
			return
		}

		messages++
		latency := float64(result.Delay.Nanoseconds())
		delay += latency
		delaysq += (latency * latency)
		bytes += uint64(result.Bytes)
		if result.Success {
			successes++
		} else {
			failures++
		}
		versions[result.Key] = result.Version
	}

}
