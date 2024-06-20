package ints

import _ "embed"

// run this to regenerate (pointlessly) the base 10 array of 4 places per entry
//go:generate go run gen/pregen.go

//go:embed base10k.txt
var base10k []byte

const base = 10000

var powers []int64

func init() {
	var begin int64 = 1
	for _ = range 20 {
		powers = append(powers, begin)
		begin *= 10
	}
}

func ExtractInt64FromByteString(b B) (n int64, rem B, err error) {
	var i int
	rem = b
	if len(b) == 0 {
		return
	}
	// count the digits
	for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
	}
	// the length of the string found
	rem = b[i:]
	b = b[:i]
	for j := range b {
		n += int64(b[j]-'0') * powers[i-j-1]
	}
	return
}

// Int64AppendToByteString encodes an int64 into ASCII decimal format in a
// []byte.
func Int64AppendToByteString(dst []byte, n int64) (b []byte) {
	var i int
	q := n / base
	r := 4 * (n - q*base)
	n = q
	bb := base10k[r : r+4]
	for i = range bb {
		if bb[i] != '0' {
			bb = bb[i:]
			break
		}
	}
	dst = append(dst, bb...)
	return dst
}
