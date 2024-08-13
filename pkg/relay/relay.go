package relay

import (
	"net/http"
	"sync"
	"time"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/protocol/relayinfo"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"github.com/fasthttp/websocket"
)

func New() *Relay {
	rl := &Relay{

		Info: &relayinfo.T{
			Software: "https://github.com/fiatjaf/khatru",
			Version:  "n/a",
			Nips:     []int{1, 11, 42, 70, 86},
		},

		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},

		clients:   make(map[*relayws.WS][]listenerSpec, 100),
		listeners: make([]listener, 0, 100),

		serveMux: &http.ServeMux{},

		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     30 * time.Second,
		MaxMessageSize: 512000,
	}

	return rl
}

type Func func(c context.T)
type FilterFunc func(c context.T, f *filter.T)
type EventFunc func(c context.T, ev *event.T)
type ProcessEventFunc func(c context.T, ev *event.T) E
type RejectFilter func(c context.T, f *filter.T) (reject bool, msg S)

type Relay struct {
	ServiceURL S

	// these structs keeps track of all the things that can be customized when handling events or requests
	RejectEvent              []func(c context.T, ev *event.T) (reject bool, msg S)
	OverwriteDeletionOutcome []func(c context.T, target *event.T,
		deletion *event.T) (accept bool, msg S)
	StoreEvent                []ProcessEventFunc
	DeleteEvent               []ProcessEventFunc
	OnEventSaved              []EventFunc
	OnEphemeralEvent          []EventFunc
	RejectFilter              []RejectFilter
	RejectCountFilter         []RejectFilter
	OverwriteFilter           []FilterFunc
	OverwriteCountFilter      []FilterFunc
	QueryEvents               []func(c context.T, f *filter.T) (event.C, E)
	CountEvents               []func(c context.T, f *filter.T) (N, E)
	RejectConnection          []func(r *http.Request) bool
	OnConnect                 []Func
	OnDisconnect              []Func
	OverwriteRelayInformation []func(c context.T, r *http.Request,
		info *relayinfo.T) *relayinfo.T
	OverwriteResponseEvent []EventFunc
	PreventBroadcast       []func(ws *relayws.WS, ev *event.T) bool

	// these are used when this relays acts as a router
	routes                []Route
	getSubRelayFromEvent  func(*event.T) *Relay // used for handling EVENTs
	getSubRelayFromFilter func(*filter.T) *Relay // used for handling REQs

	// setting up handlers here will enable these methods
	ManagementAPI RelayManagementAPI

	// editing info will affect the NIP-11 responses
	Info *relayinfo.T

	// Default logger, as set by NewServer, is a stdlib logger prefixed with "[khatru-relay] ",
	// outputting to stderr.
	// Log *log.Logger

	// for establishing websockets
	upgrader websocket.Upgrader

	// keep a connection reference to all connected clients for Server.Shutdown
	// also used for keeping track of who is listening to what
	clients      map[*relayws.WS][]listenerSpec
	listeners    []listener
	clientsMutex sync.Mutex

	// in case you call Server.Start
	Addr       S
	serveMux   *http.ServeMux
	httpServer *http.Server

	// websocket options
	WriteWait      time.Duration // Time allowed to write a message to the peer.
	PongWait       time.Duration // Time allowed to read the next pong message from the peer.
	PingPeriod     time.Duration // Send pings to peer with this period. Must be less than pongWait.
	MaxMessageSize int64         // Maximum message size allowed from peer.
}
