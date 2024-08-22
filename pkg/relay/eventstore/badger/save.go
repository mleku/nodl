package badger

import (
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/eventid"
	"nostr.mleku.dev/codec/timestamp"
	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"util.mleku.dev/context"
	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
)

func (b *Backend) SaveEvent(c context.T, ev *event.T) (err error) {
	// make sure Close waits for this to complete
	b.WG.Add(1)
	defer b.WG.Done()
	// first, search to see if the event ID already exists.
	var foundSerial []byte
	seri := serial.New(nil)
	err = b.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := index.Id.Key(id.New(eventid.NewWith(ev.ID)))
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(prf)
		if it.ValidForPrefix(prf) {
			var k []byte
			// get the serial
			k = it.Item().Key()
			// copy serial out
			keys.Read(k, index.Empty(), id.New(eventid.New()), seri)
			// save into foundSerial
			foundSerial = seri.Val
		}
		return
	})
	if chk.E(err) {
		return
	}
	// if the event is found but it has been replaced with the event ID (in case of
	// L2) we need to restore it, or otherwise return that the event exists.
	if foundSerial != nil {
		err = b.Update(func(txn *badger.Txn) (err error) {
			// retrieve the event record
			evKey := keys.Write(index.New(index.Event), seri)
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			it.Seek(evKey)
			if it.ValidForPrefix(evKey) {
				if it.Item().ValueSize() != sha256.Size {
					// not a stub, we already have it
					// log.I.Ln(" duplicate event", ev.ID)
					return eventstore.ErrDupEvent
				}
				// we only need to restore the event binary and write the access counter key
				// encode to binary
				var bin B
				if bin, err = ev.MarshalBinary(bin); chk.E(err) {
					return
				}
				if err = txn.Set(it.Item().Key(), bin); chk.E(err) {
					return
				}
				// bump counter key
				counterKey := GetCounterKey(seri)
				val := keys.Write(createdat.New(timestamp.Now()))
				if err = txn.Set(counterKey, val); chk.E(err) {
					return
				}
				return
				// } else {
				// 	log.I.Ln("pruned duplicate event", ev.ID)
				// return eventstore.ErrDupEvent
			}
			return
		})
		// if it was a dupe, we are done.
		if err != nil {
			return
		}
		// if this is a restore, we are done, no need to cache the JSON, as it is not a
		// new event.
		return
	}
	log.I.F("saving event to badger %s", ev.Serialize())
	// otherwise, save new event record.
	if err = b.Update(func(txn *badger.Txn) (err error) {
		var idx []byte
		var ser *serial.T
		idx, ser = b.SerialKey()
		// encode to binary
		var bin B
		if bin, err = ev.MarshalBinary(bin); chk.E(err) {
			return
		}
		// raw event store
		if err = txn.Set(idx, bin); chk.E(err) {
			return
		}
		// 	add the indexes
		var keyz [][]byte
		keyz = GetIndexKeysForEvent(ev, ser)
		for _, k := range keyz {
			if err = txn.Set(k, nil); chk.E(err) {
				return
			}
		}
		// initialise access counter key
		counterKey := GetCounterKey(ser)
		val := keys.Write(createdat.New(timestamp.Now()))
		if err = txn.Set(counterKey, val); chk.E(err) {
			return
		}
		// log.T.F("event saved %s %s", ev.ID, b.Path)
		return
	}); chk.E(err) {
		return
	}
	return
}
