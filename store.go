package honu

import (
	"fmt"
	"sync"
)

//===========================================================================
// Store is an interface for any key/value store and is created with NewStore
//===========================================================================

// NewStore creates and initializes a key value store
func NewStore(sequential bool) Store {
	if sequential {
		store := new(SequentialStore)
		store.namespace = make(map[string]*Value)
		info("created sequential consistency storage")
		return store
	}

	// The default is a linearizable store.
	store := new(LinearizableStore)
	store.namespace = make(map[string][]byte)
	store.versions = make(map[string]uint64)
	info("created linearizable consistency storage")
	return store

}

// Store is an interface for multiple in-memory storage types under the hood.
type Store interface {
	Get(key string) (value []byte, version uint64, err error)
	Put(key string, value []byte) (version uint64, err error)
}

//===========================================================================
// Storage with Linearizable Consistency
//===========================================================================

// LinearizableStore implements a versioned, in-memory key-value store.
// Note that this implementation does not maintain a version history, but
// keeps track of how many changes have occurred to a key's value.
type LinearizableStore struct {
	sync.RWMutex
	current   uint64            // the current version
	namespace map[string][]byte // maps keys to values
	versions  map[string]uint64 // maps keys to versions
}

// Get a value and version pair for a specific key. Returns a not found error
// if the key is not in the mapping or in the version space.
func (s *LinearizableStore) Get(key string) (value []byte, version uint64, err error) {
	s.RLock()
	defer s.RUnlock()

	var ok bool
	value, ok = s.namespace[key]
	if !ok {
		err = fmt.Errorf("key '%s' not found in namespace", key)
		return value, version, err
	}

	version, ok = s.versions[key]
	if !ok {
		err = fmt.Errorf("key '%s' not found in versions", key)
		return value, version, err
	}

	return value, version, err
}

// Put a value into the namespace and increment the version. Returns the
// version for the given key and any error that might occur.
func (s *LinearizableStore) Put(key string, value []byte) (version uint64, err error) {
	s.Lock()
	defer s.Unlock()

	// Update the version information first.
	s.current++

	// Update the value in the namespace
	s.versions[key] = s.current
	s.namespace[key] = value

	// Return the version and no error for this method
	return s.versions[key], nil
}

//===========================================================================
// Storage wtih Sequential Consistency
//===========================================================================

// Value implements a per-key value for each value.
type Value struct {
	sync.RWMutex
	version uint64
	value   []byte
}

// SequentialStore implements a key/value store where each key is versioned
// independently of all other keys. The Store is locked when a new key is
// added, but readers and writers take locks on individual keys afterward.
type SequentialStore struct {
	sync.RWMutex
	namespace map[string]*Value
}

// get is an internal method surrounded by a read lock that fetches the
// given value for a specific key.
func (s *SequentialStore) get(key string) *Value {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.namespace[key]
	if !ok {
		return nil
	}

	return value
}

// Get a value and version pair for a specific key. Returns a not found error
// if the key is not in the mapping or in the version space.
func (s *SequentialStore) Get(key string) (value []byte, version uint64, err error) {
	// Readlock to fetch the value
	val := s.get(key)

	// Handle not found error
	if val == nil {
		err = fmt.Errorf("key '%s' not found in namespace", key)
		return nil, 0, err
	}

	// Perform the read and ensure the val is read-locked
	val.RLock()
	defer val.RUnlock()
	return val.value, val.version, nil
}

// make is an internal method surrounded by a write lock that inserts new
// values into the namespace without conflicts. It returns a locked value.
func (s *SequentialStore) make(key string, value []byte) *Value {
	// Acquire a write lock
	s.Lock()
	defer s.Unlock()

	// Create a read locked value
	val := &Value{
		version: 1,
		value:   value,
	}
	val.Lock()

	// Insert the value into the namespace and return
	s.namespace[key] = val
	return val
}

// Put a value into the namespace and increment the version. Returns the
// version for the given key and any error that might occur.
func (s *SequentialStore) Put(key string, value []byte) (version uint64, err error) {
	var val *Value

	// Read lock to fetch the value
	val = s.get(key)

	// Make the value if it doesn't exist
	if val == nil {
		val = s.make(key, value)
	} else {
		// Acquire the write lock on the val
		val.Lock()
	}

	// Ensure the value is unlocked
	defer val.Unlock()

	// Perform the update
	val.value = value
	val.version++

	return val.version, nil
}
