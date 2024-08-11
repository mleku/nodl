package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/closeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/countenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/codec/eventid"
)

func (h *Handle) wsProcessMessages(msg B) (err E) {
	log.T.F("message received:\n%s\n", msg)
	rl, c, ws, svcURL, _ := h.H()
	if len(msg) == 0 {
		err = log.E.Err("empty message, probably dropped connection")
		return
	}
	if ws.OffenseCount.Load() > IgnoreAfter {
		err = log.E.Err("client keeps sending wrong req envelopes")
		return
	}
	// log.I.F("websocket receive message\n%s\n%s %s",
	//  S(msg), ws.Remote(), ws.AuthPub())
	strMsg := S(msg)
	if ws.OffenseCount.Load() > IgnoreAfter {
		if len(strMsg) > 256 {
			strMsg = strMsg[:256]
		}
		log.T.F("dropping message due to over %d errors from this client on this connection %s %0x %s",
			IgnoreAfter, ws.Remote(), ws.AuthPub(), strMsg)
		return
	}
	if len(msg) > int(rl.MaxMessageSize) {
		log.D.F("rejecting event with size: %d from %s %s", len(msg), ws.Remote(), ws.AuthPub())
		chk.E(NewOK(&eventid.T{}, false,
			Reason(Invalid, "relay limit disallows messages larger than %d bytes, this message is %d bytes",
				rl.Info.Limitation.MaxMessageLength, len(msg))).Write(ws))
		return
	}
	var l S
	if l, msg, err = envelopes.Identify(msg); chk.E(err) {
		return
	}
	switch l {
	case eventenvelope.L:
		sub := eventenvelope.NewSubmission()
		if msg, err = sub.UnmarshalJSON(msg); chk.E(err) {
			return
		}
		if err = h.processEventSubmission(msg, sub); chk.E(err) {
			return
		}
	case countenvelope.L:
		count := countenvelope.New()
		if msg, err = count.UnmarshalJSON(msg); chk.E(err) {
			return
		}
		if err = rl.processCountEnvelope(msg, count, c, ws, svcURL); chk.E(err) {
			return
		}
	case reqenvelope.L:
		req := reqenvelope.New()
		if msg, err = req.UnmarshalJSON(msg); chk.E(err) {
			return
		}
		if err = rl.processReqEnvelope(msg, req, c, ws, svcURL); chk.E(err) {
			return
		}
	case closeenvelope.L:
		clo := closeenvelope.New()
		if msg, err = clo.UnmarshalJSON(msg); chk.E(err) {
			return
		}
		RemoveListenerId(ws, clo.ID)
	case authenvelope.L:
		response := authenvelope.NewResponse()
		if msg, err = response.UnmarshalJSON(msg); chk.E(err) {
			return
		}
		if err = h.processAuthEnvelope(msg, response); chk.E(err) {
			return
		}
	}
	return
}
