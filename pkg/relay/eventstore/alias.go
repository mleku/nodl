package eventstore

import (
	"net/http"

	"git.replicatr.dev/pkg/codec/envelopes/okenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
)

type Ctx = context.T
type SubID = subscriptionid.T
type WS = *relayws.WS
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
type EV = *event.T