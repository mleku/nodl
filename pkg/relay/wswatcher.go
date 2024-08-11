package relay

import (
	"strings"
	"time"
)

func (h *Handle) websocketWatcher(t *time.Ticker) {
	rl, ctx, ws, _, kill := h.H()
	var err E
	for {
		select {
		case <-rl.Ctx.Done():
			return
		case <-ctx.Done():
			return
		case <-t.C:
			if err = ws.Ping(); chk.E(err) {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					log.T.F("error writing ping: %v; closing websocket", err)
				}
				kill()
				return
			}
		}
	}
}
