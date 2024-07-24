package eventstore

import (
	"net/http"

	"github.com/mleku/nodl/pkg/codec/envelopes/okenvelope"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
)

type Ctx = context.T
type SubID = subscriptionid.T
type WS = *relayws.WS
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
type EV = *event.T