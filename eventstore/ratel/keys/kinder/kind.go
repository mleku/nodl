package kinder

import (
	"bytes"
	"encoding/binary"
	"git.replicatr.dev/eventstore/ratel/keys"
	. "nostr.mleku.dev"

	"nostr.mleku.dev/codec/kind"
)

const Len = 2

type T struct {
	Val *kind.T
}

var _ keys.Element = &T{}

// New creates a new kinder.T for reading/writing kind.T values.
func New[V uint16 | int](c V) (p *T) { return &T{Val: kind.New(c)} }

func Make(c *kind.T) (v []byte) {
	v = make([]byte, Len)
	binary.BigEndian.PutUint16(v, c.K)
	return
}

func (c *T) Write(buf *bytes.Buffer) {
	buf.Write(Make(c.Val))
}

func (c *T) Read(buf *bytes.Buffer) (el keys.Element) {
	b := make([]byte, Len)
	if n, err := buf.Read(b); Chk.E(err) || n != Len {
		return nil
	}
	v := binary.BigEndian.Uint16(b)
	c.Val = kind.New(v)
	return c
}

func (c *T) Len() int { return Len }
