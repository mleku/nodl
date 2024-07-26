package authenvelope

import (
	envs "github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/envelopes/interface"
	"github.com/mleku/nodl/pkg/codec/event"
	"github.com/mleku/nodl/pkg/codec/text"
)

const L = "AUTH"

type Challenge struct {
	Challenge B
}

func NewChallenge() *Challenge                     { return &Challenge{} }
func NewChallengeWith(challenge B) *Challenge      { return &Challenge{Challenge: challenge} }
func (en *Challenge) Label() string                { return L }
func (en *Challenge) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *Challenge) MarshalJSON(dst B) (b B, err E) {
	b = dst
	b, err = envs.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Challenge)
			o = append(o, '"')
			return
		})
	return
}

func (en *Challenge) UnmarshalJSON(b B) (r B, err E) {
	r = b
	if en.Challenge, r, err = text.UnmarshalQuoted(r); chk.E(err) {
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

var _ enveloper.I = (*Response)(nil)

func NewResponse() *Response                      { return &Response{} }
func (en *Response) Label() string                { return L }
func (en *Response) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *Response) MarshalJSON(dst B) (b B, err E) {
	if en == nil {
		err = errorf.E("nil response")
		return
	}
	if en.Event == nil {
		err = errorf.E("nil event in response")
		return
	}
	b = dst
	b, err = envs.Marshal(b, L, en.Event.MarshalJSON)
	return
}

func (en *Response) UnmarshalJSON(b B) (r B, err E) {
	r = b
	// literally just unmarshal the event
	en.Event = event.New()
	if r, err = en.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envs.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
