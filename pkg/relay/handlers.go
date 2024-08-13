package relay

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/closedenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/closeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/countenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/okenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/protocol/auth"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/normalize"
	W "github.com/fasthttp/websocket"
	"github.com/rs/cors"
)

// ServeHTTP implements http.Handler interface.
func (rl *Relay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rl.ServiceURL == "" {
		rl.ServiceURL = getServiceBaseURL(r)
	}
	if r.Header.Get("Upgrade") == "websocket" {
		rl.HandleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		log.I.Ln("nip-11")
		cors.AllowAll().Handler(http.HandlerFunc(rl.HandleNIP11)).ServeHTTP(w, r)
	} else if r.Header.Get("Content-Type") == "application/nostr+json+rpc" {
		log.I.Ln("nip-86")
		cors.AllowAll().Handler(http.HandlerFunc(rl.HandleNIP86)).ServeHTTP(w, r)
	} else {
		log.I.S(r.Header)
		rl.serveMux.ServeHTTP(w, r)
	}
}

func (rl *Relay) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	for _, reject := range rl.RejectConnection {
		if reject(r) {
			w.WriteHeader(429) // Too many requests
			return
		}
	}
	log.T.Ln("upgrading connection")
	conn, err := rl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v\n", err)
		return
	}
	ticker := time.NewTicker(rl.PingPeriod)
	ws := relayws.New(conn, r)
	rl.clientsMutex.Lock()
	rl.clients[ws] = make([]listenerSpec, 0, 2)
	rl.clientsMutex.Unlock()
	log.T.Ln("added listener", ws.Remote())
	c, cancel := context.Cancel(context.Value(context.Bg(), wsKey, ws))
	kill := func() {
		log.I.Ln("killing listener", ws.Remote())
		for _, ondisconnect := range rl.OnDisconnect {
			ondisconnect(c)
		}
		ticker.Stop()
		cancel()
		conn.Close()
		rl.removeClientAndListeners(ws)
		log.I.Ln("removed listener", ws.Remote())
	}
	go func() {
		defer kill()
		var err E
		log.I.Ln("starting websocket handler")
		conn.SetReadLimit(rl.MaxMessageSize)
		conn.SetReadDeadline(time.Now().Add(rl.PongWait))
		conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(rl.PongWait)) })
		for i, onconnect := range rl.OnConnect {
			log.I.Ln("running onconnect", i)
			onconnect(c)
		}
		for {
			var typ int
			var message B
			if typ, message, err = conn.ReadMessage(); chk.E(err) {
				if W.IsUnexpectedCloseError(
					err,
					W.CloseNormalClosure,    // 1000
					W.CloseGoingAway,        // 1001
					W.CloseNoStatusReceived, // 1005
					W.CloseAbnormalClosure,  // 1006
					4537,                    // some client seems to send many of these
				) {
					log.E.F("unexpected close error from %s: %v\n", ws.Remote(), err)
				}
				return
			}
			if typ == W.PingMessage {
				log.I.Ln("ping")
				ws.Pong()
				continue
			}
			go func(msg []byte) {
				log.T.F("processing message\n%s", msg)
				var err E
				var t S
				var rem B
				if t, rem, err = envelopes.Identify(msg); chk.E(err) {
					return
				}
				log.T.Ln("message type", t)
				switch t {
				case eventenvelope.L:
					env := eventenvelope.NewSubmission()
					if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
						return
					}
					log.I.S(env)
					// check id
					id := env.T.GetIDBytes()
					if !equals(id, env.T.ID) {
						ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID), false,
							normalize.Reason(reasons.Invalid, "id is computed incorrectly")))
						return
					}
					var ok bool
					if ok, err = env.T.Verify(); chk.E(err) {
						ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID), false,
							normalize.Reason(reasons.Error, "failed to verify signature")))
						return
					}
					if !ok {
						ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID), false,
							normalize.Reason(reasons.Invalid, "signature invalid")))
						return
					}
					// check NIP-70 protected
					for _, v := range env.T.Tags.T {
						if len(v.Field) == 1 && equals(v.Field[0], B("-")) {
							message := "must be published by event author"
							authed := GetAuthed(c)
							if len(authed) == 0 {
								RequestAuth(c)
								ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID),
									false, normalize.Reason(reasons.AuthRequired, message)))
								return
							}
							if !equals(authed, env.T.PubKey) {
								ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID),
									false, normalize.Reason(reasons.Blocked, message)))
								return
							}
						}
					}

					srl := rl
					if rl.getSubRelayFromEvent != nil {
						srl = rl.getSubRelayFromEvent(env.T)
					}

					var writeErr error
					var skipBroadcast bool
					if env.T.Kind.ToInt() == kind.Deletion.ToInt() {
						// this always returns "blocked: " whenever it returns an error
						writeErr = srl.handleDeleteRequest(c, env.T)
					} else {
						// this will also always return a prefixed reason
						skipBroadcast, writeErr = srl.AddEvent(c, env.T)
					}

					var reason B
					if writeErr == nil {
						ok = true
						for _, ovw := range srl.OverwriteResponseEvent {
							ovw(c, env.T)
						}
						if !skipBroadcast {
							srl.notifyListeners(env.T)
						}
					} else {
						reason = B(writeErr.Error())
						if bytes.HasPrefix(reason, reasons.AuthRequired) {
							RequestAuth(c)
						}
					}
					ws.WriteEnvelope(okenvelope.NewFrom(eventid.NewWith(env.T.ID), ok, reason))
				case countenvelope.L:
					env := countenvelope.New()
					if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
						return
					}
					if rl.CountEvents == nil {
						ws.WriteEnvelope(closedenvelope.NewFrom(env.ID,
							normalize.Reason(reasons.Unsupported,
								"this relay does not support NIP-45")))
						return
					}

					var total N
					for _, f := range env.Filters.F {
						srl := rl
						if rl.getSubRelayFromFilter != nil {
							srl = rl.getSubRelayFromFilter(f)
						}
						total += srl.handleCountRequest(c, ws, f)
					}
					ws.WriteEnvelope(countenvelope.NewResponseFrom(env.ID, total, false))
				case reqenvelope.L:
					env := reqenvelope.New()
					if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
						return
					}
					log.T.S(env)
					eose := sync.WaitGroup{}
					eose.Add(env.Filters.Len())
					// a context just for the "stored events" request handler
					reqCtx, cancelReqCtx := context.CancelCause(c)
					// expose subscription id in the context
					reqCtx = context.Value(reqCtx, subscriptionIdKey, env.Subscription)
					// handle each filter separately -- dispatching events as they're loaded from databases
					for i, f := range env.Filters.F {
						log.T.F("filter %d:\n%s", i, f)
						srl := rl
						if rl.getSubRelayFromFilter != nil {
							srl = rl.getSubRelayFromFilter(f)
						}
						err := srl.handleRequest(reqCtx, env.Subscription, &eose, ws, f)
						if err != nil {
							// fail everything if any filter is rejected
							reason := B(err.Error())
							if bytes.HasPrefix(reason, B("auth-required")) {
								RequestAuth(c)
							}
							ws.WriteEnvelope(closedenvelope.NewFrom(env.Subscription, reason))
							cancelReqCtx(errors.New("filter rejected"))
							return
						} else {
							rl.addListener(ws, env.Subscription, srl, f, cancelReqCtx)
						}
					}

					go func() {
						// when all events have been loaded from databases and dispatched
						// we can cancel the context and fire the EOSE message
						eose.Wait()
						cancelReqCtx(nil)
						ws.WriteEnvelope(eoseenvelope.NewFrom(env.Subscription))
					}()
				case closeenvelope.L:
					env := eoseenvelope.New()
					if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
						return
					}
					rl.removeListenerId(ws, env.Subscription)
				case authenvelope.L:
					env := authenvelope.NewResponse()
					if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
						return
					}
					wsBaseUrl := strings.Replace(rl.ServiceURL, "http", "ws", 1)
					var ok bool
					if ok, err = auth.Validate(env.Event, ws.Challenge(), wsBaseUrl); ok {
						ws.SetAuthPubKey(env.Event.PubKey)
						ws.Authed.Q()
						// ws.authLock.Lock()
						// if ws.Authed != nil {
						// 	close(ws.Authed)
						// 	ws.Authed = nil
						// }
						// ws.authLock.Unlock()
						ws.WriteEnvelope(okenvelope.NewFrom(
							eventid.NewWith(env.Event.ID), true, B{}))
					} else {
						ws.WriteEnvelope(okenvelope.NewFrom(
							eventid.NewWith(env.Event.ID), false,
							normalize.Reason(reasons.Error, "failed to authenticate"),
						))
					}
				}
			}(message)
		}
	}()

	go func() {
		defer kill()
		for {
			select {
			case <-c.Done():
				return
			case <-ticker.C:
				if err := ws.Ping(); err != nil {
					if !strings.HasSuffix(err.Error(), "use of closed network connection") {
						log.E.F("error writing ping: %v; closing websocket\n", err)
					}
					return
				}
			}
		}
	}()
}
