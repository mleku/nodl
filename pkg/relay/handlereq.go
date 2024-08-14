package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/protocol/relayws"
)

func (rl *T) handleReq(ws *relayws.WS, env *reqenvelope.T) {
	var err E
	log.I.S(env)
	if env.Filters == nil {
		return
	}
	for i, f := range env.Filters.F {
		log.I.F("%d: %s", i, f)
	}
	if err = ws.WriteEnvelope(eoseenvelope.NewFrom(env.Subscription));chk.E(err){
		return
	}
}
