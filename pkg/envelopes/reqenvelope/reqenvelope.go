package reqenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/filters"
	"github.com/mleku/nodl/pkg/subscriptionid"
)

const L = "REQ"

type T struct {
	ID      *subscriptionid.T
	Filters *filters.T
}

func New() *T {
	return &T{ID: subscriptionid.NewStd(), Filters: filters.New()}
}

func NewFrom(id *subscriptionid.T, filters *filters.T) *T {
	return &T{ID: id, Filters: filters}
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
	if req.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = req.ID.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	req.Filters = filters.New()
	if rem, err = req.Filters.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	// expect close brackets here but actually doesn't matter if neither
	// previous blocks failed
	rem = rem[:0]
	return
}
