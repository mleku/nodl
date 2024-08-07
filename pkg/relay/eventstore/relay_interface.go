package eventstore

import (
	"errors"
	"fmt"
	"sort"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/relay/subscriptionoption"
)

// RelayInterface is a wrapper thing that unifies Store and nostr.Relay under a
// common API.
type RelayInterface interface {
	Publish(c Ctx, evt EV) E
	QuerySync(c Ctx, f *filter.T,
		opts ...subscriptionoption.I) ([]EV, E)
}

type RelayWrapper struct {
	Store
}

var _ RelayInterface = (*RelayWrapper)(nil)

func (w RelayWrapper) Publish(c Ctx, evt EV) (err E) {
	// var ch event.C
	// defer close(ch)
	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
		return nil
		// todo: we are no longer deleting old replaceable events because this is racy
		// } else if evt.Kind.IsReplaceable() {
		// replaceable event, delete before storing
		// ch, err = w.Store.QueryEvents(c, &filter.T{
		// 	Authors: []string{evt.PubKey},
		// 	Kinds:   kinds.T{evt.Kind},
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to query before replacing: %w", err)
		// }
		// if previous := <-ch; previous != nil && isOlder(previous, evt) {
		// 	if err = w.Store.DeleteEvent(c, previous); err != nil {
		// 		return fmt.Errorf("failed to delete event for replacing: %w", err)
		// 	}
		// }
		// } else if evt.Kind.IsParameterizedReplaceable() {
		// parameterized replaceable event, delete before storing
		// d := evt.Tags.GetFirst([]string{"d", ""})
		// if d != nil {
		// ch, err = w.Store.QueryEvents(c, &filter.T{
		// 	Authors: []string{evt.PubKey},
		// 	Kinds:   kinds.T{evt.Kind},
		// 	Tags:    filter.TagMap{"d": []string{d.Value()}},
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to query before parameterized replacing: %w", err)
		// }
		// if previous := <-ch; previous != nil && isOlder(previous, evt) {
		// 	if err = w.Store.DeleteEvent(c, previous); chk.D(err) {
		// 		return fmt.Errorf("failed to delete event for parameterized replacing: %w", err)
		// 	}
		// }
		// }
	}
	if err = w.SaveEvent(c, evt); err != nil && !errors.Is(err, ErrDupEvent) {
		return fmt.Errorf("failed to save: %w", err)
	}
	return nil
}

func (w RelayWrapper) QuerySync(c Ctx, f *filter.T,
	opts ...subscriptionoption.I) ([]EV, E) {

	ch, err := w.Store.QueryEvents(c, f)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	n := f.Limit
	if n != 0 {
		results := make(event.Descending, 0, n)
		for evt := range ch {
			results = append(results, evt)
		}
		sort.Sort(results)
		return results, nil
	}
	return nil, nil
}
