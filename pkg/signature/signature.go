package signature

import (
	"bytes"
	"encoding/hex"
	"errors"
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
	return &T{b: make([]byte, 0, schnorr.SignatureSize)}
}
func (t *T) Bytes() []byte    { return t.b }
func (t *T) String() string   { return hex.EncodeToString(t.Bytes()) }
func (t *T) Reset()           { t.b = t.b[:0] }
func (t *T) Equal(t2 *T) bool { return bytes.Equal(t.b, t2.b) }

// Valid parses the encoded bytes to ensure they are valid. If the signature
// needs to actually be used, don't use this, instead convert it to a
// schnorr.Signature using ToSignature which actually returns the value.
func (t *T) Valid() (err error) {
	_, err = schnorr.ParseSignature(t.b)
	return
}

func Sign(sec *ec.SecretKey, hash []byte) (sig *T, err error) {
	var s *schnorr.Signature
	if s, err = schnorr.Sign(sec, hash); chk.E(err) {
		return
	}
	return NewFromSignature(s)
}

func (t *T) Verify(hash []byte, pk *ec.PublicKey) (err error) {
	var s *schnorr.Signature
	if s, err = t.ToSignature(); chk.E(err) {
		return
	}
	if !s.Verify(hash, pk) {
		err = errors.New("signature verification failed")
	}
	return
}

func NewFromSignature(sig *schnorr.Signature) (t *T, err error) {
	t = &T{b: sig.Serialize()}
	return
}

func (t *T) ToSignature() (sig *schnorr.Signature, err error) {
	if sig, err = schnorr.ParseSignature(t.b); chk.E(err) {
		// this shouldn't happen
		return
	}
	return
}

func NewFromBytes(b []byte) (t *T, err error) {
	if len(b) != schnorr.SignatureSize {
		err = fmt.Errorf("signature.NewFromBytes: invalid length %d require %d",
			len(b), schnorr.SignatureSize)
	}
	t = &T{b: b}
	if err = t.Valid(); chk.E(err) {
		err = errorf.E("signature.NewFromBytes: invalid signature '%s'",
			err.Error())
	}
	return
}

func (t *T) Set(b []byte) (err error) {
	if len(b) != schnorr.SignatureSize {
		err = fmt.Errorf("signature.NewFromBytes: invalid length %d require %d",
			len(b), schnorr.SignatureSize)
	}
	t.b = b
	return
}
