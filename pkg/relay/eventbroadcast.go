package relay

import (
	"github.com/mleku/nodl/pkg/codec/envelopes/eventenvelope"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/kinds"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/codec/tag"
	"github.com/mleku/nodl/pkg/protocol/relayws"
)

// BroadcastEvent emits an event to all listeners whose filters' match, skipping
// all filters and actions it also doesn't attempt to store the event or trigger
// any reactions or callbacks
func (rl *R) BroadcastEvent(ev *event.T) {
	listeners.Range(func(ws *relayws.WS, subs ListenerMap) bool {

		if len(ws.AuthPubKey()) == 0 && rl.Info.Limitation.AuthRequired {
			log.E.Ln("cannot broadcast to", ws.RealRemote(), "not authorized")
			return true
		}
		subs.Range(func(id string, listener *Listener) bool {
			if !listener.filters.Match(ev) {
				return true
			}
			if kinds.IsPrivileged(ev.Kind) && rl.Info.Limitation.AuthRequired {
				if len(ws.AuthPubKey()) == 0 {
					log.T.Ln("not broadcasting privileged event to",
						ws.RealRemote(), "not authenticated")
					return true
				}
				parties := tag.T{Field: []B{ev.PubKey}}
				pTags := ev.Tags.GetAll(B("p"))
				for i := range pTags.T {
					parties.Field = append(parties.Field, pTags.T[i].Field[1])
				}
				if !parties.Contains(ws.AuthPubKey()) {
					log.T.Ln("not broadcasting privileged event to",
						ws.RealRemote(), "not party to event")
					return true
				}
			}
			// todo: there may be an issue triggering repeated broadcasts via L2 reviving
			log.D.F("sending event to subscriber %v %s (%d %s)",
				ws.RealRemote(), ws.AuthPubKey(),
				ev.Kind,
				ev.Kind.Name(),
			)
			chk.E(ws.WriteEnvelope(&eventenvelope.Result{
				Subscription: &subscriptionid.T{T: subscriptionid.B(id)},
				Event:        ev},
			))
			return true
		})
		return true
	})
}
