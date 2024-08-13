package budger

import (
	"container/heap"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/arb"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/kinder"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/pubkey"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"git.replicatr.dev/pkg/util/hex"
	"github.com/dgraph-io/badger/v4"
)

type Results struct {
	Ev  *event.T
	TS  *timestamp.T
	Ser *serial.T
}

type query struct {
	index        int
	queryFilter  *filter.T
	searchPrefix []byte
	start        []byte
	results      chan Results
	skipTS       bool
}

// type query struct {
// 	i             int
// 	prefix        []byte
// 	startingPoint []byte
// 	results       chan *nostr.Event
// 	skipTimestamp bool
// }

type queryEvent struct {
	*event.T
	Ser   *serial.T
	query int
}

var exit = errors.New("exit")

func (b BadgerBackend) QueryEvents(ctx context.Context, f *filter.T) (event.C, E) {
	ch := make(event.C)

	if len(f.Search) != 0 {
		close(ch)
		return ch, nil
	}

	queries, extraFilter, since, err := prepareQueries(f)
	if err != nil {
		return nil, err
	}

	// max number of events we'll return
	limit := b.MaxLimit / 4
	if f.Limit > 0 && f.Limit < b.MaxLimit {
		limit = f.Limit
	}

	go func() {
		defer close(ch)

		// actually iterate
		for _, q := range queries {
			q := q

			pulled := 0 // this query will be hardcapped at this global limit

			go b.View(func(txn *badger.Txn) error {
				// iterate only through keys and in reverse order
				opts := badger.IteratorOptions{
					Reverse: true,
				}

				it := txn.NewIterator(opts)
				defer it.Close()
				defer close(q.results)

				for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
					item := it.Item()
					key := item.Key()
					ser := serial.FromKey(key)

					idxOffset := len(key) - 4 // this is where the idx actually starts

					// "id" indexes don't contain a timestamp
					if !q.skipTS {
						createdAt := createdat.FromKey(key)
						if createdAt.Val.U64() < since {
							break
						}
						// createdAt := binary.BigEndian.Uint64(key[idxOffset-4 : idxOffset])
						// 					if createdAt < since {
						// 						break
						// 					}
					}

					idx := make([]byte, 5)
					idx[0] = rawEventStorePrefix
					copy(idx[1:], key[idxOffset:])

					// fetch actual event
					item, err := txn.Get(idx)
					if err != nil {
						if err == badger.ErrDiscardedTxn {
							return err
						}
						log.E.F("badger: failed to get %x based on prefix %x, index key %x from raw event store: %s\n",
							idx, q.searchPrefix, key, err)
						return err
					}

					if err := item.Value(func(val []byte) error {
						evt := &event.T{}
						var rem B
						if rem, err = evt.UnmarshalBinary(val); chk.E(err) {
							return err
						}
						_ = rem
						// if err := nostr_binary.Unmarshal(val, evt); err != nil {
						// 	log.E.F("badger: value read error (id %x): %s\n", val[0:32], err)
						// 	return err
						// }

						// check if this matches the other filters that were not part of the index
						if extraFilter == nil || extraFilter.Matches(evt) {
							res := Results{Ev: evt, TS: timestamp.Now(), Ser: ser}
							select {
							case q.results <- res:
								pulled++
								if pulled > limit {
									return exit
								}
							case <-ctx.Done():
								return exit
							}
						}

						return nil
					}); err == exit {
						return nil
					} else if err != nil {
						return err
					}
				}

				return nil
			})
		}

		// receive results and ensure we only return the most recent ones always
		emittedEvents := 0

		// first pass
		emitQueue := make(priorityQueue, 0, len(queries))
		for _, q := range queries {
			evt, ok := <-q.results
			if ok {
				emitQueue = append(emitQueue, &queryEvent{T: evt.Ev, query: q.index})
			}
		}

		// queue may be empty here if we have literally nothing
		if len(emitQueue) == 0 {
			return
		}

		heap.Init(&emitQueue)

		// iterate until we've emitted all events required
		for {
			// emit latest event in queue
			latest := emitQueue[0]
			ch <- latest.T

			// stop when reaching limit
			emittedEvents++
			if emittedEvents == limit {
				break
			}

			// fetch a new one from query results and replace the previous one with it
			if evt, ok := <-queries[latest.query].results; ok {
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
	}()

	return ch, nil
}

type priorityQueue []*queryEvent

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].CreatedAt.I64() > pq[j].CreatedAt.I64()
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x any) {
	item := x.(*queryEvent)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

func prepareQueries(f *filter.T) (
	qs []query,
	extraFilter *filter.T,
	since uint64,
	err error,
) {

	since = uint64(math.MaxInt64)
	switch {
	// first if there is IDs, just search for them, this overrides all other filters
	case f.IDs != nil && len(f.IDs.Field) > 0:
		qs = make([]query, f.IDs.Len())
		for i, idHex := range f.IDs.Field {
			ih := id.New(eventid.NewWith(B(idHex)))
			if ih == nil {
				log.E.F("failed to decode event ID: %s", idHex)
				// just ignore it, clients will be clients
				continue
			}
			prf := index.Id.Key(ih)
			// log.T.F("id prefix to search on %0x from key %0x", prf, ih.Val)
			qs[i] = query{
				index:        i,
				queryFilter:  f,
				searchPrefix: prf,
				skipTS:       true, // why are we not checking timestamps? (ID has no timestamp)
			}
		}
		// log.T.S("ids", qs)
		// second we make a set of queries based on author pubkeys, optionally with kinds
	case f.Authors != nil && f.Authors.Len() > 0:
		// if there is no kinds, we just make the queries based on the author pub keys
		if f.Kinds.Len() == 0 {
			qs = make([]query, f.Authors.Len())
			for i, pubkeyHex := range f.Authors.Field {
				var pk *pubkey.T
				if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
					// bogus filter, continue anyway
					continue
				}
				sp := index.Pubkey.Key(pk)
				// log.I.F("search only for authors %0x from pub key %0x", sp, pk.Val)
				qs[i] = query{
					index:        i,
					queryFilter:  f,
					searchPrefix: sp,
				}
			}
			// log.I.S("authors", qs)
		} else {
			// if there is kinds as well, we are searching via the kind/pubkey prefixes
			qs = make([]query, f.Authors.Len()*f.Kinds.Len())
			i := 0
		authors:
			for _, pubkeyHex := range f.Authors.Field {
				for _, kind := range f.Kinds.K {
					var pk *pubkey.T
					if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
						// skip this dodgy thing
						continue authors
					}
					ki := kinder.New(kind.K)
					sp := index.PubkeyKind.Key(pk, ki)
					// log.T.F("search for authors from pub key %0x and kind %0x", pk.Val, ki.Val)
					qs[i] = query{index: i, queryFilter: f, searchPrefix: sp}
					i++
				}
			}
			// log.T.S("authors/kinds", qs)
		}
		if f.Tags != nil && f.Tags.Len() > 0 {
			extraFilter = &filter.T{Tags: f.Tags}
			// log.T.S("extra filter", ext)
		}
	case f.Tags != nil && f.Tags.Len() > 0:
		// determine the size of the queries array by inspecting all tags sizes
		size := 0
		for _, values := range f.Tags.T {
			size += values.Len()
		}
		if size == 0 {
			return nil, nil, 0, fmt.Errorf("empty tag filters")
		}
		// we need a query for each tag search
		qs = make([]query, size)
		// and any kinds mentioned as well in extra filter
		extraFilter = &filter.T{Kinds: f.Kinds}
		i := 0
		for _, values := range f.Tags.T {
			for _, value := range values.Field {
				// get key prefix (with full length) and offset where to write the last parts
				var prf []byte
				if prf, err = GetTagKeyPrefix(S(value)); chk.E(err) {
					continue
				}
				// remove the last part to get just the prefix we want here
				// log.T.F("search for tags from %0x", prf)
				qs[i] = query{index: i, queryFilter: f, searchPrefix: prf}
				i++
			}
		}
		// log.T.S("tags", qs)
	case f.Kinds != nil && f.Kinds.Len() > 0:
		// if there is no ids, pubs or tags, we are just searching for kinds
		qs = make([]query, f.Kinds.Len())
		for i, kind := range f.Kinds.K {
			kk := kinder.New(kind.K)
			ki := index.Kind.Key(kk)
			qs[i] = query{
				index:        i,
				queryFilter:  f,
				searchPrefix: ki,
			}
		}
		// log.T.S("kinds", qs)
	default:
		if len(qs) > 0 {
			qs[0] = query{index: 0, queryFilter: f,
				searchPrefix: index.CreatedAt.Key()}
			extraFilter = nil
		}
		// log.T.S("other", qs)
	}
	var until uint64 = math.MaxUint64
	if f.Until != nil {
		if fu := uint64(*f.Until); fu < until {
			until = fu + 1
		}
	}
	for i, q := range qs {
		qs[i].start = binary.BigEndian.AppendUint64(q.searchPrefix, until)
		qs[i].results = make(chan Results, 128)
	}
	// this is where we'll end the iteration
	if f.Since != nil {
		if fs := uint64(*f.Since); fs > since {
			since = fs
		}
	}
	return

	// var index byte
	//
	// if len(f.IDs) > 0 {
	// 	index = indexIdPrefix
	// 	queries = make([]query, len(f.IDs))
	// 	for i, idHex := range f.IDs {
	// 		prefix := make([]byte, 1+8)
	// 		prefix[0] = index
	// 		if len(idHex) != 64 {
	// 			return nil, nil, 0, fmt.Errorf("invalid id '%s'", idHex)
	// 		}
	// 		idPrefix8, _ := hex.DecodeString(idHex[0 : 8*2])
	// 		copy(prefix[1:], idPrefix8)
	// 		queries[i] = query{i: i, prefix: prefix, skipTimestamp: true}
	// 	}
	// } else if len(f.Authors) > 0 {
	// 	if len(f.Kinds) == 0 {
	// 		index = indexPubkeyPrefix
	// 		queries = make([]query, len(f.Authors))
	// 		for i, pubkeyHex := range f.Authors {
	// 			if len(pubkeyHex) != 64 {
	// 				return nil, nil, 0, fmt.Errorf("invalid pubkey '%s'", pubkeyHex)
	// 			}
	// 			pubkeyPrefix8, _ := hex.DecodeString(pubkeyHex[0 : 8*2])
	// 			prefix := make([]byte, 1+8)
	// 			prefix[0] = index
	// 			copy(prefix[1:], pubkeyPrefix8)
	// 			queries[i] = query{i: i, prefix: prefix}
	// 		}
	// 	} else {
	// 		index = indexPubkeyKindPrefix
	// 		queries = make([]query, len(f.Authors)*len(f.Kinds))
	// 		i := 0
	// 		for _, pubkeyHex := range f.Authors {
	// 			for _, kind := range f.Kinds {
	// 				if len(pubkeyHex) != 64 {
	// 					return nil, nil, 0, fmt.Errorf("invalid pubkey '%s'", pubkeyHex)
	// 				}
	// 				pubkeyPrefix8, _ := hex.DecodeString(pubkeyHex[0 : 8*2])
	// 				prefix := make([]byte, 1+8+2)
	// 				prefix[0] = index
	// 				copy(prefix[1:], pubkeyPrefix8)
	// 				binary.BigEndian.PutUint16(prefix[1+8:], uint16(kind))
	// 				queries[i] = query{i: i, prefix: prefix}
	// 				i++
	// 			}
	// 		}
	// 	}
	// 	extraFilter = &nostr.Filter{Tags: f.Tags}
	// } else if len(f.Tags) > 0 {
	// 	// determine the size of the queries array by inspecting all tags sizes
	// 	size := 0
	// 	for _, values := range f.Tags {
	// 		size += len(values)
	// 	}
	// 	if size == 0 {
	// 		return nil, nil, 0, fmt.Errorf("empty tag filters")
	// 	}
	//
	// 	queries = make([]query, size)
	//
	// 	extraFilter = &nostr.Filter{Kinds: f.Kinds}
	// 	i := 0
	// 	for _, values := range f.Tags {
	// 		for _, value := range values {
	// 			// get key prefix (with full length) and offset where to write the last parts
	// 			k, offset := getTagIndexPrefix(value)
	// 			// remove the last parts part to get just the prefix we want here
	// 			prefix := k[0:offset]
	//
	// 			queries[i] = query{i: i, prefix: prefix}
	// 			i++
	// 		}
	// 	}
	// } else if len(f.Kinds) > 0 {
	// 	index = indexKindPrefix
	// 	queries = make([]query, len(f.Kinds))
	// 	for i, kind := range f.Kinds {
	// 		prefix := make([]byte, 1+2)
	// 		prefix[0] = index
	// 		binary.BigEndian.PutUint16(prefix[1:], uint16(kind))
	// 		queries[i] = query{i: i, prefix: prefix}
	// 	}
	// } else {
	// 	index = indexCreatedAtPrefix
	// 	queries = make([]query, 1)
	// 	prefix := make([]byte, 1)
	// 	prefix[0] = index
	// 	queries[0] = query{i: 0, prefix: prefix}
	// 	extraFilter = nil
	// }
	//
	// var until uint32 = 4294967295
	// if f.Until != nil {
	// 	if fu := uint32(*f.Until); fu < until {
	// 		until = fu + 1
	// 	}
	// }
	// for i, q := range queries {
	// 	queries[i].startingPoint = binary.BigEndian.AppendUint32(q.prefix, uint32(until))
	// 	queries[i].results = make(chan *nostr.Event, 12)
	// }
	//
	// // this is where we'll end the iteration
	// if f.Since != nil {
	// 	if fs := uint32(*f.Since); fs > since {
	// 		since = fs
	// 	}
	// }
	//
	// return queries, extraFilter, since, nil
}

