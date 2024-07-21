package event

import (
	"io"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
)

func (ev *T) UnmarshalJSON(b B) (rem B, err error) {
	key := make(B, 0, 9)
	rem = b
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == '{' {
			rem = rem[1:]
			goto BetweenKeys
		}
	}
	goto eof
BetweenKeys:
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == '"' {
			rem = rem[1:]
			goto InKey
		}
	}
	goto eof
InKey:
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == '"' {
			rem = rem[1:]
			goto InKV
		} else {
			key = append(key, rem[0])
		}
	}
	goto eof
InKV:
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == ':' {
			rem = rem[1:]
			goto InVal
		}
	}
	goto eof
InVal:
	switch key[0] {
	case Id[0]:
		if !equals(Id, key) {
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
		goto BetweenKV
	case Pubkey[0]:
		if !equals(Pubkey, key) {
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
		goto BetweenKV
	case Kind[0]:
		if !equals(Kind, key) {
			goto invalid
		}
		ev.Kind = kind.New(0)
		if rem, err = ev.Kind.UnmarshalJSON(rem); chk.E(err) {
			return
		}
		goto BetweenKV
	case Tags[0]:
		if !equals(Tags, key) {
			goto invalid
		}
		ev.Tags = tags.New()
		if rem, err = ev.Tags.UnmarshalJSON(rem); chk.E(err) {
			return
		}
		goto BetweenKV
	case Sig[0]:
		if !equals(Sig, key) {
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
		goto BetweenKV
	case Content[0]:
		if key[1] == Content[1] {
			if !equals(Content, key) {
				goto invalid
			}
			if ev.Content, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
				return
			}
			goto BetweenKV
		} else if key[1] == CreatedAt[1] {
			if !equals(CreatedAt, key) {
				goto invalid
			}
			ev.CreatedAt = timestamp.New()
			if rem, err = ev.CreatedAt.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			goto BetweenKV
		} else {
			goto invalid
		}
	default:
		goto invalid
	}
BetweenKV:
	key = key[:0]
	for ; len(rem) > 0; rem = rem[1:] {
		switch {
		case len(rem) == 0:
			return
		case rem[0] == '}':
			rem = rem[1:]
			goto AfterClose
		case rem[0] == ',':
			rem = rem[1:]
			goto BetweenKeys
		case rem[0] == '"':
			rem = rem[1:]
			goto InKey
		}
	}
	goto eof
AfterClose:
	return
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(rem)]),
		S(rem))
	return
eof:
	err = io.EOF
	return
}
