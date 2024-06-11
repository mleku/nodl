package ints

import (
	"os"

	"mleku.net/g/nodl/pkg/utils/lol"
)

var log, chk = lol.New(os.Stderr)

func ByteStringToInt64(b []byte) (n int64, err error) {
	place := int64(1)
	for i := range b {
		if b[i] < '0' || b[i] > '9' {
			err = log.E.Err("timestamp: invalid byte %q", b[i])
			return
		}
		n += int64(b[i]-'0') * place
		if i < len(b) {
			place *= 10
		}
	}
	return

}

func Int64AppendToByteString(dst []byte, n int64) (b []byte) {
	var d, m int64
	for d = n; d > 0; d /= 10 {
		m = d % 10
		dst = append(dst, '0'+byte(m))
	}
	return dst
}
