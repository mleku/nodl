package ratel

import (
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filter"
)

func (r *T) CountEvents(c Ctx, f *filter.T) (count N, err E) {
	var evs []*event.T
	evs, err = r.QueryEvents(c, f)
	count = len(evs)
	return
}
