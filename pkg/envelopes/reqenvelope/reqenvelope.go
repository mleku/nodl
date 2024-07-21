package reqenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/filters"
	"github.com/mleku/nodl/pkg/subscriptionid"
	"github.com/mleku/nodl/pkg/text"
)

const L = "REQ"

type T struct {
	Subscription *subscriptionid.T
	Filters      *filters.T
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{Subscription: subscriptionid.NewStd(), Filters: filters.New()}
}

func NewFrom(id *subscriptionid.T, filters *filters.T) *T {
	return &T{Subscription: id, Filters: filters}
}

func (req *T) Label() string { return L }

func (req *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = req.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = req.Filters.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (req *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if req.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = req.Subscription.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if rem, err = text.Comma(rem); chk.E(err) {
		return
	}
	req.Filters = filters.New()
	if rem, err = req.Filters.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
