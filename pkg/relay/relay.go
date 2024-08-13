package relay

import (
	"net/http"

	"git.replicatr.dev/pkg/util/atomic"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
)

type T struct {
	ListenAddress S
	ServiceURL    atomic.String
	Upgrader      websocket.Upgrader
	serveMux      *http.ServeMux
}

func (rl *T) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rl.ServiceURL.Load() == "" {
		rl.ServiceURL.Store(getServiceBaseURL(r))
	}
	if r.Header.Get("Upgrade") == "websocket" {
		rl.HandleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		cors.AllowAll().Handler(http.HandlerFunc(rl.HandleRelayInfo)).ServeHTTP(w, r)
	} else {
		rl.serveMux.ServeHTTP(w, r)
	}
	return
}
