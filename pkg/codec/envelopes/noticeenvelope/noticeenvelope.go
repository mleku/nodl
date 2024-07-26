package noticeenvelope

import (
	"github.com/mleku/nodl/pkg/codec/envelopes"
	"github.com/mleku/nodl/pkg/codec/envelopes/enveloper"
	"github.com/mleku/nodl/pkg/codec/text"
)

const L = "NOTICE"

type T struct {
	Message B
}

var _ enveloper.I = (*T)(nil)

func New() *T                                   { return &T{} }
func NewFrom[V S | B](msg V) *T                 { return &T{Message: B(msg)} }
func (en *T) Label() string                     { return L }
func (en *T) Write(ws enveloper.Writer) (err E) { return ws.WriteEnvelope(en) }

func (en *T) MarshalJSON(dst B) (b B, err E) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Message)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err E) {
	r = b
	if en.Message, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}
