package text

// NostrEscape for JSON encoding according to RFC8259.
//
// This is the efficient implementation based on the NIP-01 specification:
//
// To prevent implementation differences from creating a different event ID for the same event, the following rules MUST be followed while serializing:
//
//	No whitespace, line breaks or other unnecessary formatting should be included in the output JSON.
//	No characters except the following should be escaped, and instead should be included verbatim:
//	- A line break, 0x0A, as \n
//	- A double quote, 0x22, as \"
//	- A backslash, 0x5C, as \\
//	- A carriage return, 0x0D, as \r
//	- A tab character, 0x09, as \t
//	- A backspace, 0x08, as \b
//	- A form feed, 0x0C, as \f
//	UTF-8 should be used for encoding.
func NostrEscape(dst, src []byte) []byte {
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch {
		case c == '"':
			// quotation mark
			dst = append(dst, []byte{'\\', '"'}...)
		case c == '\\':
			// reverse solidus
			dst = append(dst, []byte{'\\', '\\'}...)
		case c == '\b':
			dst = append(dst, []byte{'\\', 'b'}...)
		case c == '\t':
			dst = append(dst, []byte{'\\', 't'}...)
		case c == '\n':
			dst = append(dst, []byte{'\\', 'n'}...)
		case c == '\f':
			dst = append(dst, []byte{'\\', 'f'}...)
		case c == '\r':
			dst = append(dst, []byte{'\\', 'r'}...)
		default:
			dst = append(dst, c)
		}
	}
	return dst
}

func NostrUnescape(dst, src []byte) []byte {
	var i int
	for ; i < len(src); i++ {
		if src[i] == '\\' {
			i++
			c := src[i]
			switch {
			case c == '"':
				dst = append(dst, '"')
			case c == '\\':
				dst = append(dst, '\\')
			case c == 'b':
				dst = append(dst, '\b')
			case c == 't':
				dst = append(dst, '\t')
			case c == 'n':
				dst = append(dst, '\n')
			case c == 'f':
				dst = append(dst, '\f')
			case c == 'r':
				dst = append(dst, '\r')
			default:
				// don't change anything that doesn't match one of the above.
				// this should leave most non-compliant escapes intact.
				dst = append(dst, '\\')
				dst = append(dst, c)
			}
		} else {
			dst = append(dst, src[i])
		}
	}
	return dst
}
