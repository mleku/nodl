package eventstore

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/relay/eventstore/badger/del"
)

// I is an interface for a persistence layer for nostr events handled by a relay.
type I interface {
	// Init is called at the very beginning by [Server.Start], after
	// [Relay.Init], allowing a storage to initialize its internal resources.
	// The parameters can be used by the database implementations to set custom
	// parameters such as cache management and other relevant parameters to the
	// specific implementation.
	Init() (err E)
	// Close must be called after you're done using the store, to free up
	// resources and so on.
	Close()
	// QueryEvents is invoked upon a client's REQ as described in NIP-01. it
	// should return a channel with the events as they're recovered from a
	// database. the channel should be closed after the events are all
	// delivered.
	QueryEvents(c Ctx, f *filter.T) (ch event.C, err E)
	// CountEvents performs the same work as QueryEvents but instead of
	// delivering the events that were found it just returns the count of events
	CountEvents(c Ctx, f *filter.T) (count int, err E)
	// DeleteEvent is used to handle deletion events, as per NIP-09.
	DeleteEvent(c Ctx, ev EV) (err E)
	// SaveEvent is called once Relay.AcceptEvent reports true.
	SaveEvent(c Ctx, ev EV) (err E)
}

// Cache is a sketch of an expanded enveloper that might be used for a
// size-constrained event store.
type Cache interface {
	I
	GCCount() (deleteItems del.Items, err E)
	Delete(serials del.Items) (err E)
}
