package filter

import (
	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/kind"
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

var (
	IDs     = B("id")
	Kinds   = B("kinds")
	Authors = B("authors")
	Tags    = B("tags")
	Since   = B("since")
	Until   = B("until")
	Limit   = B("limit")
	Search  = B("search")
)

func (t T) Marshal(dst B) (b B) {
	// open parentheses
	dst = append(dst, '{')
	if len(t.IDs) > 0 {
		dst = text.JSONKey(dst, IDs)
		dst = append(dst, '[')
		for i := range t.IDs {
			dst = text.AppendQuote(dst, t.IDs[i], hex.EncAppend)
			if i != len(t.IDs)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ']')
		dst = append(dst, ',')
	}
	if len(t.Kinds) > 0 {
		dst = text.JSONKey(dst, Kinds)
		dst = append(dst, '[')
		for i := range t.Kinds {
			dst = t.Kinds[i].Marshal(dst)
			if i != len(t.IDs)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ']')
		dst = append(dst, ',')
	}
	if len(t.Authors) > 0 {
		dst = text.JSONKey(dst, Authors)
		dst = append(dst, '[')
		for i := range t.IDs {
			dst = text.AppendQuote(dst, t.Authors[i], hex.EncAppend)
			if i != len(t.IDs)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ']')
		dst = append(dst, ',')
	}
	if len(t.Tags) > 0 {
		dst = text.JSONKey(dst, Tags)
		dst = t.Tags.Marshal(dst)
		dst = append(dst, ',')
	}
	if t.Since > 0 {
		dst = text.JSONKey(dst, Since)
		dst = ints.Int64AppendToByteString(dst, t.Since.I64())
		dst = append(dst, ',')
	}
	if t.Until > 0 {
		dst = text.JSONKey(dst, Until)
		dst = ints.Int64AppendToByteString(dst, t.Until.I64())
		dst = append(dst, ',')
	}
	if t.Limit > 0 {
		dst = text.JSONKey(dst, Limit)
		dst = ints.Int64AppendToByteString(dst, int64(t.Limit))
		dst = append(dst, ',')
	}
	if len(t.Search) > 0 {
		dst = text.JSONKey(dst, Search)
		dst = text.AppendQuote(dst, t.Search, text.NostrEscape)
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

func Unmarshal(b B) (f *T, rem B, err error) {
	f = &T{}
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
				if f.IDs, rem, err = UnmarshalHexArray(rem,
					sha256.Size); chk.E(err) {
					return
				}
				state = betweenKV
			case Kinds[0]:
				if len(key) < len(Kinds) {
					goto invalid
				}

				state = betweenKV
			case Authors[0]:
				if len(key) < len(Authors) {
					goto invalid
				}
				if f.IDs, rem, err = UnmarshalHexArray(rem,
					schnorr.PubKeyBytesLen); chk.E(err) {
					return
				}
				state = betweenKV
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}

				state = betweenKV
			case Until[0]:
				if len(key) < len(Until) {
					goto invalid
				}

				state = betweenKV
			case Limit[0]:
				if len(key) < len(Limit) {
					goto invalid
				}

				state = betweenKV
			case Search[0]:
				if len(key) < len(Since) {
					goto invalid
				}
				switch rem[1] {
				case Search[1]:
					if len(key) < len(Search) {
						goto invalid
					}

					state = betweenKV
				case Since[1]:
					if len(key) < len(Since) {
						goto invalid
					}

					state = betweenKV
				}
			}
		}
	}
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(rem)]),
		S(rem))
	return
}

// UnmarshalHexArray unpacks a JSON array containing strings with hexadecimal,
// and checks all values have the specified byte size..
func UnmarshalHexArray(b B, size int) (t []B, rem B, err error) {
	// change size to check before decoding hex
	size *= 2
	var inQuotes, openedBracket bool
	var quoteStart int
	for i := 0; i < len(b); i++ {
		if !openedBracket && b[i] == '[' {
			openedBracket = true
		} else if !inQuotes {
			if b[i] == '"' {
				inQuotes, quoteStart = true, i
			} else if b[i] == ']' {
				return t, b[i+1:], err
			}
		} else if b[i] == '"' {
			if i-quoteStart != size {
				err = errorf.E("value unexpected length\n%s\n%s", b[:i], b[i:])
			}
			i++
			var x B
			if x, rem, err = text.UnmarshalHex(b[quoteStart:i]); chk.E(err) {
				return
			}
			inQuotes, t = false, append(t, x)
		}
	}
	if !openedBracket || inQuotes {
		log.I.F("\n%v\n%s", t, rem)
		return nil, nil, errorf.E("tag: failed to list\n%s\n%s", t, rem)
	}
	return
}

func UnmarshalKindsArray(b B) (t kinds.T, rem B, err error) {
	rem = b
	var openedBracket bool
	for i := 0; i < len(b); i++ {
		if !openedBracket && b[i] == '[' {
			openedBracket = true
			continue
		} else if b[i] == ']' {
			// done
			return
		} else if b[i] == ',' {
			continue
		}
		var k int64
		if k, rem, err = ints.ExtractInt64FromByteString(rem); chk.E(err) {
			return
		}
		t = append(t, kind.T(k))
	}
	if !openedBracket {
		log.I.F("\n%v\n%s", t, rem)
		return nil, nil, errorf.E("tag: failed to list\n%s\n%s", t, rem)
	}
	return
}
