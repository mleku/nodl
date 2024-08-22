package eventstore

import (
	"net/http"

	"nostr.mleku.dev/codec/envelopes/okenvelope"
	"nostr.mleku.dev/codec/subscriptionid"
	"nostr.mleku.dev/protocol/relayws"
)

type SubID = subscriptionid.T
type WS = *relayws.WS
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
