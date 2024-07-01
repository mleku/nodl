package kinds

import (
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/kind"
)

type T []kind.T

func FromIntSlice(is []int) (k T) {
	for i := range is {
		k = append(k, kind.T(is[i]))
	}
	return
}

func (k T) ToUint16() (o []uint16) {
	for i := range k {
		o = append(o, uint16(k[i]))
	}
	return
}

// Clone makes a new kind.T with the same members.
func (k T) Clone() (c T) {
	c = make(T, len(k))
	for i := range k {
		c[i] = k[i]
	}
	return
}

// Contains returns true if the provided element is found in the kinds.T.
//
// Note that the request must use the typed kind.T or convert the number thus.
// Even if a custom number is found, this codebase does not have the logic to
// deal with the kind so such a search is pointless and for which reason static
// typing always wins. No mistakes possible with known quantities.
func (k T) Contains(s kind.T) bool {
	for i := range k {
		if k[i] == s {
			return true
		}
	}
	return false
}

// Equals checks that the provided kind.T matches.
func (k T) Equals(t1 T) bool {
	if len(k) != len(t1) {
		return false
	}
	for i := range k {
		if k[i] != t1[i] {
			return false
		}
	}
	return true
}

func (k T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	for i := range k {
		b, _ = k[i].MarshalJSON(b)
		if i != len(k)-1 {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (k T) UnmarshalJSON(b B) (a any, rem B, err error) {
	rem = b
	var openedBracket bool
	for ; len(rem) > 0; rem = rem[1:] {
		if !openedBracket && rem[0] == '[' {
			openedBracket = true
			continue
		} else if openedBracket {
			if rem[0] == ']' {
				// done
				return
			} else if rem[0] == ',' {
				continue
			}
			var kk any
			if kk, rem, err = ints.New().UnmarshalJSON(rem); chk.E(err) {
				return
			}
			k = append(k, kind.T(kk.(ints.T)))
			if rem[0] == ']' {
				rem = rem[1:]
				a = k
				return
			}
		}
	}
	if !openedBracket {
		log.I.F("\n%v\n%s", k, rem)
		return nil, nil, errorf.E("kinds: failed to unmarshal\n%s\n%s\n%s", k,
			b, rem)
	}
	a = k
	return
}
