package filter

import (
	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
)

// T is the primary query form for requesting events from a nostr relay.
type T struct {
	IDs     []B         `json:"ids,omitempty"`
	Kinds   kinds.T     `json:"kinds,omitempty"`
	Authors []B         `json:"authors,omitempty"`
	Tags    tags.T      `json:"-,omitempty"`
	Since   timestamp.T `json:"since,omitempty"`
	Until   timestamp.T `json:"until,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	Search  B           `json:"search,omitempty"`
}

func New() (f *T) { return &T{} }

var (
	IDs     = B("ids")
	Kinds   = B("kinds")
	Authors = B("authors")
	Tags    = B("tags")
	Since   = B("since")
	Until   = B("until")
	Limit   = B("limit")
	Search  = B("search")
)

func (f *T) MarshalJSON(dst B) (b B, err error) {
	// open parentheses
	dst = append(dst, '{')
	if len(f.IDs) > 0 {
		dst = text.JSONKey(dst, IDs)
		dst = text.MarshalHexArray(dst, f.IDs)
		dst = append(dst, ',')
	}
	if len(f.Kinds) > 0 {
		dst = text.JSONKey(dst, Kinds)
		dst = text.MarshalKindsArray(dst, f.Kinds)
		dst = append(dst, ',')
	}
	if len(f.Authors) > 0 {
		dst = text.JSONKey(dst, Authors)
		dst = text.MarshalHexArray(dst, f.Authors)
		dst = append(dst, ',')
	}
	if len(f.Tags) > 0 {
		dst = text.JSONKey(dst, Tags)
		dst, _ = f.Tags.MarshalJSON(dst)
		dst = append(dst, ',')
	}
	if f.Since > 0 {
		dst = text.JSONKey(dst, Since)
		dst = ints.Int64AppendToByteString(dst, f.Since.I64())
		dst = append(dst, ',')
	}
	if f.Until > 0 {
		dst = text.JSONKey(dst, Until)
		dst = ints.Int64AppendToByteString(dst, f.Until.I64())
		dst = append(dst, ',')
	}
	if f.Limit > 0 {
		dst = text.JSONKey(dst, Limit)
		dst = ints.Int64AppendToByteString(dst, int64(f.Limit))
		dst = append(dst, ',')
	}
	if len(f.Search) > 0 {
		dst = text.JSONKey(dst, Search)
		dst = text.AppendQuote(dst, f.Search, text.NostrEscape)
	}
	// close parentheses
	dst = append(dst, '}')
	b = dst
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

func (f *T) UnmarshalJSON(b B) (fa any, rem B, err error) {
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
			case IDs[0]:
				if len(key) < len(IDs) {
					goto invalid
				}
				if f.IDs, rem, err = text.UnmarshalHexArray(rem,
					sha256.Size); chk.E(err) {
					return
				}
				state = betweenKV
			case Kinds[0]:
				if len(key) < len(Kinds) {
					goto invalid
				}
				if f.Kinds, rem, err = text.UnmarshalKindsArray(rem); chk.E(err) {
					return
				}
				state = betweenKV
			case Authors[0]:
				if len(key) < len(Authors) {
					goto invalid
				}
				if f.Authors, rem, err = text.UnmarshalHexArray(rem,
					schnorr.PubKeyBytesLen); chk.E(err) {
					return
				}
				state = betweenKV
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}
				var ta any
				if ta, rem, err = f.Tags.UnmarshalJSON(rem); chk.E(err) {
					return
				}
				f.Tags = ta.(tags.T)
				state = betweenKV
			case Until[0]:
				if len(key) < len(Until) {
					goto invalid
				}
				var u int64
				if u, rem, err = ints.ExtractInt64FromByteString(rem); chk.E(err) {
					return
				}
				f.Until = timestamp.T(u)
				state = betweenKV
			case Limit[0]:
				if len(key) < len(Limit) {
					goto invalid
				}
				var l int64
				if l, rem, err = ints.ExtractInt64FromByteString(rem); chk.E(err) {
					return
				}
				f.Limit = int(l)
				state = betweenKV
			case Search[0]:
				if len(key) < len(Since) {
					goto invalid
				}
				switch key[1] {
				case Search[1]:
					if len(key) < len(Search) {
						goto invalid
					}
					var txt B
					if txt, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
						return
					}
					f.Search = txt
					state = betweenKV
				case Since[1]:
					if len(key) < len(Since) {
						goto invalid
					}
					var s int64
					if s, rem, err = ints.ExtractInt64FromByteString(rem); chk.E(err) {
						return
					}
					f.Until = timestamp.T(s)
					state = betweenKV
				}
			default:
				goto invalid
			}
			key = key[:0]
		case betweenKV:
			if len(rem) == 0 {
				fa = f
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
	fa = f
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(rem)]),
		S(rem))
	return
}
