package tags

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"

	"git.replicatr.dev/pkg/codec/tag"
)

// T is a list of T - which are lists of string elements with ordering and no
// uniqueness constraint (not a set).
type T struct {
	T []*tag.T
}

func New(fields ...*tag.T) (t *T) {
	// t = &T{T: make([]*tag.T, 0, len(fields))}
	t = &T{}
	for _, field := range fields {
		t.T = append(t.T, field)
	}
	return
}

func (t *T) ToStringSlice() (b [][]S) {
	b = make([][]S, 0, len(t.T))
	for i := range t.T {
		b = append(b, t.T[i].ToStringSlice())
	}
	return
}

func (t *T) Clone() (c *T) {
	c = &T{T: make([]*tag.T, len(t.T))}
	for i, field := range t.T {
		c.T[i] = field.Clone()
	}
	return
}

func (t *T) Equal(ta *T) bool {
	// sort them the same so if they are the same in content they compare the same.
	t1 := t.Clone()
	sort.Sort(t1)
	t2 := ta.Clone()
	sort.Sort(t2)
	for i := range t.T {
		if !t1.T[i].Equal(t2.T[i]) {
			return false
		}
	}
	return true
}

// Less evaluates two tags and returns if one is less than the other. Fields are lexicographically compared, so a < b.
//
// If two tags are greater than or equal up to the length of the shortest.
func (t *T) Less(i, j int) (less bool) {
	var field int
	for {
		// if they are greater or equal, the longer one is greater because nil is less than anything.
		if t.T[i].Len() <= field || t.T[j].Len() <= field {
			return t.T[i].Len() < t.T[j].Len()
		}
		if bytes.Compare(t.T[i].Field[field], t.T[j].Field[field]) < 0 {
			return true
		}
		field++
	}
}

func (t *T) Swap(i, j int) {
	t.T[i], t.T[j] = t.T[j], t.T[i]
}

func (t *T) Len() (l int) { return len(t.T) }

// GetFirst gets the first tag in tags that matches the prefix, see [T.StartsWith]
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
func (t *T) GetAll(tagPrefix *tag.T) *T {
	result := &T{T: make([]*tag.T, 0, len(t.T))}
	for _, v := range t.T {
		if v.StartsWith(tagPrefix) {
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

func (t *T) Append(ttt ...*T) {
	for _, tt := range ttt {
		for _, v := range tt.T {
			t.T = append(t.T, v)
		}
	}
}

// Scan parses a string or raw bytes that should be a string and embeds the values into the tags variable from which
// this method is invoked.
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

// ContainsAny returns true if any of the strings given in `values` matches any of the tag elements.
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

// MarshalTo appends the JSON encoded byte of T as [][]string to dst. String escaping is as described in RFC8259.
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

// func (t *T) String() string {
// 	buf := new(bytes.Buffer)
// 	buf.WriteByte('[')
// 	last := len(t.T) - 1
// 	for i := range t.T {
// 		_, _ = fmt.Fprint(buf, t.T[i])
// 		if i < last {
// 			buf.WriteByte(',')
// 		}
// 	}
// 	buf.WriteByte(']')
// 	return buf.String()
// }

func (t *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	if t.T == nil {
		b = append(b, ']')
		return
	}
	if len(t.T) == 0 {
		b = append(b, '[', ']')
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
			r = r[1:]
			goto inTags
		case ',':
			r = r[1:]
			// next
		case ']':
			r = r[1:]
			// the end
			return
		default:
			r = r[1:]
		}
	inTags:
		for len(r) > 0 {
			switch r[0] {
			case '[':
				tt := &tag.T{}
				if r, err = tt.UnmarshalJSON(r); chk.E(err) {
					return
				}
				t.T = append(t.T, tt)
			case ',':
				r = r[1:]
				// next
			case ']':
				r = r[1:]
				// the end
				return
			default:
				r = r[1:]
			}
		}
	}
	return
}
