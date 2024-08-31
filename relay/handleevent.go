package relay

import (
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/envelopes/eventenvelope"
	"nostr.mleku.dev/codec/envelopes/noticeenvelope"
	"nostr.mleku.dev/codec/envelopes/okenvelope"
	"nostr.mleku.dev/protocol/ws"
	"util.mleku.dev/normalize"
)

func (rl *T) handleEvent(ws *ws.Serv, env *eventenvelope.Submission) {
	var err E
	var ok bool
	// verify the event ID
	actualId := env.T.GetIDBytes()
	if !Equals(actualId, env.T.ID) {
		if err = okenvelope.NewFrom(env.T.EventID(), false,
			normalize.Error.Message("event ID %0x is incorrect, should be %0x",
				env.T.ID, actualId)).Write(ws); Chk.E(err) {
			return
		}
		return
	}
	// verify the signature
	if ok, err = env.T.Verify(); Chk.E(err) {
		// some error occurred while verifying signature
		if err = okenvelope.NewFrom(env.T.EventID(), false,
			normalize.Error.Message(err.Error())).Write(ws); Chk.E(err) {
			return
		}
		return
	}
	// signature was invalid
	if !ok {
		if err = okenvelope.NewFrom(env.T.EventID(), false, normalize.
			Invalid.Message("event signature failed verification")).Write(ws); Chk.E(err) {
			return
		}
	}
	// event was acceptable.
	if err = okenvelope.NewFrom(env.T.EventID(), true, B{}).Write(ws); Chk.E(err) {
		return
	}
	// save event to event store.
	if err = rl.Store.SaveEvent(rl.Ctx, env.T); Chk.E(err) {
		// if an error occurred, notify the
		if err = noticeenvelope.NewFrom(normalize.
			Error.Message("failed saving event %0x: %s",
			env.T.EventID(), err.Error())).Write(ws); Chk.E(err) {
			return
		}
		return
	}
	// check if a subscription filter matches

}
