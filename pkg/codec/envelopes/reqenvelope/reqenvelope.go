package reqenvelope

import (
	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/enveloper"
	"git.replicatr.dev/pkg/codec/filters"
	sid "git.replicatr.dev/pkg/codec/subscriptionid"
	"git.replicatr.dev/pkg/codec/text"
	"git.replicatr.dev/pkg/protocol/relayws"
)

const L = "REQ"

type T struct {
	Subscription *sid.T
	Filters      *filters.T
}

var _ enveloper.I = (*T)(nil)

func New() *T {
	return &T{Subscription: sid.NewStd(),
		Filters: filters.New()}
}
func NewFrom(id *sid.T, filters *filters.T) *T { return &T{Subscription: id, Filters: filters} }
func (en *T) Label() string                    { return L }
func (en *T) Write(ws *relayws.WS) (err E)     { 	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	return ws.WriteTextMessage(b)
}

func (en *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = en.Filters.MarshalJSON(o); chk.E(err) {
				return
			}
			log.I.S(en.Filters)
			return
		})
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
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	en.Filters = filters.New()
	if r, err = en.Filters.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
