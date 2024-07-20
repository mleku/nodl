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

func (c *Challenge) Label() string { return "AUTH" }

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

func (c *Challenge) UnmarshalJSON(b B) (rem B, err error) {
	// var openQuotes bool
	rem = b
	if c.Challenge, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
		return
	}
	for ; len(rem) >= 0; rem = rem[1:] {
		if rem[0] == ']' {
			rem = rem[:0]
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

func (r *Response) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	// literally just unmarshal the event
	r.Event = event.New()
	if rem, err = r.Event.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	// no need to read any more, any garbage after this point is irrelevant
	rem = rem[:0]
	return
}
