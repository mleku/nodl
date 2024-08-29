package relay

import (
	. "nostr.mleku.dev"
	"sort"

	"nostr.mleku.dev/codec/envelopes/eoseenvelope"
	"nostr.mleku.dev/codec/envelopes/eventenvelope"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filters"
	"nostr.mleku.dev/codec/subscriptionid"
	"nostr.mleku.dev/protocol/ws"
)

func (rl *T) handleReq(ws *ws.Serv, ff *filters.T, sub *subscriptionid.T) {
	var err E
	Log.T.S(ff)
	if ff == nil {
		return
	}
	var evs []*event.T
	var events event.Ts
	for i, f := range ff.F {
		Log.I.F("%d: %s", i, f)
		if evs, err = rl.Store.QueryEvents(rl.Ctx, f); Chk.E(err) {
			continue
		}
		events = append(events, evs...)
	}
	sort.Sort(events)
	for _, ev := range events {
		if err = eventenvelope.NewResultWith(sub, ev).Write(ws); Chk.E(err) {
		}
	}
	if err = eoseenvelope.NewFrom(sub).Write(ws); Chk.E(err) {
		return
	}
	Log.I.Ln("eose", ff)
}
