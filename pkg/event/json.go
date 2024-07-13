package event

import (
	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
)

var (
	Id        = B("id")
	Pubkey    = B("pubkey")
	CreatedAt = B("created_at")
	Kind      = B("kind")
	Tags      = B("tags")
	Content   = B("content")
	Sig       = B("sig")
)

func (ev *T) MarshalJSON(dst B) (b B, err error) {
	// open parentheses
	dst = append(dst, '{')
	// ID
	dst = text.JSONKey(dst, Id)
	dst = text.AppendQuote(dst, ev.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = text.JSONKey(dst, Pubkey)
	dst = text.AppendQuote(dst, ev.PubKey, hex.EncAppend)
	dst = append(dst, ',')
	// CreatedAt
	dst = text.JSONKey(dst, CreatedAt)
	if dst, err = ev.CreatedAt.MarshalJSON(dst); chk.E(err) {
		return
	}
	dst = append(dst, ',')
	// Kind
	dst = text.JSONKey(dst, Kind)
	dst, _ = ev.Kind.MarshalJSON(dst)
	dst = append(dst, ',')
	// Tags
	dst = text.JSONKey(dst, Tags)
	dst, _ = ev.Tags.MarshalJSON(dst)
	dst = append(dst, ',')
	// Content
	dst = text.JSONKey(dst, Content)
	dst = text.AppendQuote(dst, ev.Content, text.NostrEscape)
	dst = append(dst, ',')
	// Sig
	dst = text.JSONKey(dst, Sig)
	dst = text.AppendQuote(dst, ev.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

func (ev *T) UnmarshalJSON(b B) (rem B, err error) {
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
				ev.Kind = kind.New(0)
				if rem, err = ev.Kind.UnmarshalJSON(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}
				ev.Tags = tags.New()
				if rem, err = ev.Tags.UnmarshalJSON(rem); chk.E(err) {
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
					if ev.Content, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
						return
					}
					state = betweenKV
				} else if key[1] == CreatedAt[1] {
					if len(key) < len(CreatedAt) {
						goto invalid
					}
					ev.CreatedAt = timestamp.New()
					if rem, err = ev.CreatedAt.UnmarshalJSON(rem); chk.E(err) {
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
