package filter

import (
	"bytes"
	"fmt"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/btcec/secp256k1"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/tags"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
	"lukechampine.com/frand"
)

// T is the primary query form for requesting events from a nostr relay.
type T struct {
	IDs     *tag.T       `json:"ids,omitempty"`
	Kinds   *kinds.T     `json:"kinds,omitempty"`
	Authors *tag.T       `json:"authors,omitempty"`
	Tags    *tags.T      `json:"-,omitempty"`
	Since   *timestamp.T `json:"since,omitempty"`
	Until   *timestamp.T `json:"until,omitempty"`
	Limit   int          `json:"limit,omitempty"`
	Search  B            `json:"search,omitempty"`
}

func New() (f *T) {
	return &T{
		IDs:     tag.NewWithCap(100),
		Kinds:   kinds.NewWithCap(10),
		Authors: tag.NewWithCap(100),
		Tags:    tags.New(),
		Since:   timestamp.New(),
		Until:   timestamp.New(),
		Limit:   0,
		Search:  nil,
	}
}

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
	if f.IDs != nil && len(f.IDs.T) > 0 {
		dst = text.JSONKey(dst, IDs)
		dst = text.MarshalHexArray(dst, f.IDs.T)
		dst = append(dst, ',')
	}
	if f.Kinds != nil && len(f.Kinds.K) > 0 {
		dst = text.JSONKey(dst, Kinds)
		if dst, err = f.Kinds.MarshalJSON(dst); chk.E(err) {
			return
		}
		dst = append(dst, ',')
	}
	if f.Authors != nil && len(f.Authors.T) > 0 {
		dst = text.JSONKey(dst, Authors)
		dst = text.MarshalHexArray(dst, f.Authors.T)
		dst = append(dst, ',')
	}
	if f.Tags != nil && len(f.Tags.T) > 0 {
		dst = text.JSONKey(dst, Tags)
		dst, _ = f.Tags.MarshalJSON(dst)
		dst = append(dst, ',')
	}
	if f.Since != nil && f.Since.U64() > 0 {
		dst = text.JSONKey(dst, Since)
		if dst, err = f.Since.MarshalJSON(dst); chk.E(err) {
			return
		}
		dst = append(dst, ',')
	}
	if f.Until != nil && f.Until.U64() > 0 {
		dst = text.JSONKey(dst, Until)
		if dst, err = f.Until.MarshalJSON(dst); chk.E(err) {
			return
		}
		dst = append(dst, ',')
	}
	if f.Limit > 0 {
		dst = text.JSONKey(dst, Limit)
		if dst, err = ints.New(f.Limit).MarshalJSON(dst); chk.E(err) {
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

func (f *T) UnmarshalJSON(b B) (r B, err error) {
	r = b[:]
	var key B
	var state int
	for ; len(r) >= 0; r = r[1:] {
		// log.I.F("%c", rem[0])
		switch state {
		case beforeOpen:
			if r[0] == '{' {
				state = openParen
				// log.I.Ln("openParen")
			}
		case openParen:
			if r[0] == '"' {
				state = inKey
				// log.I.Ln("inKey")
			}
		case inKey:
			if r[0] == '"' {
				state = inKV
				// log.I.Ln("inKV")
			} else {
				key = append(key, r[0])
			}
		case inKV:
			if r[0] == ':' {
				state = inVal
			}
		case inVal:
			switch key[0] {
			case IDs[0]:
				if len(key) < len(IDs) {
					goto invalid
				}
				f.IDs = tag.New("")
				if f.IDs.T, r, err = text.UnmarshalHexArray(r,
					sha256.Size); chk.E(err) {
					return
				}
				state = betweenKV
				// // log.I.Ln("betweenKV")
			case Kinds[0]:
				if len(key) < len(Kinds) {
					goto invalid
				}
				f.Kinds = kinds.NewWithCap(0)
				if r, err = f.Kinds.UnmarshalJSON(r); chk.E(err) {
					return
				}
				state = betweenKV
				// log.I.Ln("betweenKV")
			case Authors[0]:
				if len(key) < len(Authors) {
					goto invalid
				}
				f.Authors = tag.New("")
				if f.Authors.T, r, err = text.UnmarshalHexArray(r,
					schnorr.PubKeyBytesLen); chk.E(err) {
					return
				}
				state = betweenKV
				// log.I.Ln("betweenKV")
			case Tags[0]:
				if len(key) < len(Tags) {
					goto invalid
				}
				f.Tags = tags.New()
				if r, err = f.Tags.UnmarshalJSON(r); chk.E(err) {
					return
				}
				state = betweenKV
				// log.I.Ln("betweenKV")
			case Until[0]:
				if len(key) < len(Until) {
					goto invalid
				}
				u := ints.New(0)
				if r, err = u.UnmarshalJSON(r); chk.E(err) {
					return
				}
				f.Until = timestamp.FromUnix(int64(u.N))
				state = betweenKV
				// log.I.Ln("betweenKV")
			case Limit[0]:
				if len(key) < len(Limit) {
					goto invalid
				}
				l := ints.New(0)
				if r, err = l.UnmarshalJSON(r); chk.E(err) {
					return
				}
				f.Limit = int(l.N)
				state = betweenKV
				// log.I.Ln("betweenKV")
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
					if txt, r, err = text.UnmarshalQuoted(r); chk.E(err) {
						return
					}
					f.Search = txt
					// log.I.F("\n%s\n%s", txt, rem)
					state = betweenKV
					// log.I.Ln("betweenKV")
				case Since[1]:
					if len(key) < len(Since) {
						goto invalid
					}
					s := ints.New(0)
					if r, err = s.UnmarshalJSON(r); chk.E(err) {
						return
					}
					f.Since = timestamp.FromUnix(int64(s.N))
					state = betweenKV
					// log.I.Ln("betweenKV")
				}
			default:
				goto invalid
			}
			key = key[:0]
		case betweenKV:
			if len(r) == 0 {
				return
			}
			if r[0] == '}' {
				state = afterClose
				// log.I.Ln("afterClose")
				// rem = rem[1:]
			} else if r[0] == ',' {
				state = openParen
				// log.I.Ln("openParen")
			} else if r[0] == '"' {
				state = inKey
				// log.I.Ln("inKey")
			}
		}
		if r[0] == '}' {
			r = r[1:]
			return
		}
	}
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", S(b), S(b[:len(r)]),
		S(r))
	return
}

