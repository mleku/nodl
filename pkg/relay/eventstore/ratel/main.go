package ratel

import (
	"encoding/binary"
	"sync"
	"time"

	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/serial"
	"util.mleku.dev/context"
	"github.com/dgraph-io/badger/v4"
)

type T struct {
	Ctx  context.T
	WG   *sync.WaitGroup
	Path string
	// DBSizeLimit is the number of bytes we want to keep the data store from exceeding.
	DBSizeLimit int
	// DBLowWater is the percentage of DBSizeLimit a GC run will reduce the used storage down
	// to.
	DBLowWater int
	// DBHighWater is the trigger point at which a GC run should start if exceeded.
	DBHighWater int
	// GCFrequency is the frequency of checks of the current utilisation.
	GCFrequency    time.Duration
	HasL2          bool
	BlockCacheSize int
	InitLogLevel   int
	Logger         *logger
	// DB is the badger db enveloper
	*badger.DB
	// seq is the monotonic collision free index for raw event storage.
	seq *badger.Sequence
	// Threads is how many CPU threads we dedicate to concurrent actions, flatten and GC mark
	Threads int
}

var _ eventstore.I = (*T)(nil)

// GetBackend returns a reasonably configured badger.Backend.
//
// The variadic params correspond to DBSizeLimit, DBLowWater, DBHighWater and
// GCFrequency as an integer multiplier of number of seconds.
//
// Note that the cancel function for the context needs to be managed by the
// caller.
func GetBackend(
	Ctx context.T,
	WG *sync.WaitGroup,
	hasL2 bool,
	blockCacheSize int,
	logLevel int,
	params ...int,
) (b *T) {
	var sizeLimit, lw, hw, freq = 0, 86, 92, 60
	switch len(params) {
	case 4:
		freq = params[3]
		fallthrough
	case 3:
		hw = params[2]
		fallthrough
	case 2:
		lw = params[1]
		fallthrough
	case 1:
		sizeLimit = params[0]
	}
	b = &T{
		Ctx:            Ctx,
		WG:             WG,
		DBSizeLimit:    sizeLimit,
		DBLowWater:     lw,
		DBHighWater:    hw,
		GCFrequency:    time.Duration(freq) * time.Second,
		HasL2:          hasL2,
		BlockCacheSize: blockCacheSize,
		InitLogLevel:   logLevel,
	}
	return
}

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (r *T) SerialKey() (idx []byte, ser *serial.T) {
	var err error
	var s []byte
	if s, err = r.SerialBytes(); chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return index.Event.Key(ser), ser
}

// Serial returns the next monotonic conflict free unique serial on the database.
func (r *T) Serial() (ser uint64, err error) {
	if ser, err = r.seq.Next(); chk.E(err) {
	}
	// log.T.F("serial %x", ser)
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *T) SerialBytes() (ser []byte, err error) {
	var serU64 uint64
	if serU64, err = r.Serial(); chk.E(err) {
		panic(err)
	}
	ser = make([]byte, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}
