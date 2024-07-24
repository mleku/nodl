package relay

import (
	"strings"

	ae "github.com/mleku/nodl/pkg/codec/envelopes/authenvelope"
	oke "github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	eid "github.com/mleku/nodl/pkg/codec/eventid"
	"github.com/mleku/nodl/pkg/protocol/auth"
	"github.com/mleku/nodl/pkg/util/normalize"
)

func (rl *R) processAuthEnvelope(msg B, env *ae.Response, ws WS, serviceURL S) (err E) {
	log.T.Ln("received auth response")
	wsBaseUrl := strings.Replace(serviceURL, "http", "ws", 1)
	var ok bool
	if ok, err = auth.Validate(env.Event, ws.Challenge(),
		wsBaseUrl); ok {
		if equals(ws.AuthPubKey(), env.Event.PubKey) {
			log.D.Ln("user already authed")
			return
		}
		log.I.F("user authenticated %0x", env.Event.PubKey)
		ws.SetAuthPubKey(env.Event.PubKey)
		log.I.Ln("closing auth chan")
		close(ws.Authed)
		chk.E(ws.WriteEnvelope(&oke.T{
			EventID: eid.NewWith(env.Event.ID),
			OK:      true,
		}))
		return
	} else {
		log.E.Ln("user sent bogus auth response\n%s", msg)
		chk.E(ws.WriteEnvelope(&oke.T{
			EventID: eid.NewWith(env.Event.ID),
			OK:      false,
			Reason:  normalize.Reason(oke.Error, "failed to authenticate"),
		}))
	}
	return
}
