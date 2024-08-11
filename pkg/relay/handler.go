package relay

import (
	"strings"
	"time"

	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/qu"
	"github.com/fasthttp/websocket"
)

type Handle struct {
	rl     *R
	c      Ctx
	ws     WS
	svcURL S
	kill   func()
}

func H(rl *R, c Ctx, ws WS, svcURL S, kill func()) (h *Handle) {
	return &Handle{rl, c, ws, svcURL, kill}
}
func (h *Handle) H() (rl *R, c Ctx, ws WS, svcURL S, kill func()) {
	return h.rl, h.c, h.ws, h.svcURL, h.kill
}

// HandleWebsocket is a http handler that accepts and manages websocket connections.
func (rl *R) HandleWebsocket(svcURL S) func(w Responder, r Req) {
	return func(w Responder, r Req) {
		var err E
		var conn *websocket.Conn
		if conn, err = rl.upgrader.Upgrade(w, r, nil); chk.E(err) {
			log.E.F("failed to upgrade websocket: %v", err)
			return
		}
		conn.SetReadLimit(int64(MaxMessageSize))
		conn.EnableWriteCompression(true)
		rl.clients.Store(conn, struct{}{})
		ticker := time.NewTicker(rl.PingPeriod)
		rem := r.Header.Get("X-Forwarded-For")
		split := strings.Split(rem, " ")
		var rr S
		switch len(split) {
		case 1:
			rr = split[0]
		case 2:
			rr = split[1]
		}
		// in case upstream doesn't set this or we are directly listening instead of via reverse proxy or just if the
		// header field is missing, put the connection remote address into the websocket state data.
		if rr == "" {
			rr = r.RemoteAddr
		}
		ws := relayws.New(conn, r, qu.T())
		log.I.F("%s", conn.RemoteAddr().String())
		ws.SetRealRemote(rr)
		// check if whitelist denies this connection
		var deny bool
		if deny = rl.Deny(ws); deny {
			// log.T.F("denying access to '%s': dropping message", ws.Remote())
			return
		}
		// NIP-42 challenge
		ws.GenerateChallenge()
		c, cancel := context.Cancel(context.Value(context.Bg(), wsKey, ws))
		if len(rl.Whitelist) > 0 {
			if !deny {
				log.T.Ln("whitelisted inbound connection from", rr)
			}
		} else {
			log.T.Ln("inbound connection from", rr)
		}
		kill := rl.Kill(c, cancel, ws, ticker)
		h := H(rl, c, ws, svcURL, kill)
		go chk.E(h.websocketReadMessages())
		go h.websocketWatcher(ticker)
	}
}
