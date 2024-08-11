package relay

import (
	"sync"

	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/codec/tag"
	"git.replicatr.dev/pkg/util/context"
)

type handleFilterParams struct {
	c    context.T
	id   *subscriptionid.T
	eose *sync.WaitGroup
	ws   WS
	f    *filter.T
}

func (rl *R) handleFilter(h handleFilterParams) (err error) {
	if !rl.IsAuthed(h.c, reqenvelope.L) {
		return
	}
	// overwrite the filter (for example, to eliminate some kinds or that we
	// know we don't support, or it could clamp the Limit).
	for _, ovw := range rl.OverwriteReqFilter {
		ovw(h.c, h.f)
	}
	if h.f.Limit < 1 {
		h.f.Limit = 50
		// err = errors.New(S(Reason(Invalid, "filter with 0/empty limit ")))
		// log.E.Ln(err)
		// return
	}
	// then check if we'll reject this filter (we apply this after overwriting
	// because we may, for example, remove some things from the incoming filters
	// that we know we don't support, and then if the end result is an empty
	// filter we can just reject it)
	for _, reject := range rl.RejectReqFilters {
		if rej, msg := reject(h.c, h.id, h.f); rej {
			return log.D.Err("%s %s", Reason(Blocked, S(msg)),
				h.ws.AuthPub())
		}
	}
	// run the functions to query events (generally just one, but we might be
	// fetching stuff from multiple places)
	for _, query := range rl.QueryEvents {
		h.eose.Add(1)
		var ch event.C
		// start up event receiver before running query on this channel
		var kindStrings []string
		if h.f.Kinds != nil && h.f.Kinds.Len() > 0 {
			for _, ks := range h.f.Kinds.K {
				kindStrings = append(kindStrings, kind.GetString(ks))
			}
		}
		if ch, err = query(h.c, h.f); chk.E(err) {
			h.ws.OffenseCount.Inc()
			chk.E(NewNotice(B(err.Error())).Write(h.ws))
			continue
		}
		// todo: eliminating goroutines that don't make sense - client on one socket, one subscription and probably
		//  one thread waiting for the response.
		//
		// go func(ch event.C) {
	out:
		for {
			select {
			case ev := <-ch:
				// if the event is nil the rest of this loop will panic
				// accessing the nonexistent event's fields
				if ev == nil {
					// log.T.Ln("query result channel closed")
					break out
				}
				var evStr B
				evStr, err = ev.MarshalJSON(B{})
				log.T.F("received result\n%s", evStr)
				for _, ovw := range rl.OverwriteResponseEvents {
					ovw(h.c, ev)
				}
				if ev.Kind.IsPrivileged() && rl.Info.Limitation.AuthRequired {
					var allow bool
					for _, v := range rl.Config.AllowIPs {
						if h.ws.Remote() == v {
							allow = true
							break
						}
					}
					if h.ws.HasAuth() && !allow {
						log.D.F("not broadcasting privileged event to %s not authenticated", h.ws.Remote())
						continue
					}
					if !allow {
						// check the filter first
						receivers1 := h.f.Tags.GetAll(tag.New("p"))
						receivers2 := h.f.Tags.GetAll(tag.New("#p"))
						receivers := append(receivers1.T, receivers2.T...)
						parties := tag.NewWithCap(len(receivers) + 1)
						for i := range receivers {
							tt := receivers[i].Value()
							if len(tt) > 0 {
								parties.Field = append(parties.Field, tt)
							}
						}
						parties.Field = append(parties.Field, ev.PubKey)
						pTags := ev.Tags.GetAll(tag.New("p"))
						for i := range pTags.T {
							parties.Field = append(parties.Field, pTags.T[i].Value())
						}
						if !parties.Contains(h.ws.AuthPub()) && rl.Info.Limitation.AuthRequired {
							log.D.F("not broadcasting privileged event to %s not party to event %s",
								h.ws.Remote(), h.ws.AuthPub())
							continue
						}
					}
				}
				chk.E(NewResult(h.id, ev).Write(h.ws))
			case <-rl.Ctx.Done():
				// log.T.Ln("shutting down")
				break out
			case <-h.c.Done():
				// log.T.Ln("query context done")
				break out
			}
		}
		// }(ch)
	}
	return nil
}
