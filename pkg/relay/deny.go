package relay

func (rl *R) Deny(ws WS) (deny bool) {
	deny = true
	if len(rl.Whitelist) > 0 {
		for i := range rl.Whitelist {
			if rl.Whitelist[i] == ws.Remote() {
				deny = false
			}
		}
	} else {
		deny = false
	}
	return
}
