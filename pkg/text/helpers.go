package text

func JSONKey(dst, k B) (b B) {
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, '"', ':')
	b = dst
	return
}
