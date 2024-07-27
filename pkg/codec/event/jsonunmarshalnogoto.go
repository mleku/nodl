package event

import (
	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/codec/tags"
	"git.replicatr.dev/pkg/codec/text"
	"git.replicatr.dev/pkg/codec/timestamp"
	"github.com/minio/sha256-simd"
)

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

var (
	states = []string{
		beforeOpen: "beforeOpen",
		openParen:  "openParen",
		inKey:      "inKey",
		inKV:       "inKV",
		inVal:      "inVal",
		betweenKV:  "betweenKV",
		afterClose: "afterClose",
	}
)

func (ev *T) UnmarshalJSONold(b B) (rem B, err error) {
	rem = b[:]
	var key B
	var state int
	for ; len(rem) > 0; rem = rem[1:] {
		log.I.F("%s %s", rem, states[state])
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
			case jId[0]:
				if len(key) < len(jId) {
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
			case jPubkey[0]:
				if len(key) < len(jPubkey) {
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
			case jKind[0]:
				if len(key) < len(jKind) {
					goto invalid
				}
				ev.Kind = kind.New(0)
				if rem, err = ev.Kind.UnmarshalJSON(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case jTags[0]:
				if len(key) < len(jTags) {
					goto invalid
				}
				ev.Tags = tags.New()
				if rem, err = ev.Tags.UnmarshalJSON(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case jSig[0]:
				if len(key) < len(jSig) {
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
			case jContent[0]:
				// this can be one of two, but minimum of the shortest
				if len(key) < len(jContent) {
					goto invalid
				}
				if key[1] == jContent[1] {
					if ev.Content, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
						return
					}
					state = betweenKV
				} else if key[1] == jCreatedAt[1] {
					if len(key) < len(jCreatedAt) {
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
	if len(rem) != 0 && state != afterClose {
		log.I.Ln("state", states[state])
		log.I.F("position\n%d %s\n\n%d %s", len(b)-len(rem),
			b[:len(b)-len(rem)],
			len(rem), rem)
		err = errorf.E("invalid event")
	}
	return
invalid:
	log.I.Ln("state", states[state])
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(rem)]),
		S(rem))
	return
}
