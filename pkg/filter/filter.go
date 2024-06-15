package filter

import (
	"github.com/mleku/nodl/pkg/bstring"
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/timestamp"
)

type T struct {
	IDs     tag.T       `json:"ids,omitempty"`
	Kinds   kinds.T     `json:"kinds,omitempty"`
	Authors tag.T       `json:"authors,omitempty"`
	Tags    TagMap      `json:"-,omitempty"`
	Since   timestamp.T `json:"since,omitempty"`
	Until   timestamp.T `json:"until,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	Search  bstring.T   `json:"search,omitempty"`
}

type TagMap map[string]tag.T

func (t TagMap) Clone() (t1 TagMap) {
	if t == nil {
		return
	}
	t1 = make(TagMap)
	for i := range t {
		t1[i] = t[i]
	}
	return
}
