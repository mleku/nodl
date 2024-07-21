package eoseenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/subscriptionid"
)

const L = "EOSE"

type T struct {
	ID *subscriptionid.T
}
var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{ID: subscriptionid.NewStd()}
}

func NewFrom(id *subscriptionid.T) *T {
	return &T{ID: id}
}

func (req *T) Label() string { return L }

func (req *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = req.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (req *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if req.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = req.ID.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
