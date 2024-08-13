package relay

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
)

type Router struct{ *Relay }

type Route struct {
	eventMatcher  func(*event.T) bool
	filterMatcher func(*filter.T) bool
	relay         *Relay
}

type routeBuilder struct {
	router        *Router
	eventMatcher  func(*event.T) bool
	filterMatcher func(*filter.T) bool
}

func NewRouter() *Router {
	rr := &Router{Relay: New()}
	rr.routes = make([]Route, 0, 3)
	rr.getSubRelayFromFilter = func(f *filter.T) *Relay {
		for _, route := range rr.routes {
			if route.filterMatcher(f) {
				return route.relay
			}
		}
		return rr.Relay
	}
	rr.getSubRelayFromEvent = func(e *event.T) *Relay {
		for _, route := range rr.routes {
			if route.eventMatcher(e) {
				return route.relay
			}
		}
		return rr.Relay
	}
	return rr
}

func (rr *Router) Route() routeBuilder {
	return routeBuilder{
		router:        rr,
		filterMatcher: func(f *filter.T) bool { return false },
		eventMatcher:  func(e *event.T) bool { return false },
	}
}

func (rb routeBuilder) Req(fn func(*filter.T) bool) routeBuilder {
	rb.filterMatcher = fn
	return rb
}

func (rb routeBuilder) Event(fn func(*event.T) bool) routeBuilder {
	rb.eventMatcher = fn
	return rb
}

func (rb routeBuilder) Relay(relay *Relay) {
	rb.router.routes = append(rb.router.routes, Route{
		filterMatcher: rb.filterMatcher,
		eventMatcher:  rb.eventMatcher,
		relay:         relay,
	})
}
