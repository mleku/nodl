package relay

import (
	"context"
	"errors"
	"fmt"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/kinds"
	"git.replicatr.dev/pkg/codec/tag"
	"git.replicatr.dev/pkg/codec/tags"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/util/normalize"
)

// AddEvent sends an event through then normal add pipeline, as if it was received from a websocket.
func (rl *Relay) AddEvent(ctx context.Context, evt *event.T) (skipBroadcast bool, err E) {
	if evt == nil {
		return false, errorf.E("error: event is nil")
	}
	for i, reject := range rl.RejectEvent {
		log.I.Ln("rejector", i)
		if reject, msg := reject(ctx, evt); reject {
			if msg == "" {
				return false, errors.New("blocked: no reason")
			} else {
				return false, errors.New(S(normalize.Reason(reasons.Blocked, msg)))
			}
		}
	}
	if 20000 <= evt.Kind.ToInt() && evt.Kind.ToInt() < 30000 {
		log.I.Ln("ephemeral event", evt.Kind.Name())
		// do not store ephemeral events
		for _, oee := range rl.OnEphemeralEvent {
			oee(ctx, evt)
		}
	} else {
		// will store

		// but first check if we already have it
		f := filter.T{IDs: tag.New(evt.ID)}
		for _, query := range rl.QueryEvents {
			ch, err := query(ctx, &f)
			if err != nil {
				continue
			}
			for range ch {
				// if we run this it means we already have this event, so we just return a success and exit
				return true, nil
			}
		}

		// if it's replaceable we first delete old versions
		if evt.Kind.ToInt() == 0 || evt.Kind.ToInt() == 3 || (10000 <= evt.Kind.ToInt() && evt.Kind.ToInt() < 20000) {
			// replaceable event, delete before storing
			f := filter.T{Authors: tag.New(evt.PubKey), Kinds: kinds.New(evt.Kind)}
			for _, query := range rl.QueryEvents {
				ch, err := query(ctx, &f)
				if err != nil {
					continue
				}
				for previous := range ch {
					if isOlder(previous, evt) {
						for _, del := range rl.DeleteEvent {
							del(ctx, previous)
						}
					}
				}
			}
		} else if 30000 <= evt.Kind.ToInt() && evt.Kind.ToInt() < 40000 {
			// parameterized replaceable event, delete before storing
			d := evt.Tags.GetFirst(tag.New("d", ""))
			if d == nil {
				return false, fmt.Errorf("invalid: missing 'd' tag on parameterized replaceable event")
			}

			f := filter.T{Authors: tag.New(evt.PubKey), Kinds: kinds.New(evt.Kind),
				Tags: tags.New(d)}
			for _, query := range rl.QueryEvents {
				ch, err := query(ctx, &f)
				if err != nil {
					continue
				}
				for previous := range ch {
					if isOlder(previous, evt) {
						for _, del := range rl.DeleteEvent {
							del(ctx, previous)
						}
					}
				}
			}
		}

		// store
		for _, store := range rl.StoreEvent {
			if saveErr := store(ctx, evt); saveErr != nil {
				switch saveErr {
				case eventstore.ErrDupEvent:
					return true, nil
				default:
					return false, fmt.Errorf(S(normalize.Reason(reasons.Error,
						saveErr.Error())))
				}
			}
		}

		for _, ons := range rl.OnEventSaved {
			ons(ctx, evt)
		}
	}

	return false, nil
}
