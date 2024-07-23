package relay

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/protocol/relayinfo"
	"github.com/mleku/nodl/pkg/util/context"
	"github.com/mleku/nodl/pkg/util/units"
	"github.com/puzpuzpuz/xsync/v2"
)

type (
	// OverwriteRelayInformation is a function that can be provided to rewrite the nip-11 relay information document.
	OverwriteRelayInformation func(c context.T, r *http.Request, info *relayinfo.T) *relayinfo.T
	// Hook is a function that does anything but responds to a context cancel from
	// the controlling process so it can be shut down along with the calling
	// process.
	Hook func(c context.T)
	// RejectEvent checks whether policy would reject an event.
	RejectEvent func(c context.T, ev *event.T) (rej bool, msg string)
	// Events are a closure type that responds to an event.
	Events func(c context.T, ev *event.T) error
	// OnEventSaved runs when an event is stored.
	OnEventSaved func(c context.T, ev *event.T)
	// OverwriteResponseEvent rewrites an event response.
	OverwriteResponseEvent func(c context.T, ev *event.T)
	// R is the data structure holding state of the relay.
	R struct {
		Ctx                    context.T
		serveMux               *http.ServeMux
		Info                   *relayinfo.T
		OverwriteRelayInfo     []OverwriteRelayInformation
		OverwriteResponseEvent []OverwriteResponseEvent
		RejectEvent            []RejectEvent
		StoreEvent             []Events
		OnEventSaved           []OnEventSaved
		OnConnect              []Hook
		OnDisconnect           []Hook
		WriteWait              time.Duration // WriteWait is the time allowed to write a message to the peer.
		PongWait               time.Duration // PongWait is the time allowed to read the next pong message from the peer.
		PingPeriod             time.Duration // PingPeriod is the tend pings to peer with this period. Must be less than pongWait.
		upgrader               websocket.Upgrader
		clients                *xsync.MapOf[*websocket.Conn, struct{}]
		Whitelist              []string // whitelist of allowed IPs for access
	}
)

const (
	wsKey = iota
	subscriptionIdKey
	IgnoreAfter    = 16
	MaxMessageSize = 4 * units.Mb
)

func getServiceBaseURL(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if host == "localhost" {
			proto = "http"
		} else if strings.Index(host, ":") != -1 {
			// has a port number
			proto = "http"
		} else if _, err := strconv.Atoi(strings.ReplaceAll(host, ".",
			"")); chk.E(err) {
			// it's a naked IP
			proto = "http"
		} else {
			proto = "https"
		}
	}
	return proto + "://" + host
}
