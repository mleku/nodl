package relay

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.replicatr.dev/pkg/codec/bech32encoding"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/crypto/keys"
	"git.replicatr.dev/pkg/protocol/relayinfo"
	"git.replicatr.dev/pkg/relay/acl"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/units"
	"github.com/fasthttp/websocket"
	"github.com/puzpuzpuz/xsync/v2"
)

var Version = "v0.0.1"
var Software = "https://git.replicatr.dev"

const (
	wsKey = iota
	subscriptionIdKey
)

const (
	WriteWait       = 3 * time.Second
	PongWait        = 6 * time.Second
	PingPeriod      = 3 * time.Second
	ReadBufferSize  = 4096
	WriteBufferSize = 4096
	IgnoreAfter     = 16
	MaxMessageSize  = 4 * units.Mb
)

type (
	// OverwriteRelayInfo is a function that can be provided to rewrite the nip-11 relay information document.
	OverwriteRelayInfo  func(c Ctx, r Req, info *relayinfo.T) *relayinfo.T
	OverwriteRelayInfos []OverwriteRelayInfo
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
	// CountEvent is a function to count the events that match the filter.
	CountEvent  func(c Ctx, f *filter.T) (cnt int, err error)
	CountEvents []CountEvent
	// OnEventSaved runs when an event is stored.
	OnEventSaved  func(c Ctx, ev EV)
	OnEventSaveds []OnEventSaved
	// OverwriteResponseEvent rewrites an event response.
	OverwriteResponseEvent  func(c Ctx, ev EV)
	OverwriteResponseEvents []OverwriteResponseEvent
	// OverwriteFilter rewrites the content of a filter according to policy.
	OverwriteFilter  func(c Ctx, f *filter.T)
	OverwriteFilters []OverwriteFilter
	// RejectFilter is a closure to examine a filter and if it is invalid, reject it.
	RejectFilter  func(c Ctx, id SubID, f *filter.T) (reject bool, msg B)
	RejectFilters []RejectFilter
	// QueryEvent is a closure for making a query to the event store.
	QueryEvent  func(c Ctx, f *filter.T) (C event.C, err error)
	QueryEvents []QueryEvent
	// OverrideDeletion checks a delete request and overrides it according to policy.
	OverrideDeletion  func(c Ctx, tgt, del *event.T) (ok bool, msg B)
	OverrideDeletions []OverrideDeletion
	// R is the data structure holding state of the relay.
	R struct {
		Ctx            Ctx
		Cancel         context.F
		WG             *sync.WaitGroup
		Router         *http.ServeMux
		Info           *relayinfo.T
		Config         *Config
		MaxMessageSize int64 // Maximum message size allowed from peer.
		RelayPubHex    S
		RelayNpub      S
		OverwriteRelayInfos
		OverwriteResponseEvents
		OverwriteReqFilter   OverwriteFilters
		OverwriteCountFilter OverwriteFilters
		OverrideDeletions
		RejectEvents
		RejectReqFilters   RejectFilters
		RejectCountFilters RejectFilters
		QueryEvents
		StoreEvents  Events
		DeleteEvents Events
		CountEvents
		OnEventSaveds
		OnConnects    Hooks
		OnDisconnects Hooks
		WriteWait     time.Duration // WriteWait is the time allowed to write a message to the peer.
		PongWait      time.Duration // PongWait is the time allowed to read the next pong message from the peer.
		PingPeriod    time.Duration // PingPeriod is the time between pings. Must be less than pongWait.
		upgrader      websocket.Upgrader
		clients       *xsync.MapOf[*websocket.Conn, struct{}]
		Whitelist     []S    // whitelist of allowed IPs for access
		ACL           *acl.T // ACL is the list of users and privileges on this relay
	}
)

func NewRelay(c Ctx, cancel context.F, inf *relayinfo.T, conf *Config) (r *R) {
	var maxMessageLength = MaxMessageSize
	if inf.Limitation.MaxMessageLength > 0 {
		maxMessageLength = inf.Limitation.MaxMessageLength
	}
	var err E
	var pubKey S
	if pubKey, err = keys.GetPublicKey(conf.SecKey); chk.E(err) {
		return
	}
	var npub B
	if npub, err = bech32encoding.HexToNpub(B(pubKey)); chk.E(err) {
		return
	}
	inf.Software = Software
	inf.Version = Version
	inf.PubKey = pubKey
	r = &R{
		Ctx:    c,
		Cancel: cancel,
		Config: conf,
		Info:   inf,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		clients: xsync.NewTypedMapOf[*websocket.Conn,
			struct{}](PointerHasher[websocket.Conn]),
		Router:         &http.ServeMux{},
		WriteWait:      WriteWait,
		PongWait:       PongWait,
		PingPeriod:     PingPeriod,
		MaxMessageSize: int64(maxMessageLength),
		Whitelist:      conf.Whitelist,
		RelayPubHex:    pubKey,
		RelayNpub:      S(npub),
	}
	log.I.F("relay identity pubkey: %s %s\n", pubKey, npub)
	// populate ACL with owners to start
	for _, owner := range r.Config.Owners {
		if err = r.ACL.AddEntry(&acl.Entry{
			Role:   acl.Owner,
			Pubkey: acl.B(owner),
		}); chk.E(err) {
			continue
		}
		log.D.Ln("added owner pubkey", owner)
	}
	return
}

func getServiceBaseURL(r Req) (svcURL S) {
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
	svcURL = proto + "://" + host
	return
}
