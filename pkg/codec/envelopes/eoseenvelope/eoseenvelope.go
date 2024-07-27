package eoseenvelope

import (
	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/enveloper"
	sid "git.replicatr.dev/pkg/codec/subscriptionid"
)

const L = "EOSE"

type T struct {
	Subscription *sid.T
}

var _ enveloper.I = (*T)(nil)

func New() *T                                   { return &T{Subscription: sid.NewStd()} }
func NewFrom(id *sid.T) *T                      { return &T{Subscription: id} }
func (en *T) Label() string                     { return L }
func (en *T) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		},
	)
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.Subscription, err = sid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
