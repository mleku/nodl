package web

import (
	"sync"

	. "nostr.mleku.dev"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
)

type Socket struct {
	*websocket.Conn
	mutex sync.Mutex

	// nip42
	Challenge B
	Authed  B
	Limiter *rate.Limiter
}

func (ws *Socket) Write(p B) (n N, err E) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	err = ws.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		n = len(p)
	}
	return
}
