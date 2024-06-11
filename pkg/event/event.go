package event

import (
	"mleku.net/g/nodl/pkg/eventid"
	"mleku.net/g/nodl/pkg/kind"
	"mleku.net/g/nodl/pkg/pubkey"
	"mleku.net/g/nodl/pkg/signature"
	"mleku.net/g/nodl/pkg/tags"
	"mleku.net/g/nodl/pkg/timestamp"
)

// T is the primary data type of the nostr protocol. It is a simple, flexible
// structure with cryptographic authenticity, flexible tags that are used for
// search filters and timestamp.
type T struct {
	// ID is the SHA256 hash of the canonical encoding of the event
	ID *eventid.T `json:"id"`
	// PubKey is the public key of the event creator in *hexadecimal* format
	PubKey *pubkey.T `json:"pubkey"`
	// CreatedAt is the UNIX timestamp of the event according to the event
	// creator (never trust a timestamp!)
	CreatedAt timestamp.T `json:"created_at"`
	// Kind is the nostr protocol code for the type of event. See kind.T
	Kind kind.T `json:"kind"`
	// Tags are a list of tags, which are a list of strings usually structured
	// as a 3 layer scheme indicating specific features of an event.
	Tags tags.T `json:"tags"`
	// Content is an arbitrary string that can contain anything, but usually
	// conforming to a specification relating to the Kind and the Tags.
	Content string `json:"content"`
	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey.
	Sig *signature.T `json:"sig"`
}

func (t *T) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (t *T) UnmarshalJSON(b []byte) error {
	panic("implement me")
}

func (t *T) MarshalBinary() (data []byte, err error) {
	panic("implement me")
}

func (t *T) UnmarshalBinary(data []byte) error {
	panic("implement me")
}
