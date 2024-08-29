package ratel

import (
	. "nostr.mleku.dev"
)

func (r *T) Close() (err E) {
	Log.T.F("closing database %s", r.Path)
	if err = r.DB.Flatten(4); Chk.E(err) {
		return
	}
	if err = r.DB.Close(); Chk.E(err) {
		return
	}
	if err = r.seq.Release(); Chk.E(err) {
		return
	}
	return
}
