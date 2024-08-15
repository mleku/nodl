package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/protocol/relayws"
)

func (rl *T) handleReq(ws *relayws.WS, env *reqenvelope.T) {
	var err E
	log.T.S(env)
	if env.Filters == nil {
		return
	}
	var evs []*event.T
	for i, f := range env.Filters.F {
		log.I.F("%d: %s", i, f)
		if evs, err = rl.Store.QueryEvents(rl.Ctx, f); chk.E(err) {
			continue
		}
		for _, ev := range evs {
			if err = ws.WriteEnvelope(eventenvelope.NewResultWith(
				env.Subscription, ev)); chk.E(err) {
			}
		}
	}
	if err = ws.WriteEnvelope(eoseenvelope.NewFrom(env.Subscription)); chk.E(err) {
		return
	}
	// todo: register the subscription for matching on save event
	log.I.Ln("finished req", env.Filters)
}
