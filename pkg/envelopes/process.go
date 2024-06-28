package envelopes

type Marshaler func(dst B) (b B, err error)

type I interface {
	Label() string
	Marshal(dst B) (b B, err error)
}

func Marshal(dst B, label string, m Marshaler) (b B, err error) {
	b = dst
	b = append(b, '[', '"')
	b = append(b, label...)
	b = append(b, '"', ',')
	if b, err = m(b); chk.E(err) {
		return
	}
	b = append(b, ']')
	return
}
