package honu

import (
	"fmt"
	"sync"
)

//===========================================================================
// Store is an interface for any key/value store and is created with NewStore
//===========================================================================

// NewStore creates and initializes a key value store
func NewStore(pid uint64, sequential bool) Store {
	var store Store

	// Create the type-specific data structures
	if sequential {
		// Create a sequential store on demand.
		store = new(SequentialStore)
		info("created sequential consistency storage")
	} else {
		// The default is a linearizable store.
		store = new(LinearizableStore)
		info("created linearizable consistency storage")
	}

	// Initialize the store and return
	store.Init(pid)
	return store

}

// Store is an interface for multiple in-memory storage types under the hood.
type Store interface {
	Init(pid uint64)                                          // Initialize the store
	Get(key string) (value []byte, version string, err error) // Get a value and version for a given key
	Put(key string, value []byte) (version string, err error) // Put a value for a given key and get associated version
	Update(key string, version Version)                       // Update the version scalar from a remote source
	Snapshot(path string) error                               // Write a snapshot of the version history to disk
}

//===========================================================================
// Storage with Linearizable Consistency
//===========================================================================

// LinearizableStore implements a versioned, in-memory key-value store that
// keeps a single montonically increasing counter across all objects such
// that a single ordering for all writes (and associated reads) exists. All
// accesses are guarded by read and write locks to ensure linearizable
// consistency and version parents are the last written object no matter the
// key to create a cross-object version history.
type LinearizableStore struct {
	sync.RWMutex
	pid       uint64            // the local process id
	current   uint64            // the current version scalar
	lastWrite *Version          // the version of the last write
	namespace map[string]*Entry // maps keys to the latest entry
	history   *History          // tracks the verion history chain
}

// Init the store creating the internal data structures.
func (s *LinearizableStore) Init(pid uint64) {
	s.pid = pid
	s.namespace = make(map[string]*Entry)

	// Create, initialize and run the history.
	s.history = new(History)
	s.history.Init()
	s.history.Run()
}

// Get the most recently seen value and version pair for a specific key.
// Returns a not found error if the key has not been written to the namespace.
//
// This operation wraps the entire store in a read lock, meaning that other
// values can be read but no values can be written on Get.
func (s *LinearizableStore) Get(key string) (value []byte, version string, err error) {
	s.RLock()
	defer s.RUnlock()

	entry, ok := s.namespace[key]
	if !ok {
		err = fmt.Errorf("key '%s' not found in namespace", key)
		return value, version, err
	}

	version = entry.Version.String()
	value = entry.Value
	return value, version, err
}

// Put a value into the namespace, incrementing the version across all
// objects. This operation creates an entry whose parent is the last written
// version of any object. Put also stores all versions and associated entries,
// maintaining a complete version history.
//
// This operation locks the entire store, waiting for all read locks to be
// released and not allowing any other read or write locks until complete.
func (s *LinearizableStore) Put(key string, value []byte) (string, error) {
	s.Lock()
	defer s.Unlock()

	// Create the new version
	s.current++
	version := &Version{s.current, s.pid}

	// Create the new entry
	entry := &Entry{
		Key:     &key,
		Version: version,
		Parent:  s.lastWrite,
		Value:   value,
	}

	// Update the namespace, versions, and last write
	s.namespace[key] = entry
	s.history.Append(entry.Key, entry.Parent, entry.Version)
	s.lastWrite = version

	// Return the version and no error for this method
	return version.String(), nil
}

// Update the current version counter with the global value.
func (s *LinearizableStore) Update(key string, version Version) {
	s.Lock()
	defer s.Unlock()

	if version.Scalar > s.current {
		s.current = version.Scalar
	}
}

// Snapshot the current version history to disk, writing the version data to
// the specified path. Returns any I/O errors if snapshotting is unsuccessful.
func (s *LinearizableStore) Snapshot(path string) error {
	return fmt.Errorf("not implemented yet")
}

//===========================================================================
// Storage wtih Sequential Consistency
//===========================================================================

