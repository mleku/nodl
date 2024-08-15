package ratel

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/serial"
	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
)

func (r *T) QueryEvents(c Ctx, f *filter.T) (ch event.C, err E) {
	log.I.F("query for events\n%s", f)
	ch = make(event.C, 1)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	log.T.S(queries, extraFilter, since)
	// search for the keys generated from the filter
	var eventKeys [][]byte
	for _, q := range queries {
		log.I.S(q)
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
				log.I.S(q.skipTS, k)
				if !q.skipTS {
					createdAt := createdat.FromKey(k)
					log.T.F("%d > %d", createdAt.Val.U64(), since)
					if createdAt.Val.U64() < since {
						break
					}
				}
				ser := serial.FromKey(k)
				log.I.S(ser)
				eventKeys = append(eventKeys, index.Event.Key(ser))
			}
			return
		})
		if chk.E(err) {
			// this can't actually happen because the View function above does not set err.
		}
		log.T.S(eventKeys)
		go func() {
			for {
				select {
				case res := <-q.results:
					if res.Ev != nil {
						log.I.S(res)
						ch <- res.Ev
					}
				case <-c.Done():
					return
				}
			}
		}()
		for _, eventKey := range eventKeys {
			var v B
			var ser *serial.T
			err = r.View(func(txn *badger.Txn) (err E) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					if v, err = item.ValueCopy(nil); chk.E(err) {
						continue
					}
					ser = serial.FromKey(item.KeyCopy(nil))
					if r.HasL2 && len(v) == sha256.Size {
						// this is a stub entry that indicates an L2 needs to be accessed for it, so
						// we populate only the event.T.ID and return the result, the caller will
						// expect this as a signal to query the L2 event store.
						evt := &event.T{}
						log.T.F("found event stub %0x must seek in L2", v)
						evt.ID = v
						select {
						case <-c.Done():
							return
						case <-r.Ctx.Done():
							log.I.Ln("backend context canceled")
							return
						default:
						}
						q.results <- Results{Ev: evt, TS: timestamp.Now(),
							Ser: ser}
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
				res := Results{Ev: ev, TS: timestamp.Now(), Ser: ser}
				// todo: this is getting stuck here and causing a major goroutine leak
				log.T.F("sending back result %s", ev)
				q.results <- res
			}
		}
	}
	log.I.Ln("query complete")
	// ch <- nil
	return
}
