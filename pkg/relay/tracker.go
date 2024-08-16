package relay

import (
	"sync"

	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/filters"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/protocol/relayws"
)

type WS = *relayws.WS
type filterId = uint64
type empty = struct{}
type subStr = S
type subId = *subscriptionid.T

type (
	FilterMap   map[filterId]*filter.T
	FilterIndex map[filterId]empty
	SubMap      map[subStr]FilterIndex
	WsMap       map[WS]SubMap
)

type Tracker struct {
	sync.Mutex
	FilterMap
	WsMap
}

func (tr *Tracker) Init() {
	tr.Lock()
	tr.FilterMap = make(FilterMap)
	tr.WsMap = make(WsMap)
	tr.Unlock()
}

// IterateFilters takes a closure that iterates the WsMap, and then scans the FilterMap for the
// matching filter IDs to get the filter, and calls a closure which has the websocket,
// subscription ID and filter.
//
// With this in the SaveEvent method all pending filters can be matched on, their event envelope
// constructed with the correct subscriptionid.T and sent via the correct websocket to the
// client.
func (tr *Tracker) IterateFilters(iter func(ws WS, sub subId, f *filter.T)) {
	tr.Lock()
	for w, sock := range tr.WsMap {
		for s, sub := range sock {
			for idx := range sub {
				if i, ok := tr.FilterMap[idx]; ok {
					sid, err := subscriptionid.New(s)
					if err != nil {
						continue
					}
					iter(w, sid, i)
				}
			}
		}
	}
	tr.Unlock()
}

// IterateByFilterId works like IterateFilters except in the opposite direction, based on a
// filter fingerprint. The provided closure will be given the websocket, subscription ID and
// full filter of every current filter that matches.
func (tr *Tracker) IterateByFilterId(fid filterId, iter func(ws WS, sub subId, f *filter.T)) {
	tr.Lock()
	for w, sock := range tr.WsMap {
		for s, sub := range sock {
			for idx := range sub {
				for i, f := range tr.FilterMap {
					if i == fid && i == idx {
						sid, err := subscriptionid.New(s)
						if err != nil {
							continue
						}
						iter(w, sid, f)
					}
				}
			}
		}
	}
	tr.Unlock()
}

func (tr *Tracker) AddWS(ws WS) {
	tr.Lock()
	var ok bool
	if _, ok = tr.WsMap[ws]; !ok {
		// allocate a new SubMap for possible incoming subscriptions.
		tr.WsMap[ws] = SubMap{}
	}
	tr.Unlock()
}

func (tr *Tracker) RemoveWS(ws WS) {
	tr.Lock()
	var sid subId
	var err E
	for sub := range tr.WsMap[ws] {
		if sid, err = subscriptionid.New(sub); chk.E(err) {
			continue
		}
		tr.RemoveSub(ws, sid)
	}
	tr.Unlock()
}

func (tr *Tracker) AddSub(ws WS, sub subId, ff *filters.T) {
	tr.AddWS(ws)
	var err E
	tr.Lock()
	var ok bool
	s := sub.String()
	if _, ok = tr.WsMap[ws][s]; !ok {
		// if the subscription doesn't exist, allocate it.
		tr.WsMap[ws][s] = make(FilterIndex)
	}
	// generate the filter fingerprints, store them in the filter index, and add the fingerprint
	// to the subscription map.
	for _, f := range ff.F {
		var fp uint64
		if fp, err = f.Fingerprint(); chk.E(err) {
			// filter is broken if it doesn't marshal. this actually can't happen if the
			// JSON unmarshalled.
			continue
		}
		// add the filter to the FilterMap if it doesn't exist.
		if _, ok = tr.FilterMap[fp]; !ok {
			tr.FilterMap[fp] = f.Clone() // this sets Limit to 1 as a reference count.
		} else {
			// increase the reference count.
			tr.FilterMap[fp].Limit++
		}
		// add the filter fingerprint to the subscription.
		tr.WsMap[ws][s][fp] = struct{}{}
	}
	tr.Unlock()
}

func (tr *Tracker) RemoveSub(ws *relayws.WS, sub *subscriptionid.T) {
	tr.Lock()
	s := sub.String()
	if _, ok := tr.WsMap[ws][s]; ok {
		// first decrement the FilterMap limit counter
		for fp := range tr.WsMap[ws][s] {
			var f *filter.T
			if f, ok = tr.FilterMap[fp]; ok {
				f.Limit--
				// if that puts it below zero, we can delete it.
				if f.Limit < 1 {
					delete(tr.FilterMap, fp)
				}
			}
		}
		// with all references decremented or deleted we can now remove the subscription.
		delete(tr.WsMap[ws], sub.String())
	}
	tr.Unlock()
}