// SequentialStore implements a key/value store where each key is versioned
// independently of all other keys. The Store is only locked when a new key is
// added, but readers and writers take locks on individual keys afterward.
// A sequential store therefore allows multi-thread access to different
// objects simultaneously.
//
// The version history for objects in a sequential store is therefore relative
// to the object itself. Parent versions of entries are simply the previous
// entry in the store. Each object has its own independent scalar component.
type SequentialStore struct {
	sync.RWMutex
	pid       uint64            // the local process id
	namespace map[string]*Entry // maps keys to the latest entry
	history   *History          // tracks the verion history chain

}

// Init the store creating the internal data structures.
func (s *SequentialStore) Init(pid uint64) {
	s.pid = pid
	s.namespace = make(map[string]*Entry)

	// Create, initialize and run the history.
	s.history = new(History)
	s.history.Init()
	s.history.Run()
}

// get is an internal method surrounded by a read lock that fetches the
// given value for a specific key. It returns a locked entry, if the mutable
// flag is true, it is write locked, otherwise it is read locked.
//
// NOTE: callers must unlock the entry themselves!
func (s *SequentialStore) get(key string, mutable bool) *Entry {
	s.RLock()
	defer s.RUnlock()

	// Get the entry from the namespace
	entry, ok := s.namespace[key]
	if !ok {
		return nil
	}

	// Lock the entry according the mutable flag
	if mutable {
		entry.Lock()
	} else {
		entry.RLock()
	}

	// Return the locked entry
	return entry
}

// Get the most recently seen value and version pair for a specific key.
// Returns a not found error if the key has not been written to the namespace.
//
// This operation only locks the store with a read-lock on fetch but also adds
// a read-lock to the entry so that it cannot be modified in flight.
func (s *SequentialStore) Get(key string) (value []byte, version string, err error) {
	// Fetch the value, read-locking the entire store
	entry := s.get(key, false)

	// Handle not found error
	if entry == nil {
		err = fmt.Errorf("key '%s' not found in namespace", key)
		return nil, "", err
	}

	// Ensure that the entry is unlocked before we're done
	defer entry.RUnlock()

	// Extract the data required from the entry.
	return entry.Value, entry.Version.String(), nil
}

// make is an internal method that surrounds the store in a write lock to
// create an empty entry for the given key. It returns a write locked entry to
// ensure that the caller can update the entry with values before unlock but
// releases the store as soon as possible to prevent write delays.
//
// NOTE: this method should not be called if the key already exists!
// NOTE: callers must unlock the entry themselves!
func (s *SequentialStore) make(key string) *Entry {
	// Acquire a write lock
	s.Lock()
	defer s.Unlock()

	// Create a write locked entry
	entry := &Entry{Key: &key}
	entry.Lock()

	// Insert the entry into the namespace and return it
	s.namespace[key] = entry
	return entry
}

// Put a value into the namespace and increment the version. Returns the
// version for the given key and any error that might occur.
func (s *SequentialStore) Put(key string, value []byte) (string, error) {
	// Attempt to get the write-locked version from the store
	entry := s.get(key, true)

	// Make an empty entry if there was no entry already in the store
	if entry == nil {
		entry = s.make(key)
	} else {
		// Update the parent of the entry to the old entry
		entry.Parent = entry.Version
	}

	// Ensure that the entry is unlocked when done
	defer entry.Unlock()

	// Create the version for the new entry
	entry.Current++
	entry.Version = &Version{entry.Current, s.pid}

	// Update the value
	entry.Value = value

	// Store the version in the version history and return it
	s.history.Append(entry.Key, entry.Parent, entry.Version)
	return entry.Version.String(), nil
}

// Update the current version counter with the global value.
func (s *SequentialStore) Update(key string, version Version) {
	entry := s.get(key, true)
	defer entry.Unlock()

	if version.Scalar > entry.Current {
		entry.Current = version.Scalar
	}
}

// Snapshot the current version history to disk, writing the version data to
// the specified path. Returns any I/O errors if snapshotting is unsuccessful.
func (s *SequentialStore) Snapshot(path string) error {
	return fmt.Errorf("not implemented yet")
}
