package closeenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/envelopes/enveloper"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
)

const L = "CLOSE"

type T struct {
	ID *subscriptionid.T
}

var _ enveloper.I = (*T)(nil)

func New() *T                                   { return &T{ID: subscriptionid.NewStd()} }
func NewFrom(id *subscriptionid.T) *T           { return &T{ID: id} }
func (en *T) Label() string                     { return L }
func (en *T) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = en.ID.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
