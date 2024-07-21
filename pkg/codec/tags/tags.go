package tags

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mleku/nodl/pkg/codec/tag"
)

// T is a list of T - which are lists of string elements with ordering and no
// uniqueness constraint (not a set).
type T struct {
	T []*tag.T
}

func New(fields ...*tag.T) (t *T) {
	t = &T{T: make([]*tag.T, len(fields))}
	for i, field := range fields {
		t.T[i] = field
	}
	return
}

// GetFirst gets the first tag in tags that matches the prefix, see
// [T.StartsWith]
func (t *T) GetFirst(tagPrefix *tag.T) *tag.T {
	for _, v := range t.T {
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetLast gets the last tag in tags that matches the prefix, see [T.StartsWith]
func (t *T) GetLast(tagPrefix *tag.T) *tag.T {
	for i := len(t.T) - 1; i >= 0; i-- {
		v := t.T[i]
		if v.StartsWith(tagPrefix) {
			return v
		}
	}
	return nil
}

// GetAll gets all the tags that match the prefix, see [T.StartsWith]
func (t *T) GetAll(tagPrefix ...B) *T {
	result := &T{T: make([]*tag.T, 0, len(t.T))}
	for _, v := range t.T {
		if v.StartsWith(tag.New(tagPrefix...)) {
			result.T = append(result.T, v)
		}
	}
	return result
}

// FilterOut removes all tags that match the prefix, see [T.StartsWith]
func (t *T) FilterOut(tagPrefix []B) *T {
	filtered := &T{T: make([]*tag.T, 0, len(t.T))}
	for _, v := range t.T {
		if !v.StartsWith(tag.New(tagPrefix...)) {
			filtered.T = append(filtered.T, v)
		}
	}
	return filtered
}

// AppendUnique appends a tag if it doesn't exist yet, otherwise does nothing.
// the uniqueness comparison is done based only on the first 2 elements of the
// tag.
func (t *T) AppendUnique(tag *tag.T) *T {
	n := tag.Len()
	if n > 2 {
		n = 2
	}
	if t.GetFirst(tag.Slice(0, n)) == nil {
		return &T{append(t.T, tag)}
	}
	return t
}

// Scan parses a string or raw bytes that should be a string and embeds the
// values into the tags variable from which this method is invoked.
//
// todo: wut is this?
func (t *T) Scan(src any) (err error) {
	var jtags []byte
	switch v := src.(type) {
	case B:
		jtags = v
	case S:
		jtags = []byte(v)
	default:
		return errors.New("couldn't scan tag, it's not a json string")
	}
	err = json.Unmarshal(jtags, &t)
	chk.E(err)
	return
}

// ContainsAny returns true if any of the strings given in `values` matches any
// of the tag elements.
func (t *T) ContainsAny(tagName B, values ...B) bool {
	for _, v := range t.T {
		if v.Len() < 2 {
			continue
		}
		if !equals(v.Key(), tagName) {
			continue
		}
		for _, candidate := range values {
			if equals(v.Value(), candidate) {
				return true
			}
		}
	}
	return false
}

// MarshalTo appends the JSON encoded byte of T as [][]string to dst. String
// escaping is as described in RFC8259.
func (t *T) MarshalTo(dst B) []byte {
	dst = append(dst, '[')
	for i, tt := range t.T {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst, _ = tt.MarshalJSON(dst)
	}
	dst = append(dst, ']')
	return dst
}

func (t *T) String() string {
	buf := new(bytes.Buffer)
	buf.WriteByte('[')
	last := len(t.T) - 1
	for i := range t.T {
		_, _ = fmt.Fprint(buf, t.T[i])
		if i < last {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(']')
	return buf.String()
}

func (t *T) Slice() (slice [][]B) {
	for i := range t.T {
		slice = append(slice, t.T[i].T)
	}
	return
}

func (t *T) Equal(ta any) bool {
	if t1, ok := ta.(*T); ok {
		for i := range t.T {
			if !t.T[i].Equal(t1.T) {
				return false
			}
		}
	}
	return true
}

func (t *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	if t == nil {
		b = append(b, ']')
		return
	}
	for i, s := range t.T {
		if i > 0 {
			b = append(b, ',')
		}
		b, _ = s.MarshalJSON(b)
	}
	b = append(b, ']')
	return
}

func (t *T) UnmarshalJSON(b B) (r B, err error) {
	r = b[:]
	for len(r) > 0 {
		switch r[0] {
		case '[':
			if r[1] == '[' {
				r = r[1:]
				continue
			} else if r[1] == ']' {
				r = r[1:]
				return
			}
			tt := tag.NewWithCap(4) // most tags are 4 or less fields
			if r, err = tt.UnmarshalJSON(r); chk.E(err) {
				return
			}
			t.T = append(t.T, tt)
			// continue
		case ',':
			r = r[1:]
			// next
		case ']':
			r = r[1:]
			// the end
			return
		}
	}
	return
}

func (t *T) Clone() (t1 *T) {
	t1 = New()
	for _, x := range t.T {
		t1.T = append(t1.T, tag.NewWithCap(x.Len()))
		for j, y := range x.T {
			t1.T[j].T = append(t1.T[j].T, y)
		}
	}
	return
}
