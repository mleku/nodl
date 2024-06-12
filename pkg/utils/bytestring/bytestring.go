package bytestring

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
