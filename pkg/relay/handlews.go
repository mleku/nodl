package relay

import (
	"net/http"

	"git.replicatr.dev/pkg/protocol/relayws"
	"github.com/fasthttp/websocket"
)

func (rl *T) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	log.T.F("HandleWebsocket")
	var err E
	var conn *websocket.Conn
	conn, err = rl.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	ws := relayws.New(conn, r)
	log.T.F("established websocket connection with %s", ws.Remote())
}
