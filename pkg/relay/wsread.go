package relay

import (
	"time"

	"github.com/fasthttp/websocket"
)

func (rl *R) websocketReadMessages(h *Handle, conn *websocket.Conn, r Req) {
	c, ws, _, _ := h.H()
	if ws.OffenseCount.Load() > IgnoreAfter {
		log.T.Ln("dropping message due to over", IgnoreAfter,
			"errors from this client on this connection",
			ws.RealRemote(), ws.AuthPubKey())
		return
	}
	deny := true
	if len(rl.Whitelist) > 0 {
		for i := range rl.Whitelist {
			if rl.Whitelist[i] == ws.RealRemote() {
				deny = false
			}
		}
	} else {
		deny = false
	}
	if deny {
		// log.T.F("denying access to '%s': dropping message",
		// 	ws.RealRemote())
		// kill()
		return
	}
	conn.SetReadLimit(int64(MaxMessageSize))
	chk.E(conn.SetReadDeadline(time.Now().Add(rl.PongWait)))
	conn.SetPongHandler(func(string) (err E) {
		err = conn.SetReadDeadline(time.Now().Add(rl.PongWait))
		chk.E(err)
		return
	})
	for _, onConnect := range rl.OnConnects {
		onConnect(c)
	}
	for {
		var err E
		var typ int
		var message B
		typ, message, err = conn.ReadMessage()
		if err != nil {
			// log.I.F("%s from %s, %d bytes message", err, ws.RealRemote(), len(message))
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,    // 1000
				websocket.CloseGoingAway,        // 1001
				websocket.CloseNoStatusReceived, // 1005
				websocket.CloseAbnormalClosure,  // 1006
			) {
				log.E.F("unexpected close error from %s: %v",
					ws.RealRemote(), err)
			}
			return
		}
		if typ == websocket.PingMessage {
			chk.E(ws.Pong())
			continue
		}
		log.T.Ln("received message", string(message), ws.RealRemote())
		if err = rl.wsProcessMessages(h, message); err != nil {
		}
	}
}
