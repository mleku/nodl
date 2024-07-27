package relay

import (
	"strings"

	ae "git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	eid "git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/protocol/auth"
)

func (h *Handle) processAuthEnvelope(msg B, env *ae.Response) (err E) {
	_, _, ws, svcURL, _ := h.H()
	log.T.Ln("received auth response")
	wsBaseUrl := strings.Replace(svcURL, "http", "ws", 1)
	var ok bool
	if ok, err = auth.Validate(env.Event, ws.Challenge(), wsBaseUrl); ok {
		if equals(ws.AuthPub(), env.Event.PubKey) {
			log.D.Ln("user already authed")
			return
		}
		log.I.F("user authenticated %0x", env.Event.PubKey)
		ws.SetAuthPubKey(env.Event.PubKey)
		log.I.Ln("closing auth chan")
		close(ws.Authed)
		chk.E(NewOK(eid.NewWith(env.Event.ID), true, B("authenticated")).Write(h.ws))
		return
	} else {
		log.E.Ln("user sent bogus auth response\n%s", msg)
		chk.E(NewOK(eid.NewWith(env.Event.ID), false, Reason(Error, "failed to authenticate")).Write(h.ws))
	}
	return
}
