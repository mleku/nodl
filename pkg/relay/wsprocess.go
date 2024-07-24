package relay

import (
	"fmt"

	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/envelopes/authenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/closeenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/countenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/eventenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/reqenvelope"
	"github.com/mleku/nodl/pkg/util/normalize"
)

func (rl *R) wsProcessMessages(msg B, c Ctx, kill func(), ws WS, serviceURL S) (err E) {
	if len(msg) == 0 {
		err = log.E.Err("empty message, probably dropped connection")
		return
	}
	if ws.OffenseCount.Load() > IgnoreAfter {
		err = log.E.Err("client keeps sending wrong req envelopes")
		return
	}
	// log.I.F("websocket receive message\n%s\n%s %s",
	//  string(msg), ws.RealRemote(), ws.AuthPubKey())
	strMsg := string(msg)
	if ws.OffenseCount.Load() > IgnoreAfter {
		if len(strMsg) > 256 {
			strMsg = strMsg[:256]
		}
		log.T.F("dropping message due to over %d errors from this client on this connection %s %0x %s",
			IgnoreAfter, ws.RealRemote(), ws.AuthPubKey(), strMsg)
		return
	}
	if len(msg) > rl.Info.Limitation.MaxMessageLength {
		log.D.F("rejecting event with size: %d from %s %s", len(msg), ws.RealRemote(), ws.AuthPubKey())
		chk.E(ws.WriteEnvelope(&OK{
			OK: false,
			Reason: normalize.Reason(fmt.Sprintf(
				"relay limit disallows messages larger than %d "+
					"bytes, this message is %d bytes",
				rl.Info.Limitation.MaxMessageLength, len(msg)), okenvelope.Invalid.S()),
		}))
		return
	}
	deny := true
	if len(rl.Whitelist) > 0 {
		for i := range rl.Whitelist {
			if rl.Whitelist[i] == ws.RealRemote() {
				deny = false
			}
		}
	} else {
		deny = false
	}
	if deny {
		log.E.F("denying access to '%s' %s: dropping message", ws.RealRemote(),
			ws.AuthPubKey())
		return
	}
	var l S
	if l, msg, err = envelopes.Identify(msg); chk.E(err) {
		return
	}
	switch l {
	case eventenvelope.L:
		sub := eventenvelope.NewSubmission()
		if msg, err = sub.MarshalJSON(msg); chk.E(err) {
			return
		}
		// if err = rl.processEventEnvelope(msg, env, c, ws, serviceURL); err != nil {
		// 	return
		// }
	case countenvelope.L:
		count := countenvelope.New()
		if msg, err = count.MarshalJSON(msg); chk.E(err) {
			return
		}
		// if err = rl.processCountEnvelope(msg, env, c, ws, serviceURL); chk.E(err) {
		// 	return
		// }
	case reqenvelope.L:
		req := reqenvelope.New()
		if msg, err = req.MarshalJSON(msg); chk.E(err) {
			return
		}
		// if err = rl.processReqEnvelope(msg, env, c, ws, serviceURL); chk.E(err) {
		// 	return
		// }
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
		if err = rl.processAuthEnvelope(msg, response, ws, serviceURL); chk.E(err) {
			return
		}
	}
	return
}
