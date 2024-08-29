package ratel

import (
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/filter"
)

func (r *T) CountEvents(c Ctx, f *filter.T) (count N, err E) {
	Log.I.F("count events\n%s", f)
	return
}
