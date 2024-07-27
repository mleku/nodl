package relay

import (
	"net/http"

	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/noticeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/okenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/normalize"
)

type (
	Ctx       = context.T
	SubID     = *subscriptionid.T
	WS        = *relayws.WS
	Responder = http.ResponseWriter
	Req       = *http.Request
	OK        = okenvelope.T
	Notice    = noticeenvelope.T
	EV        = *event.T
	Handler   = http.HandlerFunc
)

var (
	Reason       = normalize.Reason
	AuthRequired = reasons.AuthRequired
	Blocked      = reasons.Blocked
	Duplicate    = reasons.Duplicate
	Error        = reasons.Error
	Invalid      = reasons.Invalid
	Unsupported  = reasons.Unsupported
	NewOK        = okenvelope.NewFrom
	NewEID       = eventid.NewWith[B]
	NewNotice    = noticeenvelope.NewFrom[B]
	NewResult    = eventenvelope.NewResultWith
)
