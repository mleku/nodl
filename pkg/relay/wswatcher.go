package relay

import (
	"strings"
	"time"
)

func (rl *R) websocketWatcher(h *Handle, t *time.Ticker) {
	ctx, ws, _, kill := h.H()
	var err E
	for {
		select {
		case <-rl.Ctx.Done():
			return
		case <-ctx.Done():
			return
		case <-t.C:
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
				return
			}
			if err = ws.Ping(); err != nil {
				if !strings.HasSuffix(err.Error(),
					"use of closed network connection") {
					// log.T.F("error writing ping: %v; closing websocket", err)
				}
				kill()
				return
			}
		}
	}
}
