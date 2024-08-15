package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filters"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/protocol/relayws"
)

func (rl *T) handleReq(ws *relayws.WS, ff *filters.T, sub *subscriptionid.T) {
	var err E
	log.T.S(ff)
	if ff == nil {
		return
	}
	var evs []*event.T
	for i, f := range ff.F {
		log.I.F("%d: %s", i, f)
		if evs, err = rl.Store.QueryEvents(rl.Ctx, f); chk.E(err) {
			continue
		}
		for _, ev := range evs {
			if err = eventenvelope.NewResultWith(sub, ev).Write(ws); chk.E(err) {
			}
		}
	}
	if err = eoseenvelope.NewFrom(sub).Write(ws); chk.E(err) {
		return
	}
	log.I.Ln("eose", ff)
}
