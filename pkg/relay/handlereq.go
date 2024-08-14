package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/relay/eventstore"
)

func (rl *T) handleReq(ws *relayws.WS, env *reqenvelope.T) {
	var err E
	log.T.S(env)
	if env.Filters == nil {
		return
	}
	chans := make([]event.C, 0, len(env.Filters.F))
	for i, f := range env.Filters.F {
		log.I.F("%d: %s", i, f)
		var ch event.C
		if ch, err = rl.Store.QueryEvents(rl.Ctx, f); chk.E(err) {
			continue
		}
		chans = append(chans, ch)
	}
	// funnel the query channels into one.
	ch := eventstore.FanIn(ws.Ctx, chans...)
	go func(ch event.C) {
		for {
			select {
			case <-rl.Ctx.Done():
				return
			case ev := <-ch:
				// when channel is closed, we get a nil (could be erroneous send but should not
				// be).
				if ev == nil {
					return
				}
				// process incoming event matches
				log.I.S(ev)
			}
			return
		}
	}(ch)
	if err = ws.WriteEnvelope(eoseenvelope.NewFrom(env.Subscription)); chk.E(err) {
		return
	}
}
