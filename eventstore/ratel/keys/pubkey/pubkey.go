package pubkey

import (
	"bytes"
	"fmt"

	"git.replicatr.dev/eventstore/ratel/keys"
	. "nostr.mleku.dev"

	"ec.mleku.dev/v2/schnorr"
)

const Len = 8

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

// New creates a new pubkey prefix, if parameter is omitted, new one is
// allocated (for read) if more than one is given, only the first is used, and
// if the first one is not the correct hexadecimal length of 64, return error.
func New(pk ...B) (p *T, err E) {
	if len(pk) < 1 {
		// allows init with no parameter
		return &T{make([]byte, Len)}, nil
	}
	// // only the first pubkey will be used
	// if len(pk[0]) != schnorr.PubKeyBytesLen*2 {
	// 	err = Log.E.Err("pubkey hex must be 64 chars, got", len(pk[0]))
	// 	return
	// }
	// b, err := hex.Dec(pk[0][:Len*2])
	// if Chk.E(err) {
	// 	return
	// }
	return &T{Val: pk[0][:Len]}, nil
}

func NewFromBytes(pkb []byte) (p *T, err error) {
	if len(pkb) != schnorr.PubKeyBytesLen {
		err = Log.E.Err("provided key not correct length, got %d expected %d",
			len(pkb), schnorr.PubKeyBytesLen)
		Log.T.S(pkb)
		return
	}
	b := make([]byte, Len)
	copy(b, pkb[:Len])
	p = &T{Val: b}
	return
}

func (p *T) Write(buf *bytes.Buffer) {
	if p == nil {
		panic("nil pubkey")
	}
	if p.Val == nil || len(p.Val) != Len {
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
