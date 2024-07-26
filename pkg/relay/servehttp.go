package relay

import (
	"github.com/rs/cors"
)

// ServeHTTP implements http.Handler enveloper.
//
// This is the main starting function of the relay. This launches
// HandleWebsocket which runs the message handling main loop.
func (rl *R) ServeHTTP(w Responder, r Req) {
	log.T.C(SprintHeader(r.Header))
	select {
	case <-rl.Ctx.Done():
		log.W.Ln("shutting down")
		return
	default:
	}
	if r.Header.Get("Upgrade") == "websocket" {
		rl.HandleWebsocket(getServiceBaseURL(r))(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		cors.AllowAll().Handler(Handler(rl.HandleNIP11)).ServeHTTP(w, r)
	} else {
		rl.Router.ServeHTTP(w, r)
	}
}
