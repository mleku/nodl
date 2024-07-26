package relay

import (
	"github.com/mleku/nodl/pkg/codec/envelopes/closedenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/countenvelope"
)

func (rl *R) processCountEnvelope(_ B, env *countenvelope.Request, c Ctx, ws WS, _ S) (err E) {
	if !rl.IsAuthed(c, countenvelope.L) {
		return
	}
	if rl.CountEvents == nil {
		chk.E(closedenvelope.NewFrom(env.ID, Reason(Unsupported, "this relay does not support NIP-45")).Write(ws))
		return
	}
	var total int
	for _, f := range env.Filters.F {
		total += rl.handleCountRequest(c, env.ID, ws, f)
	}
	chk.E(countenvelope.NewResponseFrom(env.ID, total, false).Write(ws))
	return
}
