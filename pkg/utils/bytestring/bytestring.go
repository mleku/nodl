package bytestring

import (
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
