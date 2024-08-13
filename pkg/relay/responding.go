package relay

import (
	"errors"
	"sync"

	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/noticeenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/normalize"
)

func (rl *Relay) handleRequest(c context.T, id *subscriptionid.T, eose *sync.WaitGroup,
	ws *relayws.WS, f *filter.T) error {
	defer eose.Done()

	// overwrite the filter (for example, to eliminate some kinds or
	// that we know we don't support)
	for _, ovw := range rl.OverwriteFilter {
		ovw(c, f)
	}

	if f.Limit == 0 {
		// don't do any queries, just subscribe to future events
		return nil
	}

	// then check if we'll reject this filter (we apply this after overwriting
	// because we may, for example, remove some things from the incoming filters
	// that we know we don't support, and then if the end result is an empty
	// filter we can just reject it)
	for _, reject := range rl.RejectFilter {
		if reject, msg := reject(c, f); reject {
			return errors.New(S(normalize.Reason(reasons.Blocked, msg)))
		}
	}

	// run the functions to query events (generally just one,
	// but we might be fetching stuff from multiple places)
	eose.Add(len(rl.QueryEvents))
	for _, query := range rl.QueryEvents {
		ch, err := query(c, f)
		if err != nil {
			ws.WriteEnvelope(noticeenvelope.NewFrom(err.Error()))
			eose.Done()
			continue
		} else if ch == nil {
			eose.Done()
			continue
		}

		go func(ch chan *event.T) {
			for event := range ch {
				for _, ovw := range rl.OverwriteResponseEvent {
					ovw(c, event)
				}
				ws.WriteEnvelope(eventenvelope.NewResultWith(id, event))
			}
			eose.Done()
		}(ch)
	}

	return nil
}

func (rl *Relay) handleCountRequest(c context.T, ws *relayws.WS, f *filter.T) N {
	// overwrite the filter (for example, to eliminate some kinds or tags that we know we don't support)
	for _, ovw := range rl.OverwriteCountFilter {
		ovw(c, f)
	}

	// then check if we'll reject this filter
	for _, reject := range rl.RejectCountFilter {
		if rejecting, msg := reject(c, f); rejecting {
			ws.WriteEnvelope(noticeenvelope.NewFrom(msg))
			return 0
		}
	}

	// run the functions to count (generally it will be just one)
	var subtotal N = 0
	for _, count := range rl.CountEvents {
		res, err := count(c, f)
		if err != nil {
			ws.WriteEnvelope(noticeenvelope.NewFrom(err.Error()))
		}
		subtotal += res
	}

	return subtotal
}
