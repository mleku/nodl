package envelopes

import "github.com/mleku/nodl/pkg/codec"

type Marshaler func(dst B) (b B, err error)

type I interface {
	Label() string
	codec.JSON
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
