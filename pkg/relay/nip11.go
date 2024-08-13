package relay

import (
	"encoding/json"
	"net/http"
)

func (rl *Relay) HandleNIP11(w http.ResponseWriter, r *http.Request) {
	log.D.Ln("hande relay info")
	w.Header().Set("Content-Type", "application/nostr+json")
	info, _ := rl.Info.Clone()
	if len(rl.DeleteEvent) > 0 {
		info.Nips = append(info.Nips, 9)
	}
	if len(rl.CountEvents) > 0 {
		info.Nips = append(info.Nips, 45)
	}

	for _, ovw := range rl.OverwriteRelayInformation {
		info = ovw(r.Context(), r, info)
	}

	json.NewEncoder(w).Encode(info)
}
