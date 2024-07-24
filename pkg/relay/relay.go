package relay

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/mleku/nodl/pkg/protocol/relayinfo"
	"github.com/mleku/nodl/pkg/util/units"
	"github.com/puzpuzpuz/xsync/v2"
)

type (
	// OverwriteRelayInformation is a function that can be provided to rewrite the nip-11 relay information document.
	OverwriteRelayInformation  func(c Ctx, r Req, info *relayinfo.T) *relayinfo.T
	OverwriteRelayInformations []OverwriteRelayInformation
	// Hook is a function that does anything but responds to a context cancel from
	// the controlling process so it can be shut down along with the calling
	// process.
	Hook  func(c Ctx)
	Hooks []Hook
	// RejectEvent checks whether policy would reject an event.
	RejectEvent  func(c Ctx, ev EV) (rej bool, msg string)
	RejectEvents []RejectEvent
	// Event are a closure type that responds to an event.
	Event  func(c Ctx, ev EV) E
	Events []Event
	// OnEventSaved runs when an event is stored.
	OnEventSaved  func(c Ctx, ev EV)
	OnEventSaveds []OnEventSaved
	// OverwriteResponseEvent rewrites an event response.
	OverwriteResponseEvent  func(c Ctx, ev EV)
	OverwriteResponseEvents []OverwriteResponseEvent
	// R is the data structure holding state of the relay.
	R struct {
		Ctx                    Ctx
		Router                 *http.ServeMux
		Info                   *relayinfo.T
		OverwriteRelayInfo     OverwriteRelayInformations
		OverwriteResponseEvent OverwriteResponseEvents
		RejectEvent            RejectEvents
		StoreEvent             Events
		OnEventSaved           OnEventSaveds
		OnConnect              Hooks
		OnDisconnect           Hooks
		WriteWait              time.Duration // WriteWait is the time allowed to write a message to the peer.
		PongWait               time.Duration // PongWait is the time allowed to read the next pong message from the peer.
		PingPeriod             time.Duration // PingPeriod is the time between pings. Must be less than pongWait.
		upgrader               websocket.Upgrader
		clients                *xsync.MapOf[*websocket.Conn, struct{}]
		Whitelist              []S // whitelist of allowed IPs for access
	}
)

const (
	wsKey = iota
	subscriptionIdKey
	IgnoreAfter    = 16
	MaxMessageSize = 4 * units.Mb
)

func getServiceBaseURL(r Req) S {
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
