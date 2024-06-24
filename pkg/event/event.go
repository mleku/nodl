package event

import (
	"io"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/ints"
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

var (
	Id        = B("id")
	Pubkey    = B("pubkey")
	CreatedAt = B("created_at")
	Kind      = B("kind")
	Tags      = B("tags")
	Content   = B("content")
	Sig       = B("sig")
)

func (t T) Marshal(dst B) (b B) {
	// open parentheses
	dst = append(dst, '{')
	// ID
	dst = text.JSONKey(dst, Id)
	dst = text.AppendQuote(dst, t.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = text.JSONKey(dst, Pubkey)
	dst = text.AppendQuote(dst, t.PubKey, hex.EncAppend)
	dst = append(dst, ',')
	// CreatedAt
	dst = text.JSONKey(dst, CreatedAt)
	dst = ints.Int64AppendToByteString(dst, t.CreatedAt.I64())
	dst = append(dst, ',')
	// Kind
	dst = text.JSONKey(dst, Kind)
	dst = t.Kind.Marshal(dst)
	dst = append(dst, ',')
	// Tags
	dst = text.JSONKey(dst, Tags)
	dst = t.Tags.Marshal(dst)
	dst = append(dst, ',')
	// Content
	dst = text.JSONKey(dst, Content)
	dst = text.AppendQuote(dst, t.Content, text.NostrEscape)
	dst = append(dst, ',')
	// Sig
	dst = text.JSONKey(dst, Sig)
	dst = text.AppendQuote(dst, t.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

func UnmarshalContent(b B) (content, rem B, err error) {
	rem = b[:]
	for ; len(rem) >= 0; rem = rem[1:] {
		// advance to open quotes
		if rem[0] == '"' {
			rem = rem[1:]
			break
		}
	}
	if len(rem) == 0 {
		err = io.EOF
		return
	}
	var escaping bool
	for len(rem) > 0 {
		if rem[0] == '\\' {
			escaping = true
			content = append(content, rem[0])
			rem = rem[1:]
		} else if rem[0] == '"' {
			if !escaping {
				rem = rem[1:]
				content = text.NostrUnescape(content)
				return
			}
			content = append(content, rem[0])
			rem = rem[1:]
			escaping = false
		} else {
			escaping = false
			content = append(content, rem[0])
			rem = rem[1:]
		}
	}
	return
}

// states of the unmarshaler
const (
	beforeOpen = iota
	openParen
	inKey
	inKV
	inVal
	betweenKV
	afterClose
)

func Unmarshal(b B) (ev *T, rem B, err error) {
	ev = &T{}
	rem = b[:]
	var key B
	var state int
	for ; len(rem) >= 0; rem = rem[1:] {
		switch state {
		case beforeOpen:
			if rem[0] == '{' {
				state = openParen
			}
		case openParen:
			if rem[0] == '"' {
				state = inKey
			}
		case inKey:
			if rem[0] == '"' {
				state = inKV
			} else {
				key = append(key, rem[0])
			}
		case inKV:
			if rem[0] == ':' {
				state = inVal
			}
		case inVal:
			switch key[0] {
			case Id[0]:
				if len(key) < len(Id) {
					goto invalid
				}
				var id B
				if id, rem, err = text.UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(id) != sha256.Size {
					err = errorf.E("invalid ID, require %d got %d", sha256.Size,
						len(id))
					return
				}
				ev.ID = id
				state = betweenKV
			case Pubkey[0]:
				if len(key) < len(Pubkey) {
					goto invalid
				}
				var pk B
				if pk, rem, err = text.UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(pk) != schnorr.PubKeyBytesLen {
					err = errorf.E("invalid pubkey, require %d got %d",
						schnorr.PubKeyBytesLen, len(pk))
					return
				}
				ev.PubKey = pk
				state = betweenKV
			case Kind[0]:
				if len(key) < len(Kind) {
					goto invalid
				}
				if ev.Kind, rem, err = kind.Unmarshal(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}
				if ev.Tags, rem, err = tags.Unmarshal(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Sig[0]:
				if len(key) < len(Sig) {
					goto invalid
				}
				var sig B
				if sig, rem, err = text.UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(sig) != schnorr.SignatureSize {
					err = errorf.E("invalid sig length, require %d got %d '%s'",
						schnorr.SignatureSize, len(sig), rem)
					return
				}
				ev.Sig = sig
				state = betweenKV
			case Content[0]:
				// this can be one of two, but minimum of the shortest
				if len(key) < len(Content) {
					goto invalid
				}
				if key[1] == Content[1] {
					if ev.Content, rem, err = UnmarshalContent(rem); chk.E(err) {
						return
					}
					state = betweenKV
				} else if key[1] == CreatedAt[1] {
					if len(key) < len(CreatedAt) {
						goto invalid
					}
					if ev.CreatedAt, rem, err = timestamp.Unmarshal(rem); chk.E(err) {
						return
					}
					state = betweenKV
				} else {
					goto invalid
				}
			default:
				goto invalid
			}
			key = key[:0]
		case betweenKV:
			if len(rem) == 0 {
				return
			}
			if rem[0] == '}' {
				state = afterClose
				rem = rem[1:]
			} else if rem[0] == ',' {
				state = openParen
			} else if rem[0] == '"' {
				state = inKey
			}
		}
	}
	if len(rem) == 0 && state != afterClose {
		err = errorf.E("invalid event,'%s'", S(b))
	}
	return
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(rem)]),
		S(rem))
	return
}
