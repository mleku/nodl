package relay

import (
	"fmt"

	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/noticeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/okenvelope"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/normalize"
)

func (rl *T) handleEvent(ws *relayws.WS, env *eventenvelope.Submission) {
	var err E
	var ok bool
	if ok, err = env.T.Verify(); chk.E(err) {
		// some error occurred while verifying signature
		if err = okenvelope.NewFrom(env.T.EventID(), false,
			normalize.Error.Message(err.Error())).Write(ws); chk.E(err) {
			return
		}
		return
	}
	// signature was in valid
	if !ok {
		if err = okenvelope.NewFrom(env.T.EventID(), false, normalize.
			Invalid.Message("event signature failed verification")).Write(ws); chk.E(err) {
			return
		}
	}
	// event was acceptable.
	if err = okenvelope.NewFrom(env.T.EventID(), true, B{}).Write(ws); chk.E(err) {
		return
	}
	// save event to event store.
	if err = rl.Store.SaveEvent(rl.Ctx, env.T); chk.E(err) {
		// if an error occurred, notify the
		if err = noticeenvelope.NewFrom(normalize.
			Error.Message(fmt.Sprintf("failed saving event %0x: %s",
			env.T.EventID(), err.Error()))).Write(ws); chk.E(err) {
			return
		}
		return
	}
}
