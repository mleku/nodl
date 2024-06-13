package bytestring

import (
	"encoding/binary"
	"encoding/hex"
	"os"

	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type AppendBytesClosure func(dst, src []byte) []byte

type AppendClosure func(dst []byte) []byte

func Unquote(b []byte) []byte { return b[1 : len(b)-1] }

func AppendQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '"')
	dst = ac(dst, src)
	dst = append(dst, '"')
	return dst
}

func AppendSingleQuote(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '\'')
	dst = ac(dst, src)
	dst = append(dst, '\'')
	return dst
}

func AppendBackticks(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '`')
	dst = ac(dst, src)
	dst = append(dst, '`')
	return dst
}

func AppendBrace(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '(')
	dst = ac(dst, src)
	dst = append(dst, ')')
	return dst
}

func AppendParenthesis(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '{')
	dst = ac(dst, src)
	dst = append(dst, '}')
	return dst
}

func AppendBracket(dst, src []byte, ac AppendBytesClosure) []byte {
	dst = append(dst, '[')
	dst = ac(dst, src)
	dst = append(dst, ']')
	return dst
}

func AppendList(dst []byte, separator byte, acs ...AppendClosure) []byte {
	last := len(acs) - 1
	for i := range acs {
		dst = acs[i](dst)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}

func AppendHexFromBinary(dst, src []byte, quote bool) (b []byte) {
	if quote {
		dst = AppendQuote(dst, src, hex.AppendEncode)
	} else {
		dst = hex.AppendEncode(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src []byte, unquote bool) (b []byte, err error) {
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

// Append is a straight append with length prefix.
func Append(dst, src []byte) (b []byte) {
	// if an allocation or two may occur, do it all in one immediately.
	minLen := len(src) + len(dst) + binary.MaxVarintLen32
	if cap(dst) < minLen {
		tmp := make([]byte, 0, minLen)
		dst = append(tmp, dst...)
	}
	dst = binary.AppendUvarint(dst, uint64(len(src)))
	dst = append(dst, src...)
	b = dst
	return
}

func Extract(b []byte) (str, rem []byte, err error) {
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
