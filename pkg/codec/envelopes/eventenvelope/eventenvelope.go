package eventenvelope

import (
	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/enveloper"
	"git.replicatr.dev/pkg/codec/event"
	sid "git.replicatr.dev/pkg/codec/subscriptionid"
)

const L = "EVENT"

// Submission is a request from a client for a relay to store an event.
type Submission struct {
	*event.T
}

var _ enveloper.I = (*Submission)(nil)

func NewSubmission() *Submission                         { return &Submission{T: &event.T{}} }
func NewSubmissionWith(ev *event.T) *Submission          { return &Submission{T: ev} }
func (en *Submission) Label() string                     { return L }
func (en *Submission) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *Submission) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.T.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Submission) UnmarshalJSON(b B) (r B, err error) {
	r = b
	en.T = event.New()
	if r, err = en.T.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Result is an event matching a filter associated with a subscription.
type Result struct {
	Subscription *sid.T
	Event        *event.T
}

var _ enveloper.I = (*Result)(nil)

func NewResult() *Result { return &Result{} }
func NewResultWith(s *sid.T, ev *event.T) *Result {
	return &Result{Subscription: s,
		Event: ev}
}
func (en *Result) Label() S                          { return L }
func (en *Result) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *Result) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = en.Event.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Result) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.Subscription, err = sid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	en.Event = event.New()
	if r, err = en.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
