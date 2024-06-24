package text

import (
	"io"

	"github.com/templexxx/xhex"
)

// JSONKey generates the JSON format for an object key and terminates with the
// semicolon.
func JSONKey(dst, k B) (b B) {
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, '"', ':')
	b = dst
	return
}

// UnmarshalHex takes a byte string that should contain a quoted hexadecimal
// encoded value, decodes it in-place using a SIMD hex codec and returns the
// decoded truncated bytes (the other half will be as it was but no allocation
// is required).
func UnmarshalHex(b B) (h B, rem B, err error) {
	rem = b[:]
	var inQuote bool
	var start int
	for i := 0; i < len(b); i++ {
		if !inQuote {
			if b[i] == '"' {
				inQuote = true
				start = i + 1
			}
		} else {
			if b[i] == '"' {
				h = b[start:i]
				rem = b[i+1:]
				break
			}
		}
	}
	if !inQuote {
		err = io.EOF
		return
	}
	l := len(h)
	if l%2 != 0 {
		err = errorf.E("invalid length for hex: %d, %0x", len(h), h)
		return
	}
	if err = xhex.Decode(h, h); chk.E(err) {
		return
	}
	h = h[:l/2]
	return
}
