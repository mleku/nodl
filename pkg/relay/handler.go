package relay

import (
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
	"github.com/mleku/nodl/pkg/util/qu"
)

type Handle struct {
	c      Ctx
	ws     WS
	svcURL S
	kill   func()
}

func H(c Ctx, ws WS, svcURL S, kill func()) (h *Handle)    { return &Handle{c, ws, svcURL, kill} }
func (h *Handle) H() (c Ctx, ws WS, svcURL S, kill func()) { return h.c, h.ws, h.svcURL, h.kill }

// HandleWebsocket is a http handler that accepts and manages websocket
// connections.
func (rl *R) HandleWebsocket(serviceURL S) func(w Responder, r Req) {
	return func(w Responder, r Req) {
		var err E
		var conn *websocket.Conn
		conn, err = rl.upgrader.Upgrade(w, r, nil)
		if chk.E(err) {
			log.E.F("failed to upgrade websocket: %v", err)
			return
		}
		conn.SetReadLimit(int64(MaxMessageSize))
		conn.EnableWriteCompression(true)
		rl.clients.Store(conn, struct{}{})
		ticker := time.NewTicker(rl.PingPeriod)
		rem := r.Header.Get("X-Forwarded-For")
		splitted := strings.Split(rem, " ")
		var rr string
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
		// in case upstream doesn't set this or we are directly listening instead of
		// via reverse proxy or just if the header field is missing, put the
		// connection remote address into the websocket state data.
		if rr == "" {
			rr = r.RemoteAddr
		}
		ws := relayws.New(conn, r, qu.T())
		ws.SetRealRemote(rr)
		// NIP-42 challenge
		ws.GenerateChallenge()
		c, cancel := context.Cancel(context.Value(context.Bg(), wsKey, ws))
		if len(rl.Whitelist) > 0 {
			for i := range rl.Whitelist {
				if rr == rl.Whitelist[i] {
					log.T.Ln("whitelisted inbound connection from", rr)
				}
			}
		} else {
			log.T.Ln("inbound connection from", rr)
		}
		kill := func() {
			if len(rl.Whitelist) > 0 {
				for i := range rl.Whitelist {
					if rr == rl.Whitelist[i] {
						log.T.Ln("disconnecting whitelisted client from", rr)
					}
				}
			} else {
				log.T.Ln("disconnecting from", rr)
			}
			for _, onDisconnect := range rl.OnDisconnects {
				onDisconnect(c)
			}
			ticker.Stop()
			cancel()
			if _, ok := rl.clients.Load(conn); ok {
				rl.clients.Delete(conn)
				RemoveListener(ws)
			}
		}
		go rl.websocketReadMessages(H(c, ws, serviceURL, kill), conn, r)
		go rl.websocketWatcher(H(c, ws, "", kill), ticker)
	}
}
