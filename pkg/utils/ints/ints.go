package ints

import (
	"os"

	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

// ByteStringToInt64 decodes a decimal number encoded in []byte (rather than
// string) - there is no stdlib equivalent and this is ~2x as fast as
// strconv.Atoi because of string memory handling.
func ByteStringToInt64(b []byte) (n int64, err error) {
	place := int64(1)
	for i := range b {
		if b[i] < '0' || b[i] > '9' {
			err = errorf.E("timestamp: invalid byte %q", b[i])
			return
		}
		n += int64(b[i]-'0') * place
		if i < len(b) {
			place *= 10
		}
	}
	return
}

// Int64AppendToByteString encodes an int64 into ASCII decimal format in a
// []byte - it is almost 3x faster than using strconv.Itoa because of string
// memory handling.
func Int64AppendToByteString(dst []byte, n int64) (b []byte) {
	var d, m int64
	for d = n; d > 0; d /= 10 {
		m = d % 10
		dst = append(dst, '0'+byte(m))
	}
	return dst
}
