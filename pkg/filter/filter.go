package filter

import (
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
)

type T struct {
	IDs     tag.T        `json:"ids,omitempty"`
	Kinds   *kinds.T     `json:"kinds,omitempty"`
	Authors tag.T        `json:"authors,omitempty"`
	Tags    TagMap       `json:"-,omitempty"`
	Since   *timestamp.T `json:"since,omitempty"`
	Until   *timestamp.T `json:"until,omitempty"`
	Limit   *int         `json:"limit,omitempty"`
	Search  text.T       `json:"search,omitempty"`
}

type TagMap map[string]tag.T

func (t *T) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (t *T) UnmarshalJSON(b []byte) error {
	panic("implement me")
}

func (t *T) MarshalBinary() (data []byte, err error) {
	panic("implement me")
}

func (t *T) UnmarshalBinary(data []byte) error {
	panic("implement me")
}
