package relay

import (
	"fmt"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/tag"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/normalize"
)

func (rl *Relay) handleDeleteRequest(c context.T, evt *event.T) (err E) {
	// event deletion -- nip09
	for _, t := range evt.Tags.T {
		if len(t.Field) >= 2 && equals(t.Key(), B("e")) {
			// first we fetch the event
			for _, query := range rl.QueryEvents {
				ch, err := query(c, &filter.T{IDs: tag.New(B(t.Field[1]))})
				if err != nil {
					continue
				}
				target := <-ch
				if target == nil {
					continue
				}
				// got the event, now check if the user can delete it
				acceptDeletion := equals(target.PubKey, evt.PubKey)
				var msg string
				if acceptDeletion == false {
					msg = "you are not the author of this event"
				}
				// but if we have a function to overwrite this outcome, use that instead
				for _, odo := range rl.OverwriteDeletionOutcome {
					acceptDeletion, msg = odo(c, target, evt)
				}
				if acceptDeletion {
					// delete it
					for _, del := range rl.DeleteEvent {
						if err := del(c, target); err != nil {
							return err
						}
					}
				} else {
					// fail and stop here
					return fmt.Errorf(S(normalize.Reason(reasons.Blocked, msg)))
				}

				// don't try to query this same event again
				break
			}
		}
	}

	return
}
