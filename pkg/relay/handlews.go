package relay

import (
	"net/http"

	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"github.com/fasthttp/websocket"
)

func (rl *T) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	log.T.F("HandleWebsocket inbound connection from %s", r.RemoteAddr)
	var err E
	var conn *websocket.Conn
	conn, err = rl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	c, cancel := context.Cancel(context.Bg())
	ws := relayws.New(c, conn, r, MaxMessageSize)
	log.T.F("established websocket connection with %s", ws.Remote())
	rl.clients.Store(ws, struct{}{})
	rl.wsWatcher(ws, cancel)
}

func (rl *T) wsWatcher(ws *relayws.WS, cancel context.F) {
	for {
		select {
		case <-rl.Ctx.Done():
			cancel()
			return
		case <-ws.Ctx.Done():
			return
		}
	}
}
