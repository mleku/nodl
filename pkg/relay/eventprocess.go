package relay

import (
	"bytes"

	evEnv "git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/util/hex"
	"github.com/minio/sha256-simd"
)

func (rl *R) processEventEnvelope(msg B, env *evEnv.Submission, c Ctx, ws WS, svcURL S) (err E) {
	var ok bool
	if !rl.IsAuthed(c, "EVENT") {
		return
	}
	// reject old dated events according to nip11
	if *env.T.CreatedAt <= rl.Info.Limitation.Oldest {
		log.D.F("rejecting event with date: %s %s %s", env.T.CreatedAt.Time().String(), ws.Remote(), ws.AuthPub())
		if err = NewOK(NewEID(env.ID), false,
			Reason(Invalid, "relay limit disallows timestamps older than %d",
				rl.Info.Limitation.Oldest)).Write(ws); chk.E(err) {
			return
		}
		return
	}
	// check id
	evs := env.T.ToCanonical()
	hash := sha256.Sum256(evs)
	id := hex.EncAppend(nil, hash[:])
	if !equals(id, env.T.ID) {
		j, _ := env.T.MarshalJSON(nil)
		log.D.F("id mismatch got %s, expected %s %s %s\n%s\n%s", ws.Remote(), ws.AuthPub(), id, env.ID, j, msg)
		if err = NewOK(NewEID(env.ID), false, Reason(Invalid, "id is computed incorrectly")).Write(ws); chk.E(err) {
			return
		}
		return
	}
	// check signature
	if ok, err = env.T.Verify(); chk.E(err) {
		chk.E(NewOK(NewEID(env.ID), false, Reason(Error, "failed to verify signature: "+err.Error())).Write(ws))
		return
	} else if !ok {
		log.D.Ln("invalid: signature is invalid", ws.Remote(), ws.AuthPub())
		if err = NewOK(NewEID(env.ID), false, Reason(Invalid, "signature is invalid")).Write(ws); chk.E(err) {
			return
		}
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
	if err != nil {
		reason = B(err.Error())
		if bytes.HasPrefix(reason, reasons.AuthRequired) {
			log.I.Ln("requesting auth")
			RequestAuth(c, env.Label())
			ok = true
		}
		if bytes.HasPrefix(reason, reasons.Duplicate) {
			ok = true
		}
	} else {
		ok = true
	}
	if err = NewOK(NewEID(env.ID), ok, reason).Write(ws); chk.E(err) {
		return
	}
	return
}
