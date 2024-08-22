package badger

import (
	"errors"
	"sync/atomic"

	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/eventid"
	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"util.mleku.dev/context"
	"github.com/dgraph-io/badger/v4"
)

var deleteCounter atomic.Uint32

func (b *Backend) DeleteEvent(c context.T, ev *event.T) (err E) {
	deletionHappened := false

	err = b.Update(func(txn *badger.Txn) (err E) {
		idx := make([]byte, 1, 1+serial.Len)
		idKey := index.Id.Key(id.New(eventid.NewWith(ev.ID)))
		opts := badger.IteratorOptions{
			PrefetchValues: false,
		}
		it := txn.NewIterator(opts)
		it.Seek(idKey)
		var ser *serial.T
		if it.ValidForPrefix(idKey) {
			// we only need the serial to generate the event key
			ser = serial.New(nil)
			keys.Read(it.Item().Key(), index.Empty(), id.New(), ser)
			idx = index.Event.Key(ser)
			// log.D.Ln("added found item")
		}
		it.Close()
		// if no idx was found, end here, this event doesn't exist
		if len(idx) == 1 {
			return eventstore.ErrEventNotExists
		}
		// set this so we'll run the GC later
		deletionHappened = true
		// calculate all index keys we have for this event and delete them
		for _, k := range GetIndexKeysForEvent(ev, serial.New(idx[1:])) {
			if err = txn.Delete(k); chk.E(err) {
				return
			}
		}
		// delete the counter key
		if err = txn.Delete(GetCounterKey(ser)); chk.E(err) {
			return
		}
		// delete the raw event
		return txn.Delete(idx)
	})
	if chk.E(err) {
		return
	}
	// after deleting, run garbage collector (sometimes)
	if deletionHappened {
		if deleteCounter.Add(1)%256 == 0 {
			if err = b.RunValueLogGC(0.8); chk.E(err) &&
				!errors.Is(err, badger.ErrNoRewrite) {
				log.E.F("badger gc error:" + err.Error())
			}
		}
	}
	return nil
}
