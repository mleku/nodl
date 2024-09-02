package relay

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.replicatr.dev/relay/types"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"util.mleku.dev/context"
	"util.mleku.dev/qu"

	ws "github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

func GetUpgrader(readBuf, writeBuf int, checkOrigin func(r *http.Request) bool) *ws.Upgrader {
	return &ws.Upgrader{
		ReadBufferSize:  readBuf,
		WriteBufferSize: writeBuf,
		CheckOrigin:     checkOrigin,
	}
}

func GetDefaultUpgrader() *ws.Upgrader {
	return GetUpgrader(1024, 1024,
		func(r *http.Request) bool { return true })
}

const (
	DefaultWriteWait = 10 * time.Second

	DefaultPongWait = 60 * time.Second

	DefaultPingPeriod = DefaultPongWait / 2

	DefaultMaxMessageSize = 512000
)

type Config struct {
	// WriteWait is the time allowed to write a message to the peer.
	WriteWait time.Duration
	// PongWait is the time allowed to read the next pong message from the peer.
	PongWait time.Duration
	// PingPeriod is the frequency to end pings to peer with this period.
	//
	// Must be less than PongWait.
	PingPeriod time.Duration
	// MaxMessageSize is the maximum message size allowed from peer.
	MaxMessageSize int
}

func GetDefaultConfig() *Config {
	return &Config{
		WriteWait:      DefaultWriteWait,
		PongWait:       DefaultPongWait,
		PingPeriod:     DefaultPingPeriod,
		MaxMessageSize: DefaultMaxMessageSize,
	}
}

// Server is a base for package users to implement nostr relays. It can serve HTTP requests and
// websockets, passing control over to a relay implementation.
//
// To implement a relay, it is enough to satisfy [Relayer] types. Other interfaces are
// [Informer], [CustomWebSocketHandler], [ShutdownAware] and AdvancedXxx types. See their
// respective doc comments.
//
// The basic usage is to call Start or StartConf, which starts serving immediately. For a more
// fine-grained control, use NewServer. See [basic/main.go], [whitelisted/main.go],
// [expensive/main.go] and [rss-bridge/main.go] for example implementations.
//
// The following resource is a good starting point for details on what nostr protocol is and how
// it works: https://github.com/nostr-protocol/nostr
type Server struct {
	*Options

	relay types.Relayer

	dbPath S

	// keep a connection reference to all connected clients for Server.Shutdown
	clientsMu sync.Mutex
	clients   map[*ws.Conn]struct{}

	// in case you call Server.Start
	Addr       S
	serveMux   *http.ServeMux
	httpServer *http.Server

	*Config
}

func (s *Server) Router() *http.ServeMux { return s.serveMux }

// NewServer initializes the relay and its storage using their respective Init methods,
// returning any non-nil errors, and returns a Server ready to listen for HTTP requests.
func NewServer(rl types.Relayer, opts ...Option) (*Server, E) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	srv := &Server{
		relay:   rl,
		clients: make(map[*ws.Conn]struct{}),
		Options: options,
	}
	// start listening from events from other sources, if any
	if inj, ok := rl.(types.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				notifyListeners(ev)
			}
		}()
	}
	return srv, nil
}

// ServeHTTP implements http.Handler types.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "ws" {
		s.HandleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		s.HandleNIP11(w, r)
	} else {
		s.serveMux.ServeHTTP(w, r)
	}
}

func (s *Server) Start(host S, port N, started ...qu.C) E {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.Addr = ln.Addr().String()
	s.httpServer = &http.Server{
		Handler:      cors.Default().Handler(s),
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// notify caller that we're starting
	for _, start := range started {
		close(start)
	}

	if err = s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

// Shutdown sends a ws close control message to all connected clients.
//
// If the relay is ShutdownAware, Shutdown calls its OnShutdown, passing the context as is. Note
// that the HTTP server make some time to shutdown and so the context deadline, if any, may have
// been shortened by the time OnShutdown is called.
func (s *Server) Shutdown(c context.T) {
	Log.I.Ln("shutting down relay")
	Log.D.Ln("disconnecting clients")
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		Log.I.S("disconnecting", conn.RemoteAddr())
		conn.WriteControl(ws.CloseMessage, nil, time.Now().Add(time.Second))
		conn.Close()
		delete(s.clients, conn)
	}
	Log.D.Ln("running relay shutdown handlers")
	if f, ok := s.relay.(types.ShutdownAware); ok {
		f.OnShutdown(c)
	}
	s.relay.Storage().Close()
	Log.I.Ln("relay shutdown complete")
	s.httpServer.Shutdown(c)
}

type Option func(*Options)

type SkipEventFunc func(ev *event.T) bool

type Options struct {
	perConnectionLimiter *rate.Limiter
	skipEventFunc        SkipEventFunc
	servMux              *http.ServeMux
	config               *Config
	upgrader             *ws.Upgrader
}

func DefaultOptions() *Options {
	return &Options{
		upgrader: GetDefaultUpgrader(),
		config:   GetDefaultConfig(),
		servMux:  &http.ServeMux{},
	}
}

func WithPerConnectionLimiter(rps rate.Limit, burst int) Option {
	return func(o *Options) { o.perConnectionLimiter = rate.NewLimiter(rps, burst) }
}

func WithSkipEventFunc(skipEventFunc SkipEventFunc) Option {
	return func(o *Options) { o.skipEventFunc = skipEventFunc }
}

func WithConfig(c *Config) Option {
	return func(o *Options) { o.config = c }
}

func WithServMux(mux *http.ServeMux) Option {
	return func(o *Options) { o.servMux = mux }
}
