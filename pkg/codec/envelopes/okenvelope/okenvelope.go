package okenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/eventid"
	"github.com/mleku/nodl/pkg/codec/text"
)

const (
	L = "OK"
)

type Reason string

func (r Reason) S() string { return string(r) }

const (
	PoW         Reason = "pow"
	Duplicate   Reason = "duplicate"
	Blocked     Reason = "blocked"
	RateLimited Reason = "rate-limited"
	Invalid     Reason = "invalid"
	Error       Reason = "error"
)

type T struct {
	EventID *eventid.T
	OK      bool
	Reason  B
}

var _ envelopes.I = (*T)(nil)

func New() *T {
	return &T{}
}

func NewFrom(eid *eventid.T, ok bool, msg B) *T {
	return &T{EventID: eid, OK: ok, Reason: msg}
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
			o = text.NostrEscape(o, oe.Reason)
			o = append(o, '"')
			return
		})
	return
}

func (oe *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	var idHex B
	if idHex, r, err = text.UnmarshalHex(r); chk.E(err) {
		return
	}
	if oe.EventID, err = eventid.NewFromBytes(idHex); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	if r, oe.OK, err = text.UnmarshalBool(r); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	if oe.Reason, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
