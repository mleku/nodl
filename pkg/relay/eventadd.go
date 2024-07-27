package relay

import (
	"errors"

	"git.replicatr.dev/pkg/relay/eventstore"
)

// AddEvent sends an event through then normal add pipeline, as if it was
// received from a websocket.
func (rl *R) AddEvent(c Ctx, ev EV) (err E) {
	if !rl.IsAuthed(c, "add event") {
		return
	}
	if ev == nil {
		err = errors.New(S(Reason(Error, "event is nil")))
		log.E.Ln(err)
		return
	}
	for _, rej := range rl.RejectEvents {
		if reject, msg := rej(c, ev); reject {
			if msg == "" {
				err = errors.New(S(Reason(Blocked, "no reason")))
				log.E.Ln(err)
				return
			} else {
				err = errors.New(S(Reason(Blocked, msg)))
				log.E.Ln(err)
				return
			}
		}
	}
	if !ev.Kind.IsEphemeral() {
		// log.I.Ln("adding event", ev.ToObject().String())
		for _, store := range rl.StoreEvents {
			if saveErr := store(c, ev); saveErr != nil {
				switch {
				case errors.Is(saveErr, eventstore.ErrDupEvent):
					return saveErr
				default:
					err = log.E.Err(S(Reason(Error, saveErr.Error())))
					log.D.Ln(ev.ID, err)
					return
				}
			}
			// log.I.Ln("added event", ev.ID, "for", GetAuthed(c))
		}
		for _, ons := range rl.OnEventSaveds {
			ons(c, ev)
		}
		// log.I.Ln("saved event", ev.ID)
	} else {
		// log.I.Ln("ephemeral event")
		return
	}
	for _, ovw := range rl.OverwriteResponseEvents {
		ovw(c, ev)
	}
	rl.BroadcastEvent(ev)
	return nil
}
