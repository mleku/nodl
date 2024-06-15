package bstring

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

func AppendList(dst T, src []T, separator byte,
	ac AppendBytesClosure) T {
	last := len(src) - 1
	for i := range src {
		dst = append(dst, ac(dst, src[i])...)
		if i < last {
			dst = append(dst, separator)
		}
	}
	return dst
}
