package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	sid "git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/codec/tag"
)

// BroadcastEvent emits an event to all listeners whose filters' match, skipping
// all filters and actions it also doesn't attempt to store the event or trigger
// any reactions or callbacks
func (rl *R) BroadcastEvent(ev EV) {
	listeners.Range(func(ws WS, subs ListenerMap) bool {

		if len(ws.AuthPub()) == 0 && rl.Info.Limitation.AuthRequired {
			log.E.F("cannot broadcast to %s not authorized", ws.Remote())
			return true
		}
		subs.Range(func(id string, listener *Listener) bool {
			if !listener.filters.Match(ev) {
				return true
			}
			if ev.Kind.IsPrivileged() && rl.Info.Limitation.AuthRequired {
				if len(ws.AuthPub()) == 0 {
					log.T.F("not broadcasting privileged event to %s not authenticated", ws.Remote())
					return true
				}
				parties := tag.New(ev.PubKey)
				pTags := ev.Tags.GetAll(tag.New("p"))
				for i := range pTags.T {
					parties.Field = append(parties.Field, pTags.T[i].Field[1])
				}
				if !parties.Contains(ws.AuthPub()) {
					log.T.F("not broadcasting privileged event to %s not party to event", ws.Remote())
					return true
				}
			}
			// todo: there may be an issue triggering repeated broadcasts via L2 reviving
			log.D.F("sending event to subscriber %v %s (%d %s)",
				ws.Remote(), ws.AuthPub(), ev.Kind, ev.Kind.Name())
			var err E
			var si *sid.T
			if si, err = sid.New(id); chk.E(err) {
				return true
			}
			chk.E(eventenvelope.NewResultWith(si, ev).Write(ws))
			return true
		})
		return true
	})
}
