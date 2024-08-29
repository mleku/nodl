package ratel

import (
	. "nostr.mleku.dev"

	"nostr.mleku.dev/codec/event"
)

func (r *T) DeleteEvent(c Ctx, ev *event.T) (err E) {
	Log.I.F("delete event\n%s", ev)
	return
}
