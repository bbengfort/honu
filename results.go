package honu

import (
	"fmt"
	"io"
	"os"
	"time"
)

// ResultWriter specifies the interface for writing results
type ResultWriter interface {
	io.Writer
	io.Closer
	Sync() error
}

// Latency describes the round-trip time of a single message.
type Latency struct {
	Message uint64        // the id of the message sent
	Start   time.Time     // the time the message was sent
	Delay   time.Duration // the amount of time the message took
	Bytes   int           // the number of bytes in the message
	Key     string        // the key that was sent
	Version uint64        // the version of the key written
	Success bool          // if the write was successful or not
}

// String serializes the latency for writing.
// msgid,key,version,ts,latency,bytes,success
func (l *Latency) String() string {
	return fmt.Sprintf(
		"%d,%s,%d,%s,%d,%d,%t",
		l.Message,
		l.Key,
		l.Version,
		l.Start.Format(time.RFC3339Nano),
		l.Delay.Nanoseconds(),
		l.Bytes,
		l.Success,
	)
}

// Results instantiates a thread-safe writer to a file at a given path or to
// stdout if no path is specified. Returns a channel to queue results on.
func Results(path string) (chan<- *Latency, error) {
	// Make a large buffered results channel
	results := make(chan *Latency, 1000)

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
