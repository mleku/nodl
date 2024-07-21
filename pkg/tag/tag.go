package tag

import (
	"bytes"
	"unsafe"

	"github.com/mleku/nodl/pkg/normalize"
	"github.com/mleku/nodl/pkg/text"
)

// The tag position meanings so they are clear when reading.
const (
	Key = iota
	Value
	Relay
)

// T marker strings for e (reference) tags.
const (
	MarkerReply   = "reply"
	MarkerRoot    = "root"
	MarkerMention = "mention"
)

// T is a list of strings with a literal ordering.
//
// Not a set, there can be repeating elements.
type T struct {
	T []B
}

func NewWithCap(c int) *T { return &T{make([]B, 0, c)} }

func New[V string | B](fields ...V) (t *T) {
	t = &T{T: make([]B, len(fields))}
	for i, field := range fields {
		t.T[i] = B(field)
	}
	return
}

func (t *T) Append(b B)              { t.T = append(t.T, b) }
func (t *T) Len() int                { return len(t.T) }
func (t *T) Cap() int                { return cap(t.T) }
func (t *T) Clear()                  { t.T = t.T[:0] }
func (t *T) Slice(start, end int) *T { return &T{t.T[start:end]} }

// StartsWith checks a tag has the same initial set of elements.
//
// The last element is treated specially in that it is considered to match if
// the candidate has the same initial substring as its corresponding element.
func (t *T) StartsWith(prefix *T) bool {
	prefixLen := len(prefix.T)

	if prefixLen > len(t.T) {
		return false
	}
	// check initial elements for equality
	for i := 0; i < prefixLen-1; i++ {
		if !equals(prefix.T[i], t.T[i]) {
			return false
		}
	}
	// check last element just for a prefix
	return bytes.HasPrefix(t.T[prefixLen-1], prefix.T[prefixLen-1])
}

// Key returns the first element of the tags.
func (t *T) Key() B {
	if len(t.T) > Key {
		return t.T[Key]
	}
	return nil
}

// Value returns the second element of the tag.
func (t *T) Value() B {
	if len(t.T) > Value {
		return t.T[Value]
	}
	return nil
}

var etag, ptag = B("e"), B("p")

// Relay returns the third element of the tag.
func (t *T) Relay() (s B) {
	if (equals(t.Key(), etag) ||
		equals(t.Key(), ptag)) &&
		len(t.T) >= Relay {

		return normalize.URL(t.T[Relay])
	}
	return
}

// MarshalJSON appends the JSON form to the passed bytes.
func (t *T) MarshalJSON(dst B) (b B, err error) {
	dst = append(dst, '[')
	for i, s := range t.T {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = text.AppendQuote(dst, s, text.NostrEscape)
	}
	dst = append(dst, ']')
	return dst, err
}

// UnmarshalJSON decodes the provided JSON tag list (array of strings), and
// returns any remainder after the close bracket has been encountered.
func (t *T) UnmarshalJSON(b B) (r B, err error) {
	var inQuotes, openedBracket bool
	var quoteStart int
	for i := 0; i < len(b); i++ {
		if !openedBracket && b[i] == '[' {
			openedBracket = true
		} else if !inQuotes {
			if b[i] == '"' {
				inQuotes, quoteStart = true, i+1
			} else if b[i] == ']' {
				return b[i+1:], err
			}
		} else if b[i] == '\\' && i < len(b)-1 {
			i++
		} else if b[i] == '"' {
			inQuotes = false
			t.T = append(t.T, text.NostrUnescape(b[quoteStart:i]))
		}
	}
	if !openedBracket || inQuotes {
		log.I.F("\n%v\n%s", t, r)
		return nil, errorf.E("tag: failed to parse tag")
	}
	return
}

func (t *T) String() string {
	b, _ := t.MarshalJSON(nil)
	return unsafe.String(&b[0], len(b))
}

// Clone makes a new tag.T with the same members.
func (t *T) Clone() (c *T) {
	c = &T{T: make([]B, len(t.T))}
	for i := range t.T {
		c.T[i] = t.T[i]
	}
	return
}

// Contains returns true if the provided element is found in the tag slice.
func (t *T) Contains(s B) bool {
	for i := range t.T {
		if equals(t.T[i], s) {
			return true
		}
	}
	return false
}

// Equal checks that the provided tag list matches.
func (t *T) Equal(ta any) bool {
	if t1, ok := ta.(T); ok {
		if len(t.T) != len(t1.T) {
			return false
		}
		for i := range t.T {
			if !equals(t.T[i], t1.T[i]) {
				return false
			}
		}
	}
	return true
}
