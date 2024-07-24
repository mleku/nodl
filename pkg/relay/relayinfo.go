package relay

import (
	"encoding/json"
)

// HandleNIP11 is a http handler for NIP-11 relayinfo.T requests
func (rl *R) HandleNIP11(w Responder, r Req) {
	var err E
	log.T.Ln("NIP-11 request", getServiceBaseURL(r))
	w.Header().Set("Content-Type", "application/nostr+json")
	info := rl.Info
	for _, ovw := range rl.OverwriteRelayInfos {
		info = ovw(r.Context(), r, info)
	}
	var b B
	if b, err = json.Marshal(info); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	chk.E(err)
}
