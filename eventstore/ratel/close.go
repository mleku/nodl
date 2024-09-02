package ratel

import (
	. "nostr.mleku.dev"
)

func (r *T) Close() (err E) {
	Log.I.F("closing database %s", r.Path())
	if err = r.DB.Flatten(4); Chk.E(err) {
		return
	}
	Log.D.F("database flattened")
	if err = r.seq.Release(); Chk.E(err) {
		return
	}
	Log.D.F("database released")
	if err = r.DB.Close(); Chk.E(err) {
		return
	}
	Log.I.F("database closed")
	return
}
