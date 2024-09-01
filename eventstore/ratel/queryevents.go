package ratel

import (
	"git.replicatr.dev/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filter"
)

func (r *T) QueryEvents(c Ctx, f *filter.T) (evs []*event.T, err E) {
	Log.I.F("query for events\n%s", f)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); Chk.E(err) {
		return
	}
	var limit bool
	if f.Limit != 0 {
		Log.W.S("query has a limit")
		limit = true
	}
	Log.I.S(queries, extraFilter)
	// search for the keys generated from the filter
	var eventKeys [][]byte
	for _, q := range queries {
		Log.I.S(q, extraFilter)
		err = r.View(func(txn *badger.Txn) (err E) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			// for it.Rewind(); it.Valid(); it.Next() {
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				Log.I.S(k)
				if !q.skipTS {
					if len(k) < createdat.Len+serial.Len {
						continue
					}
					createdAt := createdat.FromKey(k)
					Log.T.F("%d < %d", createdAt.Val.U64(), since)
					if createdAt.Val.U64() < since {
						break
					}
				}
				ser := serial.FromKey(k)
				eventKeys = append(eventKeys, index.Event.Key(ser))
			}
			return
		})
		if Chk.E(err) {
			// this can't actually happen because the View function above does not set err.
		}
		for _, eventKey := range eventKeys {
			// Log.I.S(eventKey)
			var v B
			err = r.View(func(txn *badger.Txn) (err E) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
				// for it.Rewind(); it.Valid(); it.Next() {
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					k := item.KeyCopy(nil)
					// if len(k) < len(q.searchPrefix) {
					// 	continue
					// }
					// if !bytes.HasPrefix(k, eventKey) {
					// 	continue
					// }
					Log.I.S(k)
					if v, err = item.ValueCopy(nil); Chk.E(err) {
						continue
					}
					// Log.W.S(v)
					if r.HasL2 && len(v) == sha256.Size {
						// this is a stub entry that indicates an L2 needs to be accessed for it, so
						// we populate only the event.T.ID and return the result, the caller will
						// expect this as a signal to query the L2 event store.
						ev := &event.T{}
						Log.T.F("found event stub %0x must seek in L2", v)
						ev.ID = v
						select {
						case <-c.Done():
							return
						case <-r.Ctx.Done():
							Log.I.Ln("backend context canceled")
							return
						default:
						}
						evs = append(evs, ev)
						return
					}
				}
				return
			})
			if v == nil {
				continue
			}
			ev := &event.T{}
			var rem B
			if rem, err = ev.UnmarshalBinary(v); Chk.E(err) {
				return
			}
			Log.I.S(ev)
			if len(rem) > 0 {
				Log.T.S(rem)
			}
			Log.W.Ln(extraFilter == nil)
			if extraFilter != nil {
				Log.W.Ln(extraFilter.Matches(ev))
			}
			// Log.I.S(ev)
			// check if this matches the other filters that were not part of the index
			if extraFilter == nil || extraFilter.Matches(ev) {
				Log.T.F("sending back result\n%s\n", ev)
				evs = append(evs, ev)
				if limit {
					f.Limit--
					if f.Limit == 0 {
						return
					}
				} else {
					// if there is no limit, cap it at the MaxLimit, assume this was the intent
					// or the client is erroneous, if any limit greater is requested this will
					// be used instead as the previous clause.
					if len(evs)>r.MaxLimit {
						return
					}
				}
			}
		}
	}
	Log.I.S(evs)
	Log.I.Ln("query complete")
	return
}
