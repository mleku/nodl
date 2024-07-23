package eventenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/subscriptionid"
)

const L = "EVENT"

// Submission is a request from a client for a relay to store an event.
type Submission struct {
	Event *event.T
}

var _ envelopes.I = (*Submission)(nil)

func NewSubmission() *Submission                { return &Submission{Event: &event.T{}} }
func NewSubmissionWith(ev *event.T) *Submission { return &Submission{Event: ev} }
func (sub *Submission) Label() string           { return L }

func (sub *Submission) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = sub.Event.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (sub *Submission) UnmarshalJSON(b B) (r B, err error) {
	r = b
	sub.Event = event.New()
	if r, err = sub.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Result is an event matching a filter associated with a subscription.
type Result struct {
	Subscription *subscriptionid.T
	Event        *event.T
}

var _ envelopes.I = (*Result)(nil)

func NewResult() *Result { return &Result{} }
func NewResultWith(s *subscriptionid.T, ev *event.T) *Result {
	return &Result{Subscription: s, Event: ev}
}
func (res *Result) Label() string { return L }

func (res *Result) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = res.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = res.Event.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (res *Result) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if res.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = res.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	res.Event = event.New()
	if r, err = res.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
