package ratel

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
)

func (r *T) QueryEvents(c Ctx, f *filter.T) (ch event.C, err E) {
	log.I.F("query for events\n%s", f)
	return
}
