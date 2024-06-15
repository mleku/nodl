package bstring

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"os"
	"unsafe"

	"github.com/mleku/nodl/pkg/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type T []byte

func (t T) String() string {
	return unsafe.String(&(t[0]), len(t))
}

func (t T) Equal(other T) bool { return bytes.Equal(t, other) }

type AppendBytesClosure func(dst, src T) T

type AppendClosure func(dst T) T

func Unquote(b T) T { return b[1 : len(b)-1] }

func Noop(dst, src T) T { return append(dst, src...) }

func AppendQuote(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

func Quote(dst, src T) T { return AppendQuote(dst, src, Noop) }

func AppendSingleQuote(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

func AppendBackticks(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

func AppendBrace(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

func AppendParenthesis(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

func AppendBracket(dst, src T, ac AppendBytesClosure) T {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

func AppendList(dst T, src []T, separator byte, ac AppendBytesClosure) T {
	last := len(src) - 1
	for i := range src {
		dst = append(dst, ac(dst, src[i])...)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}

func AppendHex(dst, src T) T { return hex.AppendEncode(dst, src) }

func AppendHexFromBinary(dst, src T, quote bool) (b T) {
	if quote {
		dst = AppendQuote(dst, src, AppendHex)
	} else {
		dst = hex.AppendEncode(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src T, unquote bool) (b T, err error) {
	if unquote {
		if dst, err = hex.AppendDecode(dst,
			Unquote(src)); chk.E(err) {

			return
		}
	} else {
		if dst, err = hex.AppendDecode(dst, src); chk.E(err) {
			return
		}
	}
	b = dst
	return
}

// AppendBinary is a straight append with length prefix.
func AppendBinary(dst, src T) (b T) {
	// if an allocation or two may occur, do it all in one immediately.
	minLen := len(src) + len(dst) + binary.MaxVarintLen32
	if cap(dst) < minLen {
		tmp := make(T, 0, minLen)
		dst = append(tmp, dst...)
	}
	dst = binary.AppendUvarint(dst, uint64(len(src)))
	dst = append(dst, src...)
	b = dst
	return
}

// ExtractBinary decodes the data based on the length prefix and returns a the the
// remaining data from the provided slice.
func ExtractBinary(b T) (str, rem T, err error) {
	l, read := binary.Uvarint(b)
	if read < 1 {
		err = errorf.E("failed to read uvarint length prefix")
		return
	}
	if len(b) < int(l)+read {
		err = errorf.E("insufficient data in buffer, require %d have %d",
			int(l)+read, len(b))
		return
	}
	str = b[read : read+int(l)]
	rem = b[read+int(l):]
	return
}
