package id

import (
	"bytes"
	"fmt"
	"git.replicatr.dev/eventstore/ratel/keys"
	. "nostr.mleku.dev"
	"strings"

	"nostr.mleku.dev/codec/eventid"
	"util.mleku.dev/hex"
)

const Len = 8

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

func New(evID ...*eventid.T) (p *T) {
	if len(evID) < 1 || len(evID[0].String()) < 1 {
		return &T{make([]byte, Len)}
	}
	evid := evID[0].String()
	if len(evid) < 64 {
		evid = strings.Repeat("0", 64-len(evid)) + evid
	}
	if len(evid) > 64 {
		evid = evid[:64]
	}
	b, err := hex.Dec(evid[:Len*2])
	if Chk.E(err) {
		return
	}
	return &T{Val: b}
}

func (p *T) Write(buf *bytes.Buffer) {
	if len(p.Val) != Len {
		panic(fmt.Sprintln("must use New or initialize Val with len", Len))
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf *bytes.Buffer) (el keys.Element) {
	// allow uninitialized struct
	if len(p.Val) != Len {
		p.Val = make([]byte, Len)
	}
	if n, err := buf.Read(p.Val); Chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() int { return Len }
