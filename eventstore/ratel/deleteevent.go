package ratel

import (
	"git.replicatr.dev/eventstore/ratel/keys"
	"git.replicatr.dev/eventstore/ratel/keys/id"
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	"github.com/dgraph-io/badger/v4"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/eventid"
)

func (r *T) DeleteEvent(c Ctx, eid *eventid.T) (err E) {
	var foundSerial []byte
	seri := serial.New(nil)
	err = r.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := index.Id.Key(id.New(eid))
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
	if foundSerial == nil {
		return
	}
	var indexKeys []B
	ev := &event.T{}
	var evKey, evb, counterKey B
	// fetch the event to get its index keys
	err = r.View(func(txn *badger.Txn) (err error) {
		// retrieve the event record
		evKey = keys.Write(index.New(index.Event), seri)
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(evKey)
		if it.ValidForPrefix(evKey) {
			if evb, err = it.Item().ValueCopy(evb); Chk.E(err) {
				return
			}
			if _, err = ev.MarshalJSON(evb); Chk.E(err) {
				return
			}
			indexKeys = GetIndexKeysForEvent(ev, seri)
			counterKey = GetCounterKey(seri)
			return
		}
		return
	})
	if Chk.E(err) {
		return
	}
	err = r.Update(func(txn *badger.Txn) (err E) {
		if err = txn.Delete(evKey); Chk.E(err) {
		}
		for _, key := range indexKeys {
			if err = txn.Delete(key); Chk.E(err) {
			}
		}
		if err = txn.Delete(counterKey); Chk.E(err) {
			return
		}
		return
	})
	return
}
