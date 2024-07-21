package authenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/text"
)

const L = "AUTH"

type Challenge struct {
	Challenge B
}

func NewChallenge() *Challenge { return &Challenge{} }

func (c *Challenge) Label() string { return L }

func (c *Challenge) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, c.Challenge)
			o = append(o, '"')
			return
		})
	return
}

func (c *Challenge) UnmarshalJSON(b B) (r B, err error) {
	// var openQuotes bool
	r = b
	if c.Challenge, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	for ; len(r) >= 0; r = r[1:] {
		if r[0] == ']' {
			r = r[:0]
			return
		}
	}
	return
}

type Response struct {
	Event *event.T
}

func NewResponse() *Response { return &Response{} }

var _ envelopes.I = (*Response)(nil)

func (r *Response) Label() string { return L }

func (r *Response) MarshalJSON(dst B) (b B, err error) {
	if r.Event == nil {
		err = errorf.E("nil event in response")
		return
	}
	b = dst
	b, err = envelopes.Marshal(b, L, r.Event.MarshalJSON)
	return
}

func (r *Response) UnmarshalJSON(b B) (r B, err error) {
	r = b
	// literally just unmarshal the event
	r.Event = event.New()
	log.I.F("%s", r)
	if r, err = r.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	log.I.F("%s", r)
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
