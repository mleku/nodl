package relay

import (
	"fmt"

	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/noticeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/okenvelope"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/normalize"
)

func (rl *T) handleEvent(ws *relayws.WS, env *eventenvelope.Submission) {
	var err E
	var ok bool
	if ok, err = env.T.Verify(); chk.E(err) {
		// some error occurred while verifying signature
		if err = ws.WriteEnvelope(okenvelope.NewFrom(
			env.T.EventID(),
			false,
			normalize.Reason(reasons.Error, err.Error())),
		); chk.E(err) {
			return
		}
		return
	}
	// signature was in valid
	if !ok {
		if err = ws.WriteEnvelope(okenvelope.NewFrom(
			env.T.EventID(),
			false,
			normalize.Reason(reasons.Invalid, "event signature failed verification")),
		); chk.E(err) {
			return
		}
	}
	// event was acceptable.
	if err = ws.WriteEnvelope(okenvelope.NewFrom(env.T.EventID(), true, B{})); chk.E(err) {
		return
	}
	// save event to event store.
	if err = rl.Store.SaveEvent(rl.Ctx, env.T); chk.E(err) {
		// if an error occurred, notify the
		if err = ws.WriteEnvelope(noticeenvelope.NewFrom(
			normalize.Reason(reasons.Error, fmt.Sprintf("failed saving event %0x: %s",
				env.T.EventID(), err.Error()))),
		); chk.E(err) {
			return
		}
		return
	}
}
