package agent

import (
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filter"
	"nostr.mleku.dev/codec/kind"
	"nostr.mleku.dev/codec/tag"
	"nostr.mleku.dev/codec/tags"
	"nostr.mleku.dev/codec/timestamp"
	"util.mleku.dev/hex"
)

func TagsToKV(t *tags.T) (keys []KeyValuePair) {
	keys = make([]KeyValuePair, 0, len(t.T))
	for k := range t.T {
		keys = append(keys, KeyValuePair{S(t.T[k].Field[0]), t.T[k].ToStringSlice()})
	}
	return
}

func FilterToCandid(f *filter.T) (result *Filter) {
	result = &Filter{
		IDs:     f.IDs.ToStringSlice(),
		Kinds:   f.Kinds.ToUint16(),
		Authors: f.Authors.ToStringSlice(),
		Tags:    TagsToKV(f.Tags),
		Search:  S(f.Search),
	}
	if f.Since != nil {
		result.Since = f.Since.I64()
	} else {
		result.Since = -1
	}

	if f.Until != nil {
		result.Until = f.Until.I64()
	} else {
		result.Until = -1
	}

	if f.Limit > 0 {
		result.Limit = int64(f.Limit)
	} else {
		result.Limit = 500
	}

	return
}

func EventToCandid(e *event.T) Event {
	return Event{
		e.IDString(),
		e.PubKeyString(),
		e.CreatedAt.I64(),
		e.Kind.ToU16(),
		e.TagStrings(),
		e.ContentString(),
		e.SigString(),
	}
}

func (e *Event) ToEvent() (ev *event.T) {
	t := &tags.T{}
	for _, v := range e.Tags {
		var bb []B
		for _, w := range v {
			bb = append(bb, B(w))
		}
		t.T = append(t.T, tag.New(bb...))
	}
	var err E
	var id B
	if id, err = hex.Dec(e.ID); chk.E(err) {
		return
	}
	var pk B
	if pk, err = hex.Dec(e.Pubkey); chk.E(err) {
		return
	}
	var sig B
	if sig, err = hex.Dec(e.Sig); chk.E(err) {
		return
	}
	return &event.T{
		ID:        id,
		PubKey:    pk,
		CreatedAt: timestamp.FromUnix(e.CreatedAt),
		Kind:      kind.New(e.Kind),
		Tags:      t,
		Content:   B(e.Content),
		Sig:       sig,
	}
}
