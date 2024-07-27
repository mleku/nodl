package badger

import (
	"container/heap"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"git.replicatr.dev/pkg/relay/eventstore/badger/priority"
	"git.replicatr.dev/pkg/util/context"
	"github.com/dgraph-io/badger/v4"
	"github.com/minio/sha256-simd"
)

func (b *Backend) QueryEvents(c context.T, f *filter.T) (ch event.C,
	err error) {
	ch = make(event.C, 1)

	var queries []query
	var extraFilter *filter.T
	var since uint64
	queries, extraFilter, since, err = PrepareQueries(f)
	if chk.E(err) {
		return
	}
	accessChan := make(chan *AccessEvent)
	// start up the access counter
	go b.AccessLoop(c, accessChan)
	// max number of events we'll return
	limit := b.MaxLimit
	if f.Limit > 0 && f.Limit < limit {
		limit = f.Limit
	}
	go b.QueryEventsLoop(c, ch, accessChan, queries, limit, extraFilter, since)
	return ch, nil
}

func (b *Backend) QueryEventsLoop(c context.T, ch event.C,
	accessChan chan *AccessEvent,
	queries []query, limit int, extraFilter *filter.T, since uint64) {

	var err error
	defer func() {
		close(ch)
		close(accessChan)
	}()
	// actually iterate
	for _, q1 := range queries {
		select {
		case <-c.Done():
			// log.T.Ln("websocket closed")
			return
		case <-b.Ctx.Done():
			// log.I.Ln("backend context canceled")
			return
		default:
		}
		q2 := q1
		go b.QueryEventsSearch(c, q2, since, extraFilter)
	}
	// receive results and ensure we only return the most recent ones always
	emittedEvents := 0
	// first pass
	emitQueue := make(priority.Queue, 0, len(queries)+limit)
	for _, q := range queries {
		q := q
		evt, ok := <-q.results
		if ok {
			emitQueue = append(emitQueue,
				&priority.QueryEvent{
					T:     evt.Ev,
					Query: q.index,
					Ser:   evt.Ser,
				})
		}
	}
	// queue may be empty here if we have literally nothing
	if len(emitQueue) == 0 {
		return
	}
	heap.Init(&emitQueue)
	// iterate until we've emitted all events required
	for {
		select {
		case <-c.Done():
			// websocket closed
			log.T.Ln("websocket closed")
			return
		case <-b.Ctx.Done():
			// backend context canceled
			log.T.Ln("backend context canceled")
			return
		default:
		}
		// emit latest event in queue
		latest := emitQueue[0]
		ch <- latest.T
		// send ID to be incremented for access
		ae := MakeAccessEvent(eventid.NewWith(latest.ID), latest.Ser)
		accessChan <- ae
		// stop when reaching limit
		emittedEvents++
		if emittedEvents == limit {
			break
		}
		// fetch a new one from query results and replace the previous one with it
		if evt, ok := <-queries[latest.Query].results; ok {
			emitQueue[0].T = evt.Ev
			emitQueue[0].Ser = evt.Ser
			heap.Fix(&emitQueue, 0)
		} else {
			// if this query has no more events we just remove this and proceed normally
			heap.Remove(&emitQueue, 0)
			// check if the list is empty and end
			if len(emitQueue) == 0 {
				break
			}
		}
	}
	if chk.E(err) {
		log.D.F("badger: query txn error: %s", err)
	}
	return
}

func (b *Backend) QueryEventsSearch(c context.T, q2 query, since uint64,
	extraFilter *filter.T) {
	var eventKeys [][]byte
	err := b.View(func(txn *badger.Txn) (err error) {
		// iterate only through keys and in reverse order
		opts := badger.IteratorOptions{Reverse: true}
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(q2.start); it.ValidForPrefix(q2.searchPrefix); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			if !q2.skipTS {
				createdAt := createdat.FromKey(k)
				if createdAt.Val.U64() < since {
					break
				}
			}
			ser := serial.FromKey(k)
			eventKeys = append(eventKeys, index.Event.Key(ser))
		}
		return
	})
	if err != nil {
		close(q2.results)
		for _ = range q2.results {
		}
		return
	}
	for _, eventKey := range eventKeys {
		var ev *event.T
		err = b.View(func(txn *badger.Txn) (err error) {
			opts := badger.IteratorOptions{Reverse: true}
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
				item := it.Item()
				var v []byte
				if v, err = item.ValueCopy(nil); chk.E(err) {
					continue
				}
				ser := serial.FromKey(item.KeyCopy(nil))
				if len(v) == sha256.Size {
					// this is a stub entry that indicates an L2 needs to be accessed for it, so we
					// populate only the event.T.ID and return the result.
					evt := &event.T{}
					log.T.F("found event stub %0x must seek in L2", v)
					evt.ID = v
					select {
					case <-c.Done():
						return
					case <-b.Ctx.Done():
						log.I.Ln("backend context canceled")
						return
					default:
					}
					q2.results <- Results{Ev: evt, TS: timestamp.Now(),
						Ser: ser}
					return
				}
				var rem B
				if rem, err = ev.UnmarshalBinary(v); chk.E(err) {
					return
				}
				if len(rem) > 0 {
					log.I.S(rem)
				}
				// check if this matches the other filters that were not part of the index
				if extraFilter == nil || extraFilter.Matches(ev) {
					res := Results{Ev: ev, TS: timestamp.Now(), Ser: ser}
					// todo: this is getting stuck here and causing a major goroutine leak
					q2.results <- res
				}
			}
			return
		})
	}
	close(q2.results)
	for _ = range q2.results {
	}
	return
}
