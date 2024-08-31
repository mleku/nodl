package relay

import (
	"fmt"
	"sync"
	"time"

	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/envelopes/closedenvelope"
	"nostr.mleku.dev/protocol/ws"

	"nostr.mleku.dev/codec/filter"
	"nostr.mleku.dev/codec/filters"
	"nostr.mleku.dev/codec/subscriptionid"
)

var (
	// DefaultMaxAge is a reasonable limit for subscriptions to stay live for, as humans do not
	// generally stay awake for more than 18 hours. A machine that wants to maintain a filter
	// for longer will just restart the subscription when it is timed out, or a custom relay
	// with other mechanisms for containing subscription resource use might set other values.
	DefaultMaxAge = time.Hour * 18
)

type WS = *ws.Serv
type filterFingerpint = uint64
type subStr = S
type subId = *subscriptionid.T

type (
	FilterMap   map[filterFingerpint]*filter.T
	FilterIndex map[filterFingerpint]time.Time
	SubMap      map[subStr]FilterIndex
	WsMap       map[WS]SubMap
)

type Tracker struct {
	sync.Mutex
	Ctx
	FilterMap
	WsMap
	// MaxAge is the amount of time that a subscription will be left open. Subscriptions should
	// have some reasonable lifespan as while they are live they consume resources storing the
	// filters and matching on them every time a new event arrives.
	MaxAge time.Duration
}

func (tr *Tracker) Do(fn func()) {
	Log.W.Ln("locking tracker")
	tr.Lock()
	fn()
	Log.W.Ln("unlocking tracker")
	tr.Unlock()
}

func (tr *Tracker) Init(c Ctx) {
	tr.Ctx = c
	tr.FilterMap = make(FilterMap)
	tr.WsMap = make(WsMap)
	if tr.MaxAge == 0 {
		tr.MaxAge = DefaultMaxAge
	}
	// Start up subscription GC.
	//
	// We do not disconnect sockets in the subscriptions GC, but rather track the age of
	// filters. When the filters exceed the configured MaxAge, the associated subscription is
	// closed and the filters are purged from the active FilterMap.
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for {
			select {
			case <-ticker.C:
				tr.Do(func() {
					// find expired filters and delete them from the WsMap, and close related
					// subscriptions.
					cancelSubs := make(map[subStr]struct{})
					for _, wm := range tr.WsMap {
						for sub, subMap := range wm {
							for fingerPrint, fpm := range subMap {
								if fpm.Sub(time.Now()) > tr.MaxAge {
									delete(subMap, fingerPrint)
									cancelSubs[sub] = struct{}{}
								}
							}
						}
					}
					// enumerate the remaining filter fingerprints
					fps := make(map[filterFingerpint]struct{})
					for w := range tr.WsMap {
						for s := range tr.WsMap[w] {
							for fp := range tr.WsMap[w][s] {
								fps[fp] = struct{}{}
							}
						}
					}
					// iterate the FilterMap and remove filters not found in the active filter
					// set.
					for fp := range tr.FilterMap {
						if _, ok := fps[fp]; !ok {
							delete(tr.FilterMap, fp)
						}
					}
					// send out CLOSED messages and delete subscriptions
					for w := range tr.WsMap {
						for s := range tr.WsMap[w] {
							if _, ok := cancelSubs[s]; ok {
								delete(tr.WsMap[w], s)
								sid, err := subscriptionid.New(s)
								if err != nil {
									continue
								}
								err = closedenvelope.NewFrom(sid, B(fmt.Sprintf(
									"closing subscription due to being older than %v", tr.MaxAge))).Write(w.Conn.NetConn())
								Chk.E(err)
							}
						}
					}
				})

			case <-tr.Ctx.Done():
				return
			}
		}
	}()
}

// IterateFilters takes a closure that iterates the WsMap, and then scans the FilterMap for the
// matching filter IDs to get the filter, and calls a closure which has the websocket,
// subscription ID and filter.
//
// With this in the SaveEvent method all pending filters can be matched on, their event envelope
// constructed with the correct subscriptionid.T and sent via the correct websocket to the
// client.
func (tr *Tracker) IterateFilters(iter func(ws WS, sub subId, f *filter.T)) {
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
}

// IterateByFilterId works like IterateFilters except in the opposite direction, based on a
// filter fingerprint. The provided closure will be given the websocket, subscription ID and
// full filter of every current filter that matches.
func (tr *Tracker) IterateByFilterId(fid filterFingerpint,
	iter func(ws WS, sub subId, f *filter.T)) {
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
}

func (tr *Tracker) AddWS(ws WS) {
	var ok bool
	if _, ok = tr.WsMap[ws]; !ok {
		Log.W.Ln("adding submap")
		// allocate a new SubMap for possible incoming subscriptions.
		tr.WsMap[ws] = SubMap{}
	}
}

func (tr *Tracker) RemoveWS(ws WS) {
	var sid subId
	var err E
	Log.T.F("removing websocket %s", ws.Remote())
	for sub := range tr.WsMap[ws] {
		if sid, err = subscriptionid.New(sub); Chk.E(err) {
			continue
		}
		tr.RemoveSub(ws, sid)
	}
}

func (tr *Tracker) AddSub(ws WS, sub subId, ff *filters.T) {
	tr.AddWS(ws)
	var err E
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
		if fp, err = f.Fingerprint(); Chk.E(err) {
			// filter is broken if it doesn't marshal. this actually can't happen if the JSON
			// unmarshalled.
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
		tr.WsMap[ws][s][fp] = time.Now()
	}
}

func (tr *Tracker) RemoveSub(ws *ws.Serv, sub *subscriptionid.T) {
	Log.T.F("removing subscription %s", sub.String())
	s := sub.String()
	if _, ok := tr.WsMap[ws]; ok {
		if _, ok = tr.WsMap[ws][s]; ok {
			// first decrement the FilterMap limit counter
			for fp := range tr.WsMap[ws][s] {
				var f *filter.T
				if f, ok = tr.FilterMap[fp]; ok {
					f.Limit--
					// if that puts it below zero, we can delete it.
					if f.Limit < 1 {
						Log.T.F("removing filter %s", f.Serialize())
						delete(tr.FilterMap, fp)
					}
				}
			}
			// with all references decremented or deleted we can now remove the subscription.
			delete(tr.WsMap[ws], sub.String())
		}

	}
}
