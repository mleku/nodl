package ratel

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/serial"
	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
)

func (r *T) QueryEvents(c Ctx, f *filter.T) (evs []*event.T, err E) {
	log.I.F("query for events\n%s", f)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	// search for the keys generated from the filter
	var eventKeys [][]byte
	for _, q := range queries {
		// log.I.S(q)
		err = r.View(func(txn *badger.Txn) (err E) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				if !q.skipTS {
					createdAt := createdat.FromKey(k)
					log.T.F("%d > %d", createdAt.Val.U64(), since)
					if createdAt.Val.U64() < since {
						break
					}
				}
				ser := serial.FromKey(k)
				eventKeys = append(eventKeys, index.Event.Key(ser))
			}
			return
		})
		if chk.E(err) {
			// this can't actually happen because the View function above does not set err.
		}
		for _, eventKey := range eventKeys {
			var v B
			err = r.View(func(txn *badger.Txn) (err E) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					if v, err = item.ValueCopy(nil); chk.E(err) {
						continue
					}
					if r.HasL2 && len(v) == sha256.Size {
						// this is a stub entry that indicates an L2 needs to be accessed for it, so
						// we populate only the event.T.ID and return the result, the caller will
						// expect this as a signal to query the L2 event store.
						ev := &event.T{}
						log.T.F("found event stub %0x must seek in L2", v)
						ev.ID = v
						select {
						case <-c.Done():
							return
						case <-r.Ctx.Done():
							log.I.Ln("backend context canceled")
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
			if rem, err = ev.UnmarshalBinary(v); chk.E(err) {
				return
			}
			if len(rem) > 0 {
				log.T.S(rem)
			}
			// check if this matches the other filters that were not part of the index
			if extraFilter == nil || extraFilter.Matches(ev) {
				// todo: this is getting stuck here and causing a major goroutine leak
				log.T.F("sending back result %s", ev)
				evs = append(evs, ev)
			}
		}
	}
	log.I.Ln("query complete")
	return
}
