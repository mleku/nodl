package pubkey

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mleku/nodl/pkg/utils/ec"
	"github.com/mleku/nodl/pkg/utils/ec/schnorr"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type T struct {
	b []byte
}

func New() *T {
	return &T{b: make([]byte, 0, schnorr.PubKeyBytesLen)}
}
func (t *T) Bytes() []byte    { return t.b }
func (t *T) String() string   { return hex.EncodeToString(t.Bytes()) }
func (t *T) Reset()           { t.b = t.b[:0] }
func (t *T) Equal(t2 *T) bool { return bytes.Equal(t.b, t2.b) }

// Valid parses the encoded bytes to ensure they are valid. If the pubkey needs
// to actually be used, don't use this, instead convert it to an ec.PublicKey
// using ToPubkey which actually returns the value.
func (t *T) Valid() (err error) {
	_, err = schnorr.ParsePubKey(t.b)
	return
}

func NewFromPubKey(pk *ec.PublicKey) (t *T, err error) {
	t = &T{b: schnorr.SerializePubKey(pk)}
	return
}

func (t *T) ToPubkey() (pk *ec.PublicKey, err error) {
	if pk, err = schnorr.ParsePubKey(t.b); chk.E(err) {
		// this shouldn't happen
		return
	}
	return
}

func NewFromBytes(b []byte) (t *T, err error) {
	if len(b) != schnorr.PubKeyBytesLen {
		err = fmt.Errorf("pubkey.NewFromBytes: invalid length %d require %d",
			len(b), schnorr.PubKeyBytesLen)
	}
	t = &T{b: b}
	if err = t.Valid(); chk.E(err) {
		err = errorf.E("pubkey.NewFromBytes: invalid pubkey '%s'", err.Error())
	}
	return
}

func (t *T) Set(b []byte) (err error) {
	if len(b) != schnorr.PubKeyBytesLen {
		err = fmt.Errorf("eventid.NewFromBytes: invalid length %d require %d",
			len(b), schnorr.PubKeyBytesLen)
	}
	t.b = b
	return
}
