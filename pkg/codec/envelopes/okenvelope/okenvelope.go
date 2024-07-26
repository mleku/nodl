package okenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/envelopes/enveloper"
	"github.com/mleku/nodl/pkg/codec/eventid"
	"github.com/mleku/nodl/pkg/codec/text"
)

const (
	L = "OK"
)

type T struct {
	EventID *eventid.T
	OK      bool
	Reason  B
}

var _ enveloper.I = (*T)(nil)

func New() *T                                   { return &T{} }
func NewFrom(eid *eventid.T, ok bool, msg B) *T { return &T{EventID: eid, OK: ok, Reason: msg} }
func (en *T) Label() string                     { return L }
func (en *T) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = en.EventID.ByteString(o)
			o = append(o, '"')
			o = append(o, ',')
			o = text.MarshalBool(o, en.OK)
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, en.Reason)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	var idHex B
	if idHex, r, err = text.UnmarshalHex(r); chk.E(err) {
		return
	}
	if en.EventID, err = eventid.NewFromBytes(idHex); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	if r, en.OK, err = text.UnmarshalBool(r); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	if en.Reason, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
