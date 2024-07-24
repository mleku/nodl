package relay

import (
	"net/http"

	"github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
)

type (
	Ctx       = context.T
	SubID     = subscriptionid.T
	WS        = *relayws.WS
	Responder = http.ResponseWriter
	Req       = *http.Request
	OK        = okenvelope.T
	EV        = *event.T
	Handler   = http.HandlerFunc
)
