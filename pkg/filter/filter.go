package filter

import (
	"github.com/mleku/nodl/pkg/hex"
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
