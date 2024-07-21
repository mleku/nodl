package closedenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/subscriptionid"
	"github.com/mleku/nodl/pkg/text"
)

const L = "CLOSED"

type T struct {
	Subscription *subscriptionid.T
	Message      B
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{Subscription: subscriptionid.NewStd()}
}

func NewFrom(id *subscriptionid.T, msg B) *T {
	return &T{Subscription: id, Message: msg}
}

func (ce *T) Label() string { return L }

func (ce *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = ce.Subscription.MarshalJSON(o); chk.E(err) {
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
	if ce.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = ce.Subscription.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if ce.Message, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
