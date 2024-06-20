package event

import (
	"io"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
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
	dst = append(dst, '"')
	dst = append(dst, Id...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = append(dst, ',', '"')
	dst = append(dst, Pubkey...)
	dst = append(dst, '"', ':')
	dst = append(dst, '"')
	dst = text.AppendQuote(dst, t.PubKey, hex.EncAppend)
	dst = append(dst, '"', ',')
	// CreatedAt
	dst = append(dst, ',', '"')
	dst = append(dst, CreatedAt...)
	dst = append(dst, '"', ':')
	dst = t.CreatedAt.FromVarint(dst)
	dst = append(dst, ',')
	// Kind
	dst = append(dst, ',', '"')
	dst = append(dst, Kind...)
	dst = append(dst, '"', ':')
	dst = t.Kind.Marshal(dst)
	dst = append(dst, ',')
	// Tags
	dst = append(dst, ',', '"')
	dst = append(dst, Tags...)
	dst = append(dst, '"', ':')
	dst = t.Tags.Marshal(dst)
	dst = append(dst, ',')
	// Content
	dst = append(dst, ',', '"')
	dst = append(dst, Content...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.Content, text.NostrEscape)
	dst = append(dst, ',')
	// Sig
	dst = append(dst, ',', '"')
	dst = append(dst, Sig...)
	dst = append(dst, '"', ':')
	dst = text.AppendQuote(dst, t.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
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

func UnmarshalHex(b B) (t B, rem B, err error) {
	rem = b[:]
	var inQuote bool
	var start int
	for i := 0; i < len(b); i++ {
		if !inQuote {
			if b[i] == '"' {
				inQuote = true
				start = i + 1
			}
		} else {
			if b[i] == '"' {
				t = b[start:i]
				rem = b[i+1:]
				break
			}
		}
	}
	if !inQuote {
		err = io.EOF
		return
	}
	l := len(t)
	if l%2 != 0 {
		err = errorf.E("invalid length for hex: %d, %0x", len(t), t)
		return
	}
	if _, err = hex.DecBytes(t, t); chk.E(err) {
		return
	}
	t = t[:l/2]
	return
}

func Unmarshal(b B) (t T, rem B, err error) {
	rem = b[:]
	var key B
	var state int
	for ; len(rem) > 0; rem = rem[1:] {
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
				if id, rem, err = UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(id) != sha256.Size {
					err = errorf.E("invalid ID, require %d got %d", sha256.Size,
						len(id))
					return
				}
				t.ID = id
				state = betweenKV
			case Pubkey[0]:
				if len(key) < len(Pubkey) {
					goto invalid
				}
				var pk B
				if pk, rem, err = UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(pk) != schnorr.PubKeyBytesLen {
					err = errorf.E("invalid pubkey, require %d got %d",
						schnorr.PubKeyBytesLen, len(pk))
					return
				}
				t.PubKey = pk
				state = betweenKV
			case Kind[0]:
				if len(key) < len(Kind) {
					goto invalid
				}
				if t.Kind, rem, err = kind.Unmarshal(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}
				if t.Tags, rem, err = tags.Unmarshal(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Sig[0]:
				if len(key) < len(Sig) {
					goto invalid
				}
				var sig B
				if sig, rem, err = UnmarshalHex(rem); chk.E(err) {
					return
				}
				if len(sig) != schnorr.SignatureSize {
					err = errorf.E("invalid sig, require %d got %d",
						schnorr.SignatureSize, len(sig))
					return
				}
				t.Sig = sig
				state = betweenKV
			case Content[0]:
				// this can be one of two, but minimum of the shortest
				if len(key) < len(Content) {
					goto invalid
				}
				if key[1] == Content[1] {

				} else if key[1] == CreatedAt[1] {
					if len(key) < len(CreatedAt) {
						goto invalid
					}
					if t.CreatedAt, rem, err = timestamp.Unmarshal(rem); chk.E(err) {
						return
					}
				} else {
					goto invalid
				}
			default:
				goto invalid
			}
			key = key[:0]
		case betweenKV:
			if rem[0] == '}' {
				rem = rem[1:]
				state = afterClose
			} else if rem[0] == ',' {
				state = openParen
				rem = rem[1:]
				continue
			}
		}
	}
	if len(rem) == 0 && state != afterClose {
		err = errorf.E("invalid event,'%s'", S(b))
	}
	return
invalid:
	err = errorf.E("invalid key, rem: '%s'", S(rem))
	return
}
