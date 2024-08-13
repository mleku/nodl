package budger

import (
	"encoding/binary"
	"fmt"

	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"github.com/dgraph-io/badger/v4"
)

const (
	dbVersionKey          byte = 255
	rawEventStorePrefix   byte = 0
	indexCreatedAtPrefix  byte = 1
	indexIdPrefix         byte = 2
	indexKindPrefix       byte = 3
	indexPubkeyPrefix     byte = 4
	indexPubkeyKindPrefix byte = 5
	indexTagPrefix        byte = 6
	indexTag32Prefix      byte = 7
	indexTagAddrPrefix    byte = 8
)

var _ eventstore.Store = (*BadgerBackend)(nil)

type BadgerBackend struct {
	Path     string
	MaxLimit int

	*badger.DB
	seq *badger.Sequence
}

func (b *BadgerBackend) Init() error {
	db, err := badger.Open(badger.DefaultOptions(b.Path))
	if err != nil {
		return err
	}
	b.DB = db
	b.seq, err = db.GetSequence([]byte("events"), 1000)
	if err != nil {
		return err
	}

	if err := b.runMigrations(); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	if b.MaxLimit == 0 {
		b.MaxLimit = 500
	}

	return nil
}

func (b BadgerBackend) Close() {
	b.DB.Close()
	b.seq.Release()
}

// func (b BadgerBackend) Serial() []byte {
// 	v, _ := b.seq.Next()
// 	vb := make([]byte, 5)
// 	vb[0] = rawEventStorePrefix
// 	binary.BigEndian.PutUint32(vb[1:], uint32(v))
// 	return vb
// }

func (b *BadgerBackend) Serial() (ser uint64, err error) {
	if ser, err = b.seq.Next(); chk.E(err) {
	}
	// log.T.F("serial %x", ser)
	return
}

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (b *BadgerBackend) SerialKey() (idx []byte, ser *serial.T) {
	var err error
	var s []byte
	if s, err = b.SerialBytes(); chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return index.Event.Key(ser), ser
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (b *BadgerBackend) SerialBytes() (ser []byte, err error) {
	var serU64 uint64
	if serU64, err = b.Serial(); chk.E(err) {
		panic(err)
	}
	ser = make([]byte, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}