// GetTagKeyPrefix returns tag index prefixes based on the initial field of a
// tag.
//
// There is 3 types of index tag keys:
//
// - TagAddr:   [ 8 ][ 2b Kind ][ 8b Pubkey ][ address/URL ][ 8b Serial ]
// - Tag32:     [ 7 ][ 8b Pubkey ][ 8b Serial ]
// - Tag:       [ 6 ][ address/URL ][ 8b Serial ]
//
// This function produces the initial bytes without the index.
func GetTagKeyPrefix(tagValue string) (key []byte, err error) {
	if k, pkb, d := GetAddrTagElements(tagValue); len(pkb) == 32 {
		// store value in the new special "a" tag index
		var pk *pubkey.T
		if pk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		els := []keys.Element{kinder.New(k), pk}
		if len(d) > 0 {
			els = append(els, arb.NewFromString(d))
		}
		key = index.TagAddr.Key(els...)
	} else if pkb, _ := hex.Dec(tagValue); len(pkb) == 32 {
		// store value as bytes
		var pkk *pubkey.T
		if pkk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		key = index.Tag32.Key(pkk)
	} else {
		// store whatever as utf-8
		if len(tagValue) > 0 {
			var a *arb.T
			a = arb.NewFromString(tagValue)
			key = index.Tag.Key(a)
		}
		key = index.Tag.Key()
	}
	return
}

func GetAddrTagElements(tagValue S) (k uint16, pkb B, d S) {
	split := strings.Split(tagValue, ":")
	if len(split) == 3 {
		if pkb, _ = hex.Dec(split[1]); len(pkb) == 32 {
			if key, err := strconv.ParseUint(split[0], 10, 16); err == nil {
				return uint16(key), pkb, split[2]
			}
		}
	}
	return 0, nil, ""
}
