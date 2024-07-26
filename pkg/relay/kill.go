package relay

import (
	"time"

	"github.com/mleku/nodl/pkg/util/context"
)

func (rl *R) Kill(c Ctx, cancel context.F, ws WS, ticker *time.Ticker) func() {
	return func() {
		if len(rl.Whitelist) > 0 {
			for i := range rl.Whitelist {
				if ws.Remote() == rl.Whitelist[i] {
					log.T.Ln("disconnecting whitelisted client from", ws.Remote())
				}
			}
		} else {
			log.T.Ln("disconnecting from", ws.Remote())
		}
		for _, onDisconnect := range rl.OnDisconnects {
			onDisconnect(c)
		}
		ticker.Stop()
		cancel()
		if _, ok := rl.clients.Load(ws.Conn); ok {
			rl.clients.Delete(ws.Conn)
			RemoveListener(ws)
		}
	}
}
