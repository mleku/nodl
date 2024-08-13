package main

import (
	"net/http"

	"git.replicatr.dev/pkg/relay"
	"git.replicatr.dev/pkg/util/lol"
	"github.com/fasthttp/websocket"
)

const DefaultListener = "0.0.0.0:3334"

func main() {
	var err E
	lol.SetLogLevel("trace")
	log.T.Ln("trace")
	rl := &relay.T{
		ListenAddress: DefaultListener,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	log.T.F("listening on %s", rl.ListenAddress)
	if err = http.ListenAndServe(rl.ListenAddress, rl); chk.E(err) {
	}
}
