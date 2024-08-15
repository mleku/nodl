package relay

import (
	"hash/maphash"
	"net/http"
	"sync"
	"time"
	"unsafe"

	"git.replicatr.dev/pkg/codec/filters"
	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/ratel"
	"git.replicatr.dev/pkg/util/atomic"
	C "git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/interrupt"
	"git.replicatr.dev/pkg/util/lol"
	"git.replicatr.dev/pkg/util/units"
	W "github.com/fasthttp/websocket"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/rs/cors"
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

type Subscription struct {
	Initiated time.Time
	Filters   filters.T
}

type Subscriptions map[S]Subscription

type T struct {
	Ctx             C.T
	Cancel          C.F
	WG              sync.WaitGroup
	ListenAddresses []S
	serviceURL      atomic.String
	upgrader        W.Upgrader
	serveMux        *http.ServeMux
	clients         *xsync.MapOf[*relayws.WS, Subscriptions]
	identity        *p256k.Signer
	Store           eventstore.I
}

func (rl T) Init(path S) (r *T) {
	var err E
	rl.Ctx, rl.Cancel = C.Cancel(C.Bg())
	interrupt.AddHandler(func() {
		rl.Cancel()
	})
	rl.upgrader = W.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	if rl.clients == nil {
		rl.clients = xsync.NewTypedMapOf[*relayws.WS, Subscriptions](PointerHasher[relayws.WS])
	}
	rl.identity = &p256k.Signer{}
	if err = rl.identity.Generate(); chk.E(err) {
	}
	rl.Store = ratel.GetBackend(rl.Ctx, &rl.WG, false, MaxMessageSize, N(lol.Level.Load()))
	if err = rl.Store.Init(path); chk.E(err) {
		return
	}
	interrupt.AddHandler(func() {
		chk.E(rl.Store.Close())
	})
	return &rl
}

func PointerHasher[V any](_ maphash.Seed, k *V) uint64 {
	return uint64(uintptr(unsafe.Pointer(k)))
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
