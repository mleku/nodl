package relay

import (
	"encoding/json"
	"net/http"

	"git.replicatr.dev/pkg/protocol/relayinfo"
)

func (rl *T) HandleRelayInfo(w http.ResponseWriter, r *http.Request) {
	var err E
	log.T.Ln("HandleRelayInfo")
	info := relayinfo.NewInfo(&relayinfo.T{})
	w.Header().Set("Content-Type", "application/nostr+json")
	var b []byte
	if b, err = json.Marshal(info); chk.E(err) {
		return
	}
	log.I.F("%s", b)
	if _, err = w.Write(b); chk.E(err) {
		return
	}
}
