package badger

import (
	"encoding/binary"
	"sync"
	"time"

	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/badger/del"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/units"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

var _ eventstore.I = (*Backend)(nil)

type PruneFunc func(ifc any, deleteItems del.Items) (err error)

type GCCountFunc func(ifc any) (deleteItems del.Items, err error)

type Backend struct {
	Ctx  context.T
	WG   *sync.WaitGroup
	Path string
	// MaxLimit is the largest a single event JSON can be, in bytes.
	MaxLimit int
	// DBSizeLimit is the number of bytes we want to keep the data store from
	// exceeding.
	DBSizeLimit int
	// DBLowWater is the percentage of DBSizeLimit a GC run will reduce the used
	// storage down to.
	DBLowWater int
	// DBHighWater is the trigger point at which a GC run should start if
	// exceeded.
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

func (b *Backend) Nuke() (err eventstore.E) {
	// TODO implement me
	panic("implement me")
}

const DefaultMaxLimit = 1024

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
	params ...int,
) (b *Backend) {
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
	b = &Backend{
		Ctx:            Ctx,
		WG:             WG,
		MaxLimit:       DefaultMaxLimit,
		DBSizeLimit:    sizeLimit,
		DBLowWater:     lw,
		DBHighWater:    hw,
		GCFrequency:    time.Duration(freq) * time.Second,
		HasL2:          hasL2,
		BlockCacheSize: blockCacheSize,
	}
	return
}

func (b *Backend) Init(path S) (err error) {
	b.Path = path
	log.I.Ln("opening badger event store at", b.Path)
	opts := badger.DefaultOptions(b.Path)
	opts.Compression = options.None
	opts.BlockCacheSize = int64(b.BlockCacheSize)
	opts.BlockSize = units.Mb
	opts.CompactL0OnClose = true
	opts.LmaxCompaction = true
	// opts.Compression = options.ZSTD
	b.Logger = NewLogger(b.InitLogLevel, b.Path)
	opts.Logger = b.Logger
	if b.DB, err = badger.Open(opts); chk.E(err) {
		return err
	}
	log.T.Ln("getting event store sequence index", b.Path)
	if b.seq, err = b.DB.GetSequence([]byte("events"), 1000); chk.E(err) {
		return err
	}
	log.T.Ln("running migrations", b.Path)
	if err = b.runMigrations(); chk.E(err) {
		return log.E.Err("error running migrations: %w; %s", err, b.Path)
	}
	if b.MaxLimit == 0 {
		b.MaxLimit = DefaultMaxLimit
	}
	if b.DBSizeLimit > 0 {
		go b.GarbageCollector()
	} else {
		go b.GCCount()
		// go b.IndexGCCount()
	}
	return nil
}

func (b *Backend) Close() (err error) {
	if err = b.DB.Flatten(4); chk.E(err) {
		return
	}
	if err = b.DB.Close(); chk.E(err) {
		return
	}
	if err = b.seq.Release(); chk.E(err) {
		return
	}
	return
}

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (b *Backend) SerialKey() (idx []byte, ser *serial.T) {
	var err error
	var s []byte
	if s, err = b.SerialBytes(); chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return index.Event.Key(ser), ser
}

func (b *Backend) Serial() (ser uint64, err error) {
	if ser, err = b.seq.Next(); chk.E(err) {
	}
	// log.T.F("serial %x", ser)
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (b *Backend) SerialBytes() (ser []byte, err error) {
	var serU64 uint64
	if serU64, err = b.Serial(); chk.E(err) {
		panic(err)
	}
	ser = make([]byte, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}

func (b *Backend) Update(fn func(txn *badger.Txn) (err error)) (err error) {
	err = b.DB.Update(fn)
	return
}

func (b *Backend) View(fn func(txn *badger.Txn) (err error)) (err error) {
	err = b.DB.View(fn)
	return
}
