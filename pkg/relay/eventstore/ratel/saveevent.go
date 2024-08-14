package ratel

import "git.replicatr.dev/pkg/codec/event"

func (r *T) SaveEvent(c Ctx, ev *event.T) (err E) {
	log.I.F("saving event\n%s", ev)
	return
}
