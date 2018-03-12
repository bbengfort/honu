package honu

import (
	"encoding/json"
	"os"
	"time"
)

// VisibilityBufferSize describes the maximum number of async visiblity log
// statements before the caller will have to block.
const VisibilityBufferSize = 10000

// NewVisibilityLogger creates a logger for write visibility at the path.
func NewVisibilityLogger(path string) (*VisibilityLogger, error) {
	out, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	vl := &VisibilityLogger{
		file: out,
		err:  nil,
		msgs: make(chan *visibilityMessage, VisibilityBufferSize),
		done: make(chan bool),
	}

	go vl.flusher()
	return vl, nil
}

// VisibilityLogger records the time a write becomes visible on the local
// replica, storing the information on disk. It uses an asynchronous writer
// so it doesn't block other store operations.
type VisibilityLogger struct {
	file *os.File
	err  error
	msgs chan *visibilityMessage
	done chan bool
}

// simple data structure for storing visibility information.
type visibilityMessage struct {
	key       string
	version   string
	timestamp time.Time
}

// Log a Put to the key/value store
func (l *VisibilityLogger) Log(key, version string) {
	l.msgs <- &visibilityMessage{
		key: key, version: version, timestamp: time.Now(),
	}
}

// Close the logger and wait until it's done writing all buffered messages.
func (l *VisibilityLogger) Close() error {
	close(l.msgs)
	<-l.done

	if l.err != nil {
		return l.err
	}

	if err := l.file.Sync(); err != nil {
		return err
	}

	return l.file.Close()
}

// Error returns any issues the visibility logger had
func (l *VisibilityLogger) Error() error {
	return l.err
}

// routine that reads visibility log messages off the msgs channel and
// writes them to disk.
func (l *VisibilityLogger) flusher() {
	for msg := range l.msgs {
		var data []byte

		data, l.err = json.Marshal(msg)
		if l.err != nil {
			close(l.msgs)
			break
		}

		data = append(data, byte('\n'))

		_, l.err = l.file.Write(data)
		if l.err != nil {
			close(l.msgs)
			break
		}
	}

	l.done <- true
}
