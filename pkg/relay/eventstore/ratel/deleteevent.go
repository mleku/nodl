package ratel

import "nostr.mleku.dev/codec/event"

func (r *T) DeleteEvent(c Ctx, ev *event.T) (err E) {
	log.I.F("delete event\n%s", ev)
	return
}
