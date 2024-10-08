package ratel

import (
	"encoding/binary"
	"fmt"
	"math"

	"git.replicatr.dev/eventstore/ratel/keys/id"
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/kinder"
	"git.replicatr.dev/eventstore/ratel/keys/pubkey"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	. "nostr.mleku.dev"

	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/eventid"
	"nostr.mleku.dev/codec/filter"
	"nostr.mleku.dev/codec/timestamp"
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
	skipTS       bool
}

// PrepareQueries analyses a filter and generates a set of query specs that produce
// key prefixes to search for in the badger key indexes.
func PrepareQueries(f *filter.T) (
	qs []query,
	ext *filter.T,
	since uint64,
	err error,
) {
	if f == nil {
		err = Errorf.E("filter cannot be nil")
	}
	switch {
	// first if there is IDs, just search for them, this overrides all other filters
	case len(f.IDs.Field) > 0:
		qs = make([]query, f.IDs.Len())
		for i, idHex := range f.IDs.Field {
			ih := id.New(eventid.NewWith(B(idHex)))
			if ih == nil {
				Log.E.F("failed to decode event ID: %s", idHex)
				// just ignore it, clients will be clients
				continue
			}
			prf := index.Id.Key(ih)
			// Log.T.F("id prefix to search on %0x from key %0x", prf, ih.Val)
			qs[i] = query{
				index:        i,
				queryFilter:  f,
				searchPrefix: prf,
				skipTS:       true,
			}
		}
		// Log.T.S("ids", qs)
		// second we make a set of queries based on author pubkeys, optionally with kinds
	case f.Authors.Len() > 0:
		// if there is no kinds, we just make the queries based on the author pub keys
		if f.Kinds.Len() == 0 {
			qs = make([]query, f.Authors.Len())
			for i, pubkeyHex := range f.Authors.Field {
				var pk *pubkey.T
				if pk, err = pubkey.New(pubkeyHex); Chk.E(err) {
					// bogus filter, continue anyway
					continue
				}
				sp := index.Pubkey.Key(pk)
				// Log.I.F("search only for authors %0x from pub key %0x", sp, pk.Val)
				qs[i] = query{
					index:        i,
					queryFilter:  f,
					searchPrefix: sp,
				}
			}
			// Log.I.S("authors", qs)
		} else {
			// if there is kinds as well, we are searching via the kind/pubkey prefixes
			qs = make([]query, f.Authors.Len()*f.Kinds.Len())
			i := 0
		authors:
			for _, pubkeyHex := range f.Authors.Field {
				for _, kind := range f.Kinds.K {
					var pk *pubkey.T
					if pk, err = pubkey.New(pubkeyHex); Chk.E(err) {
						// skip this dodgy thing
						continue authors
					}
					ki := kinder.New(kind.K)
					sp := index.PubkeyKind.Key(pk, ki)
					// Log.T.F("search for authors from pub key %0x and kind %0x", pk.Val, ki.Val)
					qs[i] = query{index: i, queryFilter: f, searchPrefix: sp}
					i++
				}
			}
			// Log.T.S("authors/kinds", qs)
		}
		if f.Tags != nil && f.Tags.T != nil || f.Tags.Len() > 0 {
			ext = &filter.T{Tags: f.Tags}
			// Log.T.S("extra filter", ext)
		}
	case f.Tags.Len() > 0:
		// determine the size of the queries array by inspecting all tags sizes
		size := 0
		for _, values := range f.Tags.T {
			size += values.Len() - 1
		}
		if size == 0 {
			return nil, nil, 0, fmt.Errorf("empty tag filters")
		}
		// we need a query for each tag search
		qs = make([]query, size)
		// and any kinds mentioned as well in extra filter
		ext = &filter.T{Kinds: f.Kinds}
		i := 0
		Log.T.S(f.Tags.T)
		for _, values := range f.Tags.T {
			Log.T.S(values.Field)
			for _, value := range values.Field[1:] {
				// get key prefix (with full length) and offset where to write the last parts
				var prf []byte
				if prf, err = GetTagKeyPrefix(S(value)); Chk.E(err) {
					continue
				}
				// remove the last part to get just the prefix we want here
				Log.T.F("search for tags from %0x", prf)
				qs[i] = query{index: i, queryFilter: f, searchPrefix: prf}
				i++
			}
		}
		// Log.T.S("tags", qs)
	case f.Kinds.Len() > 0:
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
		// Log.T.S("kinds", qs)
	default:
		if len(qs) > 0 {
			qs[0] = query{index: 0, queryFilter: f,
				searchPrefix: index.CreatedAt.Key()}
			ext = nil
		}
		// Log.T.S("other", qs)
	}
	var until uint64 = math.MaxUint64
	if f.Until != nil {
		if fu := uint64(*f.Until); fu < until {
			until = fu - 1
		}
	}
	for i, q := range qs {
		qs[i].start = binary.BigEndian.AppendUint64(q.searchPrefix, until)
	}
	// this is where we'll end the iteration
	if f.Since != nil {
		if fs := uint64(*f.Since); fs > since {
			since = fs
		}
	}
	// if we got an empty filter, we still need a query for scraping the newest
	if len(qs) == 0 {
		qs = append(qs, query{index: 0, queryFilter: f, searchPrefix: B{1},
			start: B{1, 255, 255, 255, 255, 255, 255, 255, 255}})
	}
	return
}
