package ratel

import "nostr.mleku.dev/codec/filter"

func (r *T) CountEvents(c Ctx, f *filter.T) (count N, err E) {
	log.I.F("count events\n%s", f)
	return
}
