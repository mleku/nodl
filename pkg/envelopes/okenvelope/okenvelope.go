package okenvelope

import (
	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/eventid"
	"github.com/mleku/nodl/pkg/text"
)

const (
	L = "OK"
)

type T struct {
	EventID *eventid.T
	OK      bool
	Message B
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{}
}

func NewFrom(eid *eventid.T, ok bool, msg B) *T {
	return &T{EventID: eid, OK: ok, Message: msg}
}

func (oe *T) Label() string { return L }

func (oe *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = oe.EventID.ByteString(o)
			o = append(o, '"')
			o = append(o, ',')
			o = text.MarshalBool(o, oe.OK)
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, oe.Message)
			o = append(o, '"')
			return
		})
	return
}

func (oe *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	var idHex B
	if idHex, rem, err = text.UnmarshalHex(rem); chk.E(err) {
		return
	}
	if oe.EventID, err = eventid.NewFromBytes(idHex); chk.E(err) {
		return
	}
	if rem, err = text.Comma(rem); chk.E(err) {
		return
	}
	if rem, oe.OK, err = text.UnmarshalBool(rem); chk.E(err) {
		return
	}
	if rem, err = text.Comma(rem); chk.E(err) {
		return
	}
	if oe.Message, rem, err = text.UnmarshalQuoted(rem); chk.E(err) {
		return
	}
	if rem, err = envelopes.SkipToTheEnd(rem); chk.E(err) {
		return
	}
	return
}
