package ints

import (
	_ "embed"
)

// run this to regenerate (pointlessly) the base 10 array of 4 places per entry
//go:generate go run ./gen/.

//go:embed base10k.txt
var base10k []byte

const base = 10000

type T uint64

func New() T { return 0 }

var powers = []T{
	1,
	1_0000,
	1_0000_0000,
	1_0000_0000_0000,
	1_0000_0000_0000_0000,
}

const zero = '0'
const nine = '9'

// MarshalJSON encodes an uint64 into ASCII decimal format in a
// []byte.
func (n T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	if n == 0 {
		b = append(b, '0')
		return
	}
	var i int
	var trimmed bool
	k := len(powers)
	for k > 0 {
		k--
		q := n / powers[k]
		if !trimmed && q == 0 {
			continue
		}
		offset := q * 4
		bb := base10k[offset : offset+4]
		if !trimmed {
			for i = range bb {
				if bb[i] != '0' {
					bb = bb[i:]
					trimmed = true
					break
				}
			}
		}
		b = append(b, bb...)
		n = n - q*powers[k]
	}
	return
}

// UnmarshalJSON reads a string, which must be a positive integer no larger than
// math.MaxUint64, skipping any non-numeric content before it.
func (n T) UnmarshalJSON(b B) (a any, rem B, err error) {
	var sLen int
	// count the digits
	for ; sLen < len(b) && b[sLen] >= zero && b[sLen] <= nine && b[sLen] != ','; sLen++ {
	}
	if sLen == 0 {
		err = errorf.E("zero length number")
		return
	}
	if sLen > 20 {
		err = errorf.E("too big number for uint64")
		return
	}
	// the length of the string found
	rem = b[sLen:]
	b = b[:sLen]
	for _, ch := range b {
		ch -= zero
		n = n*10 + T(ch)
	}
	a = n
	return
}
