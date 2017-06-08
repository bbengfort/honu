package honu

//===========================================================================
// Version Chain and Consistency Analysis
//===========================================================================

// VersionNode is similar to an Entry, but omits the value, allowing for a
// compact version chain that can be stored in memory, written to disk and
// otherwise evaluated as needed.
type VersionNode struct {
	Key     *string
	Parent  *Version
	Version *Version
}

// History keeps track of the version linkage (but does not store values) in
// order to evaluate the consistency of one model or another. It stores all
// versions in a single array, serializing appends via a channel that allows
// multiple go routines to stream version information to the history.
type History struct {
	versions []*VersionNode    // The array of version tree nodes in the chain
	queue    chan *VersionNode // The queue of entries to ad to the history
}

// Init the history with a buffered channel and dynamic array.
func (h *History) Init() {
	h.versions = make([]*VersionNode, 0, 1000)
	h.queue = make(chan *VersionNode, 1000)
}

// Run the history to continually pull entries off the queue, create version
// tree nodes and add them to the ordered version history.
func (h *History) Run() {
	go func() {
		for {
			node := <-h.queue
			h.versions = append(h.versions, node)
		}
	}()
}

// Append an entry to the version history.
func (h *History) Append(key *string, parent, version *Version) {
	node := &VersionNode{
		Key:     key,
		Parent:  parent,
		Version: version,
	}
	h.queue <- node
}
