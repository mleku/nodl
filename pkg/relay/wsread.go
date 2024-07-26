package relay

import (
	"time"

	w "github.com/fasthttp/websocket"
)

func (h *Handle) websocketReadMessages() (err E) {
	rl, c, ws, _, _ := h.H()
	ws.Conn.SetReadLimit(rl.MaxMessageSize)
	chk.E(ws.Conn.SetReadDeadline(time.Now().Add(rl.PongWait)))
	ws.Conn.SetPongHandler(func(S) (err E) { return ws.Conn.SetReadDeadline(time.Now().Add(rl.PongWait)) })
	for _, onConnect := range rl.OnConnects {
		onConnect(c)
	}
	for {
		var typ int
		var message B
		if typ, message, err = ws.Conn.ReadMessage(); err != nil {
			// log.I.F("%s from %s, %d bytes message", err, ws.Remote(), len(message))
			if w.IsUnexpectedCloseError(
				err,
				w.CloseNormalClosure,    // 1000
				w.CloseGoingAway,        // 1001
				w.CloseNoStatusReceived, // 1005
				w.CloseAbnormalClosure,  // 1006
			) {
				log.E.F("unexpected close error from %s: %v", ws.Remote(), err)
			}
			return
		}
		if typ == w.PingMessage {
			chk.E(ws.Pong())
			continue
		}
		log.T.Ln("received message", S(message), ws.Remote())
		if err = h.wsProcessMessages(message); err != nil {
			return
		}
	}
}
