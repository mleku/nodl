package eventstore

import (
	"net/http"
	"nostr.mleku.dev/protocol/ws"

	"nostr.mleku.dev/codec/envelopes/okenvelope"
	"nostr.mleku.dev/codec/subscriptionid"
)

type SubID = subscriptionid.T
type WS = *ws.Serv
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