func (f *T) Matches(ev *event.T) bool {
	if ev == nil {
		// log.T.F("nil event")
		return false
	}
	if f.IDs != nil && len(f.IDs.T) > 0 && !f.IDs.Contains(ev.ID) {
		// log.T.F("no ids in filter match event\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if f.Kinds != nil && len(f.Kinds.K) > 0 && !f.Kinds.Contains(ev.Kind) {
		// log.T.F("no matching kinds in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if f.Authors != nil && len(f.Authors.T) > 0 && !f.Authors.Contains(ev.PubKey) {
		// log.T.F("no matching authors in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if f.Tags != nil {
		for i, v := range f.Tags.T {
			// remove the hash prefix (idk why this thing even exists tbh)
			if bytes.HasPrefix(v.T[0], B("#")) {
				f.Tags.T[i].T[0] = f.Tags.T[i].T[0][1:]
			}
			if len(v.T) > 0 && !ev.Tags.ContainsAny(v.T[0], v.T...) {
				// log.T.F("no matching tags in filter\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
				return false
			}
			// special case for p tags
		}
	}
	if f.Since != nil && f.Since.Int() != 0 && ev.CreatedAt != nil && ev.CreatedAt.I64() < f.Since.I64() {
		// log.T.F("event is older than since\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	if f.Until != nil && f.Until.Int() != 0 && ev.CreatedAt.I64() > f.Until.I64() {
		// log.T.F("event is newer than until\nEVENT %s\nFILTER %s", ev.ToObject().String(), f.ToObject().String())
		return false
	}
	return true
}

func GenFilter() (f *T, err error) {
	f = New()
	n := frand.Intn(16)
	for _ = range n {
		id := make(B, sha256.Size)
		frand.Read(id)
		f.IDs.T = append(f.IDs.T, id)
	}
	n = frand.Intn(16)
	for _ = range n {
		f.Kinds.K = append(f.Kinds.K, kind.New(frand.Intn(65535)))
	}
	n = frand.Intn(16)
	for _ = range n {
		var sk *secp256k1.SecretKey
		if sk, err = secp256k1.GenerateSecretKey(); chk.E(err) {
			return
		}
		pk := sk.PubKey()
		f.Authors.T = append(f.Authors.T, schnorr.SerializePubKey(pk))
	}
	a := frand.Intn(16)
	if a < n {
		n = a
	}
	for i := range n {
		p := make(B, 0, schnorr.PubKeyBytesLen*2)
		p = hex.EncAppend(p, f.Authors.T[i])
		f.Tags.T = append(f.Tags.T, tag.New(B("p"), p))
		idb := make(B, sha256.Size)
		frand.Read(idb)
		id := make(B, 0, sha256.Size*2)
		id = hex.EncAppend(id, idb)
		f.Tags.T = append(f.Tags.T, tag.New(B("e"), id))
		f.Tags.T = append(f.Tags.T,
			tag.New(B("a"),
				B(fmt.Sprintf("%d:%s:", frand.Intn(65535), id))))
	}
	tn := int(timestamp.Now().I64())
	before := timestamp.T(tn - frand.Intn(10000))
	f.Since = &before
	f.Until = timestamp.Now()
	f.Search = B("token search text")
	return
}
