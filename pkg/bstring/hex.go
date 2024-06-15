package bstring

import (
	"encoding/hex"
)

func AppendHexEncode(dst, src T) T {
	return hex.AppendEncode(dst,
		src)
}

func AppendHexFromBinary(dst, src T, quote bool) (b T) {
	if quote {
		dst = AppendQuote(dst, src, AppendHexEncode)
	} else {
		dst = AppendHexEncode(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src T, unquote bool) (b T,
	err error) {
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
