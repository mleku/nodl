package types

import (
	"encoding/json"

	"git.replicatr.dev/relay/web"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/filter"
	"nostr.mleku.dev/codec/filters"
	"nostr.mleku.dev/protocol/relayinfo"
	"util.mleku.dev/context"

	"git.replicatr.dev/eventstore"
	"nostr.mleku.dev/codec/event"
)

// Relayer is the main types for implementing a nostr relay.
type Relayer interface {
	// Name is used as the "name" field in NIP-11 and as a prefix in default Server logging.
	// For other NIP-11 fields, see [Informer].
	Name() S
	// Init is called at the very beginning by [Server.Start], allowing a relay
	// to initialize its internal resources.
	// Also see [eventstore.Store.Init].
	Init() E
	// AcceptEvent is called for every nostr event received by the server.
	// If the returned value is true, the event is passed on to [Storage.SaveEvent].
	// Otherwise, the server responds with a negative and "blocked" message as described
	// in NIP-20.
	AcceptEvent(context.T, *event.T) bool
	// Path is the filesystem location of the event store.
	Path() S
	// Storage returns the relay storage implementation.
	Storage() eventstore.I
}

// RequestAcceptor is the main types for implementing a nostr relay.
type RequestAcceptor interface {
	// AcceptReq is called for every nostr request filters received by the
	// server. If the returned value is true, the filtres is passed on to
	// [Storage.QueryEvent].
	AcceptReq(ctx context.T, id S, ff *filters.T, authedPubkey B) bool
}

// Authenticator is the types for implementing NIP-42.
// ServiceURL() returns the URL used to verify the "AUTH" event from clients.
type Authenticator interface {
	ServiceURL() string
}

type Injector interface {
	InjectEvents() event.C
}

// Informer is called to compose NIP-11 response to an HTTP request with
// application/nostr+json mime type. See also [Relayer.Name].
type Informer interface {
	GetNIP11InformationDocument() *relayinfo.T
}

// CustomWebSocketHandler is passed nostr message types unrecognized by the server. The server
// handles "EVENT", "REQ" and "CLOSE" messages, as described in NIP-01.
type CustomWebSocketHandler interface {
	HandleUnknownType(ws *web.Socket, typ string, request []json.RawMessage)
}

// ShutdownAware is called during the server shutdown.
// See [Server.Shutdown] for details.
type ShutdownAware interface {
	OnShutdown(context.T)
}


// AdvancedDeleter methods are called before and after [Storage.DeleteEvent].
type AdvancedDeleter interface {
	BeforeDelete(c context.T, id, pubkey S)
	AfterDelete(id, pubkey S)
}

// AdvancedSaver methods are called before and after [Storage.SaveEvent].
type AdvancedSaver interface {
	BeforeSave(c context.T,ev  *event.T)
	AfterSave(ev *event.T)
}

type EventCounter interface {
	CountEvents(c Ctx, f *filter.T) (count N, err E)
}
