package noticeenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/text"
)

const L = "NOTICE"

type T struct {
	Message B
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{}
}

func NewFrom(msg B) *T {
	return &T{Message: msg}
}

func (ce *T) Label() string { return L }

func (ce *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, ce.Message)
			o = append(o, '"')
			return
		})
	return
}

func (ce *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if ce.Message, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
