package relay

import (
	"hash/maphash"
	"net/http"
	"time"
	"unsafe"

	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/atomic"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/interrupt"
	"git.replicatr.dev/pkg/util/units"
	"github.com/fasthttp/websocket"
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
)

type T struct {
	Ctx           context.T
	Cancel        context.F
	ListenAddress S
	serviceURL    atomic.String
	upgrader      websocket.Upgrader
	serveMux      *http.ServeMux
	clients       *xsync.MapOf[*relayws.WS, struct{}]
}

func (rl T) Init() *T {
	rl.Ctx, rl.Cancel = context.Cancel(context.Bg())
	interrupt.AddHandler(func(){
		rl.Cancel()
	})
	rl.upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	if rl.clients == nil {
		rl.clients = xsync.NewTypedMapOf[*relayws.WS,
			struct{}](PointerHasher[relayws.WS])
	}
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
	} else {
		rl.serveMux.ServeHTTP(w, r)
	}
	return
}
