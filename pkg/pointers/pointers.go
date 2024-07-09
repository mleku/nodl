package pointers

import (
	"github.com/mleku/nodl/pkg/eventid"
	"github.com/mleku/nodl/pkg/kind"
)

type Profile struct {
	PublicKey B   `json:"pubkey"`
	Relays    []B `json:"relays,omitempty"`
}

type Event struct {
	ID     *eventid.T `json:"id"`
	Relays []B        `json:"relays,omitempty"`
	Author B          `json:"author,omitempty"`
	Kind   *kind.T    `json:"kind,omitempty"`
}

type Entity struct {
	PublicKey  B       `json:"pubkey"`
	Kind       *kind.T `json:"kind,omitempty"`
	Identifier B       `json:"identifier,omitempty"`
	Relays     []B     `json:"relays,omitempty"`
}
