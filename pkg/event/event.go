package event

import (
	"reflect"

	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
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
	CreatedAt timestamp.T `json:"created_at"`
	// Kind is the nostr protocol code for the type of event. See kind.T
	Kind kind.T `json:"kind"`
	// Tags are a list of tags, which are a list of strings usually structured
	// as a 3 layer scheme indicating specific features of an event.
	Tags tags.T `json:"tags"`
	// Content is an arbitrary string that can contain anything, but usually
	// conforming to a specification relating to the Kind and the Tags.
	Content B `json:"content"`
	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey.
	Sig B `json:"sig"`
}

var fields []B

func init() {
	v := reflect.ValueOf(T{})
	for i := 0; i < v.Type().NumField(); i++ {
		fields = append(fields, B(v.Type().Field(i).Tag.Get("json")))
	}
}

func (t T) Marshal(dst B) (b B) {
	// open parentheses
	dst = append(dst, '{')
	// ID
	dst = append(dst, '"')
	dst = append(dst, fields[0]...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = append(dst, ',', '"')
	dst = append(dst, fields[1]...)
	dst = append(dst, '"', ':')
	dst = append(dst, '"')
	dst = text.AppendQuote(dst, t.PubKey, hex.EncAppend)
	dst = append(dst, '"', ',')
	// CreatedAt
	dst = append(dst, ',', '"')
	dst = append(dst, fields[2]...)
	dst = append(dst, '"', ':')
	dst = t.CreatedAt.Marshal(dst)
	dst = append(dst, ',')
	// Kind
	dst = append(dst, ',', '"')
	dst = append(dst, fields[3]...)
	dst = append(dst, '"', ':')
	dst = t.Kind.Marshal(dst)
	dst = append(dst, ',')
	// Tags
	dst = append(dst, ',', '"')
	dst = append(dst, fields[4]...)
	dst = append(dst, '"', ':')
	dst = t.Tags.Marshal(dst)
	dst = append(dst, ',')
	// Content
	dst = append(dst, ',', '"')
	dst = append(dst, fields[5]...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.Content, text.NostrEscape)
	dst = append(dst, ',')
	// Sig
	dst = append(dst, ',', '"')
	dst = append(dst, fields[6]...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
	return
}
func Unmarshal(b B) (t T, rem B, err error) {
	rem = b[:]

	return
}
