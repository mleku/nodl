package closedenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
	"github.com/mleku/nodl/pkg/codec/text"
)

const L = "CLOSED"

type T struct {
	Subscription *subscriptionid.T
	Reason       B
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{Subscription: subscriptionid.NewStd()}
}

func NewFrom(id *subscriptionid.T, msg B) *T {
	return &T{Subscription: id, Reason: msg}
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
			o = text.NostrEscape(o, ce.Reason)
			o = append(o, '"')
			return
		})
	return
}

func (ce *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if ce.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = ce.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if ce.Reason, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
