package ratel

import (
	"fmt"

	"git.replicatr.dev/eventstore"
	"git.replicatr.dev/eventstore/ratel/keys"
	"git.replicatr.dev/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/eventstore/ratel/keys/id"
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	. "nostr.mleku.dev"

	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/eventid"
	"nostr.mleku.dev/codec/timestamp"
)

func (r *T) SaveEvent(c Ctx, ev *event.T) (err E) {
	if ev.Kind.IsEphemeral() {
		Log.I.F("not saving ephemeral event\n%s", ev.Serialize())
		return
	}
	Log.T.C(func() S {
		evs, _ := ev.MarshalJSON(nil)
		return fmt.Sprintf("saving event\n%d %s", len(evs), evs)
	})
	// make sure Close waits for this to complete
	r.WG.Add(1)
	defer r.WG.Done()
	// first, search to see if the event ID already exists.
	var foundSerial []byte
	seri := serial.New(nil)
	err = r.View(func(txn *badger.Txn) (err error) {
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
	if Chk.E(err) {
		return
	}
	if foundSerial != nil {
		Log.T.Ln("found possible duplicate or stub for %s", ev)
		err = r.Update(func(txn *badger.Txn) (err error) {
			// retrieve the event record
			evKey := keys.Write(index.New(index.Event), seri)
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			it.Seek(evKey)
			if it.ValidForPrefix(evKey) {
				if it.Item().ValueSize() != sha256.Size {
					// not a stub, we already have it
					Log.T.Ln("duplicate event", ev.ID)
					return eventstore.ErrDupEvent
				}
				// we only need to restore the event binary and write the access counter key
				// encode to binary
				var bin B
				if bin, err = ev.MarshalBinary(bin); Chk.E(err) {
					return
				}
				if err = txn.Set(it.Item().Key(), bin); Chk.E(err) {
					return
				}
				// bump counter key
				counterKey := GetCounterKey(seri)
				val := keys.Write(createdat.New(timestamp.Now()))
				if err = txn.Set(counterKey, val); Chk.E(err) {
					return
				}
				return
			}
			return
		})
		// if it was a dupe, we are done.
		if err != nil {
			return
		}
		return
	}
	var bin B
	if bin, err = ev.MarshalBinary(bin); Chk.E(err) {
		return
	}
	Log.I.F("saving event to badger %s", ev)
	// otherwise, save new event record.
	if err = r.Update(func(txn *badger.Txn) (err error) {
		var idx []byte
		var ser *serial.T
		idx, ser = r.SerialKey()
		// encode to binary
		// raw event store
		if err = txn.Set(idx, bin); Chk.E(err) {
			return
		}
		// 	add the indexes
		var indexKeys [][]byte
		indexKeys = GetIndexKeysForEvent(ev, ser)
		for _, k := range indexKeys {
			if err = txn.Set(k, nil); Chk.E(err) {
				return
			}
		}
		// initialise access counter key
		counterKey := GetCounterKey(ser)
		val := keys.Write(createdat.New(timestamp.Now()))
		if err = txn.Set(counterKey, val); Chk.E(err) {
			return
		}
		Log.T.F("event saved %0x %s", ev.ID, r.Path)
		return
	}); Chk.E(err) {
		return
	}
	return
}
