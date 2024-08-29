package ratel

import (
	"git.replicatr.dev/eventstore/ratel/keys/index"
	. "nostr.mleku.dev"
)

func (r *T) Nuke() (err E) {
	Log.W.F("nukening database at %s", r.Path)
	if err = r.DB.DropPrefix([][]byte{
		{index.Event.B()},
		{index.CreatedAt.B()},
		{index.Id.B()},
		{index.Kind.B()},
		{index.Pubkey.B()},
		{index.PubkeyKind.B()},
		{index.Tag.B()},
		{index.Tag32.B()},
		{index.TagAddr.B()},
		{index.Counter.B()},
	}...); Chk.E(err) {
		return
	}
	if err = r.DB.RunValueLogGC(0.8); Chk.E(err) {
		return
	}
	return
}
