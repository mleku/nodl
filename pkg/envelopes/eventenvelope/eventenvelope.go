package eventenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/subscriptionid"
)

const L = "EVENT"

// Submission is a request for a relay to store an event.
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

func (sub *Submission) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	sub.Event = event.New()
	if rem, err = sub.Event.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
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

func (res *Result) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if res.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = res.Subscription.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	res.Event = event.New()
	if rem, err = res.Event.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
