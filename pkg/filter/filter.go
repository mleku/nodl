package filter

import (
	"bytes"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
)

// T is the primary query form for requesting events from a nostr relay.
type T struct {
	IDs     tag.T       `json:"ids,omitempty"`
	Kinds   kinds.T     `json:"kinds,omitempty"`
	Authors tag.T       `json:"authors,omitempty"`
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
		if dst, err = ints.T(f.Since).MarshalJSON(dst); chk.E(err) {
			return
		}
		dst = append(dst, ',')
	}
	if f.Until > 0 {
		dst = text.JSONKey(dst, Until)
		if dst, err = ints.T(f.Until).MarshalJSON(dst); chk.E(err) {
			return
		}
		dst = append(dst, ',')
	}
	if f.Limit > 0 {
		dst = text.JSONKey(dst, Limit)
		if dst, err = ints.T(f.Limit).MarshalJSON(dst); chk.E(err) {
			return
		}
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
				var u any
				if u, rem, err = ints.New().UnmarshalJSON(rem); chk.E(err) {
					return
				}
				f.Until = timestamp.T(u.(ints.T))
				state = betweenKV
			case Limit[0]:
				if len(key) < len(Limit) {
					goto invalid
				}
				var l any
				if l, rem, err = ints.New().UnmarshalJSON(rem); chk.E(err) {
					return
				}
				f.Limit = int(l.(ints.T))
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
					var s any
					if s, rem, err = ints.New().UnmarshalJSON(rem); chk.E(err) {
						return
					}
					f.Until = timestamp.T(s.(ints.T))
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

func (f *T) Matches(ev *event.T) bool {
	if ev == nil {
		// log.T.F("nil event")
		return false
	}
	if len(f.IDs) > 0 && !f.IDs.Contains(ev.ID) {
		// log.T.F("no ids in filter match event\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if len(f.Kinds) > 0 && !f.Kinds.Contains(ev.Kind) {
		// log.T.F("no matching kinds in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if len(f.Authors) > 0 && !f.Authors.Contains(ev.PubKey) {
		// log.T.F("no matching authors in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	for i, v := range f.Tags {
		// remove the hash prefix (idk why this thing even exists tbh)
		if bytes.HasPrefix(v[0], B("#")) {
			f.Tags[i][0] = f.Tags[i][0][1:]
		}
		if len(v) > 0 && !ev.Tags.ContainsAny(v[0], v...) {
			// log.T.F("no matching tags in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
			return false
		}
		// special case for p tags
	}
	if f.Since != 0 && ev.CreatedAt < f.Since {
		// log.T.F("event is older than since\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if f.Until != 0 && ev.CreatedAt > f.Until {
		// log.T.F("event is newer than until\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	return true
}
