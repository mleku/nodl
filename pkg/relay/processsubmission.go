package relay

import (
	"bytes"

	"github.com/mleku/nodl/pkg/codec/envelopes/eventenvelope"
	"github.com/mleku/nodl/pkg/codec/kind"
)

func (h *Handle) processEventSubmission(msg B, env *eventenvelope.Submission) (err E) {
	rl, c, ws, _, _ := h.H()
	// if auth is required and not authed, reject
	if !rl.IsAuthed(c, eventenvelope.L) {
		return
	}
	// reject old dated events according to nip11
	if env.T.CreatedAt.I64() <= rl.Info.Limitation.Oldest.I64() {
		log.D.F("rejecting event with date: %s %s %s", env.T.CreatedAt.Time().String(), ws.Remote(), ws.AuthPub())
		chk.E(NewOK(NewEID(env.ID), false, Reason(Invalid, "invalid: relay limit disallows timestamps older than %d",
			rl.Info.Limitation.Oldest)).Write(ws))
		return
	}
	// check id
	id := env.T.GetIDBytes()
	if !equals(id, env.T.ID) {
		j, _ := env.T.MarshalJSON(B{})
		log.D.F("id mismatch got %s, expected %s %s %s\n%s\n%s", ws.Remote(), ws.AuthPub(), id, env.T.ID, j, msg)
		chk.E(NewOK(NewEID(env.ID), false, Reason(Invalid, "id is computed incorrectly")).Write(ws))
		return
	}
	// check signature
	var ok bool
	if ok, err = env.T.Verify(); chk.E(err) {
		chk.E(NewOK(NewEID(env.ID), false, Reason(Error, "failed to verify signature: `%s`", err.Error())).Write(ws))
		return
	} else if !ok {
		log.E.Ln("invalid: signature is invalid", ws.Remote(), ws.AuthPub())
		chk.E(NewOK(NewEID(env.ID), false, Reason(Invalid, "signature is invalid")).Write(ws))
		return
	}
	if env.T.Kind == kind.Deletion {
		// this always returns "blocked: " whenever it returns an error
		err = rl.handleDeleteRequest(c, env.T)
	} else {
		// this will also always return a prefixed reason
		err = rl.AddEvent(c, env.T)
	}
	var reason B
	ok = true
	if err != nil {
		reason = B(err.Error())
		if bytes.HasPrefix(reason, B(AuthRequired)) {
			log.I.Ln("requesting auth")
			RequestAuth(c, env.Label())
			ok = true
		} else {
			ok = bytes.HasPrefix(reason, B(Duplicate))
		}
	} else {
		ok = true
	}
	chk.E(NewOK(NewEID(env.ID), false, reason).Write(ws))
	return
}
