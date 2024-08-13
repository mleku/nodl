package budger

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"git.replicatr.dev/pkg/util/context"
	"github.com/dgraph-io/badger/v4"
	"github.com/fiatjaf/eventstore"
)

func (b *BadgerBackend) SaveEvent(c context.T, ev *event.T) (err E) {
	return b.Update(func(txn *badger.Txn) error {
		var idx []byte
		var ser *serial.T
		idx, ser = b.SerialKey()
		// query event by id to ensure we don't save duplicates
		// id, _ := hex.DecodeString(ev.ID)
		// prefix := make([]byte, 1+8)
		// prefix[0] = indexIdPrefix
		// copy(prefix[1:], id)
		prefix := index.Id.Key(id.New(eventid.NewWith(ev.ID)))
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(prefix)
		if it.ValidForPrefix(prefix) {
			// event exists
			return eventstore.ErrDupEvent
		}

		// encode to binary
		var bin B
		bin, err = ev.MarshalBinary(bin)
		// bin, err := nostr_binary.Marshal(evt)
		if err != nil {
			return err
		}

		// idx, ser = b.SerialKey()
		// raw event store
		if err := txn.Set(idx, bin); err != nil {
			return err
		}

		for _, k := range GetIndexKeysForEvent(ev, ser) {
			if err = txn.Set(k, nil); err != nil {
				return err
			}
		}

		return nil
	})
}
