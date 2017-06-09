package honu

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	pb "github.com/bbengfort/honu/rpc"
)

//===========================================================================
// Entry is a thin wrapper around a value in the key/value store
//===========================================================================

// Entry is a thin wrapper around values in the key/value store. It tracks
// meta data and is lockable for different types of consistency requirements.
type Entry struct {
	sync.RWMutex
	Key     *string  // The associated key with the entry
	Version *Version // The conflict-free version of the entry
	Parent  *Version // The version of the parent the entry was derived from
	Value   []byte   // The data value of the entry
	Current uint64   // The current version scalar
}

//===========================================================================
// Temporary to/from Protobuf
//===========================================================================

func (e *Entry) topb() *pb.Entry {
	return &pb.Entry{
		Parent:  e.Parent.topb(),
		Version: e.Version.topb(),
		Value:   e.Value,
	}
}

// not thread safe
func (e *Entry) frompb(in *pb.Entry) {
	e.Parent = new(Version)
	e.Version = new(Version)

	e.Parent.frompb(in.Parent)
	e.Version.frompb(in.Version)
	e.Value = in.Value
}

//===========================================================================
// Version struct and methods
//===========================================================================

// NullVersion is the zero value version that does not exist.
var NullVersion = Version{0, 0}

// Version implements conflict-free or concurrent versioning for objects.
type Version struct {
	Scalar uint64 // monotonically increasing scalar version number (starts at one)
	PID    uint64 // process identifier for tie-breaks (should not be zero)
}

// ParseVersion converts a version string into a version object.
func ParseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return NullVersion, fmt.Errorf("incorrect number of version components, could not parse '%s'", s)
	}

	scalar, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return NullVersion, fmt.Errorf("could not parse scalar component: '%s'", parts[0])
	}

	pid, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return NullVersion, fmt.Errorf("could not parse pid component: '%s'", parts[1])
	}

	return Version{scalar, pid}, nil
}

// String returns a parsable representation of the version number.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Scalar, v.PID)
}

// IsZero determines if a version is null
func (v Version) IsZero() bool {
	return v.Scalar == 0 && v.PID == 0
}

// Equals compares two *Versions to determine if they're identical.
func (v Version) Equals(o *Version) bool {
	if o == nil {
		return v.IsZero()
	}

	if v.Scalar == o.Scalar && v.PID == o.PID {
		return true
	}
	return false
}

// Greater returns true if the local version is later than the other version.
func (v Version) Greater(o *Version) bool {
	if o == nil {
		return true
	}

	if v.Scalar == o.Scalar {
		return v.PID > o.PID
	}
	return v.Scalar > o.Scalar
}

// GreaterEqual returns true if the local version is greater than or equal to
// the other version.
func (v Version) GreaterEqual(o *Version) bool {
	if o == nil {
		return true
	}

	if v.Scalar == o.Scalar {
		return v.PID >= o.PID
	}
	return v.Scalar > o.Scalar
}

// Lesser returns true if the local version is earlier than the other version.
func (v Version) Lesser(o *Version) bool {
	if o == nil {
		return false
	}

	if v.Scalar == o.Scalar {
		return v.PID < o.PID
	}
	return v.Scalar < o.Scalar
}

// LesserEqual returns true if the local version is less than or equal to the
// other version.
func (v Version) LesserEqual(o *Version) bool {
	if o == nil {
		return v.Equals(o)
	}

	if v.Scalar == o.Scalar {
		return v.PID <= o.PID
	}
	return v.Scalar < o.Scalar
}

//===========================================================================
// Temporary to/from Protobuf
//===========================================================================

func (v *Version) topb() *pb.Version {
	return &pb.Version{
		Scalar: v.Scalar,
		Pid:    v.PID,
	}
}

func (v *Version) frompb(in *pb.Version) {
	v.Scalar = in.Scalar
	v.PID = in.Pid
}

//===========================================================================
// Version Factory
//===========================================================================

// VersionFactory tracks version information and returns new versions on a
// per-key basis. Implements Lamport scalar versioning. Note that the factory
// is not thread-safe and should be used in a thread-safe object.
type VersionFactory struct {
	pid    uint64            // the current process id
	latest map[string]uint64 // map of keys to latest seen scalar
}

// Next creates and returns the next version for the given key.
func (f *VersionFactory) Next(key string) *Version {
	f.latest[key]++
	return &Version{
		Scalar: f.latest[key],
		PID:    f.pid,
	}
}

// Update the latest version with the version for the given key.
func (f *VersionFactory) Update(key string, vers *Version) {
	if vers.Scalar > f.latest[key] {
		f.latest[key] = vers.Scalar
	}
}
