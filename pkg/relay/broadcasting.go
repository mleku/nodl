package relay

import "git.replicatr.dev/pkg/codec/event"

// BroadcastEvent emits an event to all listeners whose filters' match, skipping all filters and actions
// it also doesn't attempt to store the event or trigger any reactions or callbacks
func (rl *Relay) BroadcastEvent(evt *event.T) {
	rl.notifyListeners(evt)
}
