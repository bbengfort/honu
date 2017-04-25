package honu

import (
	"fmt"
	"sync"
)

// NewStore creates and initializes a key value store
func NewStore() *Store {
	store := new(Store)
	store.namespace = make(map[string][]byte)
	store.versions = make(map[string]uint64)

	return store
}

// Store implements a versioned, in-memory key-value store. Note that this
// implementation does not maintain a version history, but keeps track of how
// many changes have occurred to a key's value in the store.
type Store struct {
	sync.RWMutex
	namespace map[string][]byte // maps keys to values
	versions  map[string]uint64 // maps keys to versions
}

// Get a value and version pair for a specific key. Returns a not found error
// if the key is not in the mapping or in the version space.
func (s *Store) Get(key string) (value []byte, version uint64, err error) {
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
func (s *Store) Put(key string, value []byte) (version uint64, err error) {
	s.Lock()
	defer s.Unlock()

	// Update the version information first.
	var ok bool
	version, ok = s.versions[key]
	if !ok {
		version = 1
	} else {
		version++
	}

	// Update the value in the namespace
	s.versions[key] = version
	s.namespace[key] = value

	// Return the version and no error for this method
	return version, nil
}
