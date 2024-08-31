package relay

import (
	"net/http"
	"sync"
	"time"

	"git.replicatr.dev/eventstore"
	"git.replicatr.dev/eventstore/ratel"
	. "nostr.mleku.dev"

	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"nostr.mleku.dev/crypto/p256k"
	"nostr.mleku.dev/protocol/relayinfo"
	"util.mleku.dev/atomic"
	"util.mleku.dev/context"
	"util.mleku.dev/interrupt"
	"util.mleku.dev/lol"
	"util.mleku.dev/units"
)

const (
	WriteWait           = 3 * time.Second
	PongWait            = 6 * time.Second
	PingPeriod          = 3 * time.Second
	ReadBufferSize      = 4096
	WriteBufferSize     = 4096
	MaxMessageSize  int = 4 * units.Mb
	DefaultLimit        = 50
	MaxLimit            = 500
	DBSizeLimit         = 0 // disables GC
	DBLowWater          = 86
	DBHighWater         = 92
	GCFrequency         = 300
)

// T is the state and configuration data of a relay.
//
// ClientMap keeps track of current websocket connections that are open. Access only with Mutex
// locked.
//
// subMap keeps track of distinctive filters for which each associates with a websocket
// connection. Access only with Mutex locked, using Tracker.Do.
type T struct {
	Ctx             context.T
	Cancel          context.F
	WG              sync.WaitGroup
	ListenAddresses []S
	serviceURL      atomic.String
	upgrader        websocket.Upgrader
	relayInfo       *relayinfo.T
	serveMux        *http.ServeMux
	identity        *p256k.Signer
	Store           eventstore.I
	Tracker
}

func (rl T) Init(path S) (r *T, err E) {
	rl.Ctx, rl.Cancel = context.Cancel(context.Bg())
	interrupt.AddHandler(func() { rl.Cancel() })
	rl.upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	rl.Tracker.Do(func() { rl.Tracker.Init(rl.Ctx) })
	rl.identity = &p256k.Signer{}
	if err = rl.identity.Generate(); Chk.E(err) {
	}
	rl.Store = ratel.GetBackend(rl.Ctx, &rl.WG, false, MaxMessageSize, N(lol.Level.Load()))
	if err = rl.Store.Init(path); Chk.E(err) {
		return
	}
	interrupt.AddHandler(func() {
		Chk.E(rl.Store.Close())
	})
	r = &rl
	return
}

func (rl *T) ServiceURL() S { return rl.serviceURL.Load() }

func (rl *T) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rl.serviceURL.Load() == "" {
		rl.serviceURL.Store(getServiceBaseURL(r))
	}
	if r.Header.Get("Upgrade") == "websocket" {
		rl.HandleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		cors.AllowAll().Handler(http.HandlerFunc(rl.HandleRelayInfo)).ServeHTTP(w, r)
	}
	return
}
