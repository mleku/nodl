package relay

import (
	"net/http"

	"github.com/mleku/nodl/pkg/codec/envelopes/eventenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/noticeenvelope"
	"github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/eventid"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/protocol/reasons"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
	"github.com/mleku/nodl/pkg/util/normalize"
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
