package event

import (
	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg"
	"github.com/mleku/nodl/pkg/codec/kind"
	"github.com/mleku/nodl/pkg/codec/tags"
	"github.com/mleku/nodl/pkg/codec/text"
	"github.com/mleku/nodl/pkg/codec/timestamp"
	"github.com/mleku/nodl/pkg/util/hex"
	"lukechampine.com/frand"
)

// T is the primary datatype of nostr. This is the form of the structure that
// defines its JSON string based format.
type T struct {
	// ID is the SHA256 hash of the canonical encoding of the event
	ID B `json:"id"`
	// PubKey is the public key of the event creator
	PubKey B `json:"pubkey"`
	// CreatedAt is the UNIX timestamp of the event according to the event
	// creator (never trust a timestamp!)
	CreatedAt *timestamp.T `json:"created_at"`
	// Kind is the nostr protocol code for the type of event. See kind.T
	Kind *kind.T `json:"kind"`
	// Tags are a list of tags, which are a list of strings usually structured
	// as a 3 layer scheme indicating specific features of an event.
	Tags *tags.T `json:"tags"`
	// Content is an arbitrary string that can contain anything, but usually
	// conforming to a specification relating to the Kind and the Tags.
	Content B `json:"content"`
	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey.
	Sig B `json:"sig"`
}

func New() (ev *T) { return &T{} }

type C chan *T

func (ev *T) Serialize() (b B) {
	b, _ = ev.MarshalJSON(nil)
	return
}

func (ev *T) ToCanonical() (b B) {
	b = append(b, "[0,\""...)
	b = hex.EncAppend(b, ev.PubKey)
	b = append(b, "\","...)
	var err error
	if b, err = ev.CreatedAt.MarshalJSON(b); chk.E(err) {
		return
	}
	b = append(b, ',')
	if b, err = ev.Kind.MarshalJSON(b); chk.E(err) {
		return
	}
	b = append(b, ',')
	if b, err = ev.Tags.MarshalJSON(b); chk.E(err) {
		panic(err)
	}
	b = append(b, ',')
	b = text.AppendQuote(b, ev.Content, text.NostrEscape)
	b = append(b, ']')
	return
}

func Hash(in []byte) (out []byte) {
	h := sha256.Sum256(in)
	return h[:]
}

// GetIDBytes returns the raw SHA256 hash of the canonical form of an T.
func (ev *T) GetIDBytes() []byte { return Hash(ev.ToCanonical()) }

func GenerateRandomTextNoteEvent(signer pkg.Signer, maxSize int) (ev *T,
	err error) {

	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &T{
		PubKey:    signer.Pub(),
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   text.NostrEscape(nil, frand.Bytes(l)),
	}
	if err = ev.Sign(signer); chk.E(err) {
		return
	}
	return
}