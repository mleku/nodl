package relay

import (
	"fmt"
	"hash/maphash"
	"unsafe"

	"github.com/mleku/nodl/pkg/codec/filter"
	"github.com/mleku/nodl/pkg/codec/filters"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/protocol/relayws"
	"github.com/mleku/nodl/pkg/util/context"
	"github.com/puzpuzpuz/xsync/v2"
)

type Listener struct {
	filters filters.T
	cancel  context.C
	ws      WS
}

type ListenerMap = *xsync.MapOf[S, *Listener]

var listeners = xsync.NewTypedMapOf[WS, ListenerMap](PointerHasher[relayws.WS])

func GetListeningFilters() (respFilters *filters.T) {
	respFilters = filters.Make(listeners.Size() * 2)
	// here we go through all the existing listeners
	listeners.Range(func(_ WS, subs ListenerMap) bool {
		subs.Range(func(_ string, listener *Listener) bool {
			for _, listenerFilter := range listener.filters.F {
				for _, respFilter := range respFilters.F {
					// check if this filter specifically is already added to
					// respFilters
					if filter.Equal(listenerFilter, respFilter) {
						goto next
					}
				}
				// field not yet present on respFilters, add it
				respFilters.F = append(respFilters.F, listenerFilter)
				// continue to the next filter
			next:
				continue
			}
			return true
		})
		return true
	})
	return
}

// SetListener adds a filter to a connection.
func SetListener(id string, ws WS, f filters.T, c context.C) {
	subs, _ := listeners.LoadOrCompute(ws, func() ListenerMap {
		return xsync.NewMapOf[*Listener]()
	})
	subs.Store(id, &Listener{filters: f, cancel: c, ws: ws})
}

// RemoveListenerId removes a specific subscription id from listeners for a
// given ws client and cancel its specific context
func RemoveListenerId(ws WS, id *subscriptionid.T) {
	if subs, ok := listeners.Load(ws); ok {
		if listener, ok := subs.LoadAndDelete(S(id.T)); ok {
			listener.cancel(fmt.Errorf("subscription closed by client"))
		}
		if subs.Size() == 0 {
			listeners.Delete(ws)
		}
	}
}

// RemoveListener removes WebSocket conn from listeners (no need to cancel
// contexts as they are all inherited from the main connection context)
func RemoveListener(ws WS) { listeners.Delete(ws) }

func PointerHasher[V any](_ maphash.Seed, k *V) uint64 {
	return uint64(uintptr(unsafe.Pointer(k)))
}
