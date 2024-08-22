package relay

import (
	"sort"

	"nostr.mleku.dev/codec/envelopes/eoseenvelope"
	"nostr.mleku.dev/codec/envelopes/eventenvelope"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filters"
	"nostr.mleku.dev/codec/subscriptionid"
	"nostr.mleku.dev/protocol/relayws"
)

func (rl *T) handleReq(ws *relayws.WS, ff *filters.T, sub *subscriptionid.T) {
	var err E
	log.T.S(ff)
	if ff == nil {
		return
	}
	var evs []*event.T
	var events event.Ts
	for i, f := range ff.F {
		log.I.F("%d: %s", i, f)
		if evs, err = rl.Store.QueryEvents(rl.Ctx, f); chk.E(err) {
			continue
		}
		events=append(events, evs...)
	}
	sort.Sort(events)
	for _, ev := range events {
		if err = eventenvelope.NewResultWith(sub, ev).Write(ws); chk.E(err) {
		}
	}
	if err = eoseenvelope.NewFrom(sub).Write(ws); chk.E(err) {
		return
	}
	log.I.Ln("eose", ff)
}
