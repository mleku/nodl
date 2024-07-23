package relay

import (
	"fmt"
	"strings"

	"github.com/minio/sha256-simd"
	evEnv "github.com/mleku/nodl/pkg/codec/envelopes/eventenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	"github.com/mleku/nodl/pkg/codec/eventid"
	"github.com/mleku/nodl/pkg/codec/kind"
	"github.com/mleku/nodl/pkg/protocol/auth"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
	"github.com/mleku/nodl/pkg/util/hex"
	"github.com/mleku/nodl/pkg/util/normalize"
)

func (rl *R) processEventEnvelope(msg []byte, env *evEnv.Submission,
	c context.T, ws *relayws.WS, serviceURL S) (err error) {

	var ok bool
	if !rl.IsAuthed(c, "EVENT") {
		return
	}
	// reject old dated events according to nip11
	if *env.Event.CreatedAt <= rl.Info.Limitation.Oldest {
		log.D.F("rejecting event with date: %s %s %s",
			env.Event.CreatedAt.Time().String(), ws.RealRemote(),
			ws.AuthPubKey())
		chk.E(ws.WriteEnvelope(&okenvelope.T{
			EventID: eventid.NewWith(env.Event.ID),
			OK:      false,
			Reason: B(fmt.Sprintf(
				"invalid: relay limit disallows timestamps older than %d",
				rl.Info.Limitation.Oldest)),
		}))
		return
	}
	// check id
	evs := env.Event.ToCanonical()
	hash := sha256.Sum256(evs)
	id := hex.EncAppend(nil, hash[:])
	if !equals(id, env.Event.ID) {
		j, _ := env.Event.MarshalJSON(nil)
		log.D.F("id mismatch got %s, expected %s %s %s\n%s\n%s",
			ws.RealRemote(), ws.AuthPubKey(), id, env.Event.ID, j, msg)
		chk.E(ws.WriteEnvelope(&okenvelope.T{
			EventID: eventid.NewWith(env.Event.ID),
			OK:      false,
			Reason:  normalize.Reason(okenvelope.Invalid, "id is computed incorrectly"),
		}))
		return
	}
	// check signature
	if ok, err = env.Event.CheckSignature(); chk.E(err) {
		chk.E(ws.WriteEnvelope(&okenvelope.T{
			EventID: eventid.NewWith(env.Event.ID),
			OK:      false,
			Reason: normalize.Reason(okenvelope.Error,
				okenvelope.Reason("failed to verify signature: "+err.Error())),
		}))
		return
	} else if !ok {
		log.E.Ln("invalid: signature is invalid", ws.RealRemote(),
			ws.AuthPubKey())
		chk.E(ws.WriteEnvelope(&okenvelope.T{
			EventID: eventid.NewWith(env.Event.ID),
			OK:      false,
			Reason:  normalize.Reason(okenvelope.Invalid, "signature is invalid")}))
		return
	}
	if env.Event.Kind == kind.Deletion {
		// this always returns "blocked: " whenever it returns an error
		err = rl.handleDeleteRequest(c, env.Event)
	} else {
		// this will also always return a prefixed reason
		err = rl.AddEvent(c, env.Event)
	}
	var reason string
	if err != nil {
		reason = err.Error()
		if strings.HasPrefix(reason, auth.Required) {
			log.I.Ln("requesting auth")
			RequestAuth(c, env.Label())
			ok = true
		}
		if strings.HasPrefix(reason, "duplicate") {
			ok = true
		}
	} else {
		ok = true
	}
	chk.E(ws.WriteEnvelope(&okenvelope.T{
		EventID: eventid.NewWith(env.Event.ID),
		OK:      ok,
		Reason:  B(reason),
	}))

	return
}
