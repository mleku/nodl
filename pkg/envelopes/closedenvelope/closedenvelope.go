package closedenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/subscriptionid"
	"github.com/mleku/nodl/pkg/text"
)

const L = "CLOSED"

type T struct {
	ID      *subscriptionid.T
	Message B
}

func New() *T {
	return &T{ID: subscriptionid.NewStd()}
}

func NewFrom(id *subscriptionid.T, msg B) *T {
	return &T{ID: id, Message: msg}
}

func (ce *T) Label() string { return L }

func (ce *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = ce.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, ce.Message)
			o = append(o, '"')
			return
		})
	return
}

func (ce *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if ce.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = ce.ID.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if ce.Message, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
		return
	}
	for ; len(rem) >= 0; rem = rem[1:] {
		if rem[0] == ']' {
			rem = rem[:0]
			return
		}
	}
	return
}
