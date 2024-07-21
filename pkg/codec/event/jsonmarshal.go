package event

import (
	"github.com/mleku/nodl/pkg/codec/text"
	"github.com/mleku/nodl/pkg/util/hex"
)

func (ev *T) MarshalJSON(dst B) (b B, err error) {
	// open parentheses
	dst = append(dst, '{')
	// ID
	dst = text.JSONKey(dst, Id)
	dst = text.AppendQuote(dst, ev.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = text.JSONKey(dst, Pubkey)
	dst = text.AppendQuote(dst, ev.PubKey, hex.EncAppend)
	dst = append(dst, ',')
	// CreatedAt
	dst = text.JSONKey(dst, CreatedAt)
	if dst, err = ev.CreatedAt.MarshalJSON(dst); chk.E(err) {
		return
	}
	dst = append(dst, ',')
	// Kind
	dst = text.JSONKey(dst, Kind)
	dst, _ = ev.Kind.MarshalJSON(dst)
	dst = append(dst, ',')
	// Tags
	dst = text.JSONKey(dst, Tags)
	dst, _ = ev.Tags.MarshalJSON(dst)
	dst = append(dst, ',')
	// Content
	dst = text.JSONKey(dst, Content)
	dst = text.AppendQuote(dst, ev.Content, text.NostrEscape)
	dst = append(dst, ',')
	// Sig
	dst = text.JSONKey(dst, Sig)
	dst = text.AppendQuote(dst, ev.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}
