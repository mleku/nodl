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

func (c Challenge) Label() string { return "AUTH" }

func (c Challenge) Marshal(dst B) (b B, err error) {
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

func UnmarshalChallenge(b B) (challenge *Challenge, rem B, err error) {
	var openQuotes bool
	rem = b
	for ; len(rem) > 0; rem = rem[1:] {
		if !openQuotes && rem[0] == '"' {
			openQuotes = true
		} else if openQuotes {
			for i := range rem {
				if rem[i] == '"' {
					challenge = &Challenge{Challenge: text.
						NostrUnescape(rem[:i])}
					// no need to read any more, any garbage after this point is
					// irrelevant
					rem = rem[:0]
					return
				}
			}
		}
	}
	return
}

type Response struct {
	Event *event.T
}

var _ envelopes.I = (*Response)(nil)

func (r Response) Label() string { return L }

func (r Response) Marshal(dst B) (b B, err error) {
	if r.Event == nil {
		err = errorf.E("nil event in response")
		return
	}
	b = dst
	b, err = envelopes.Marshal(b, L, r.Event.Marshal)
	return
}

func UnmarshalResponse(b B) (aut *Response, rem B, err error) {
	rem = b
	aut = &Response{}
	// literally just unmarshal the event
	if aut.Event, rem, err = event.Unmarshal(rem); chk.E(err) {
		return
	}
	// no need to read any more, any garbage after this point is irrelevant
	rem = rem[:0]
	return
}
