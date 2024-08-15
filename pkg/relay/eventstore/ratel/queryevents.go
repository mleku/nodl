package ratel

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/ratel/keys/serial"
	"github.com/dgraph-io/badger/v4"
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
		err = r.View(func(txn *badger.Txn) (err E) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			// for it.Seek(q.start); it.Valid(); it.Next() {
			// 	item := it.Item()
			// 	key := item.Key()
			// 	if !equals(q.searchPrefix, key[:len(q.searchPrefix)]) {
			// 		continue
			// 	}
			// 	var val B
			// 	if val, err = item.ValueCopy(nil); chk.E(err) {
			// 		return
			// 	}
			// 	log.I.S(q.searchPrefix, key, val)
			// }
			// it.Rewind()
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				log.I.S(q.skipTS, k)
				if !q.skipTS {
					createdAt := createdat.FromKey(k)
					log.T.F("%d < %d", createdAt.Val.U64(), since)
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
	}
	log.T.S(eventKeys)
	_ = extraFilter
	return
}
