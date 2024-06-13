package pubkey

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mleku/nodl/pkg/utils/bytestring"
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

func AppendHexFromBinary(dst, src []byte, quote bool) (b []byte) {
	if quote {
		dst = bytestring.AppendQuote(dst, src, hex.AppendEncode)
	} else {
		dst = hex.AppendEncode(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src []byte, unquote bool) (b []byte, err error) {
	if unquote {
		if dst, err = hex.AppendDecode(dst,
			bytestring.Unquote(src)); chk.E(err) {

			return
		}
	} else {
		if dst, err = hex.AppendDecode(dst, src); chk.E(err) {
			return
		}
	}
	b = dst
	return
}

func (t *T) MarshalJSON() (b []byte, err error) {
	b = make([]byte, 0, schnorr.PubKeyBytesLen*2+2)
	b = AppendHexFromBinary(b, t.b, true)
	return
}

func (t *T) UnmarshalJSON(b []byte) (err error) {
	if len(b) < schnorr.PubKeyBytesLen*2+2 {
		return errorf.E("pubkey: not enough bytes got %d required %d",
			len(b), schnorr.PubKeyBytesLen*2+2)
	}
	// reset the slice
	t.Reset()
	if t.b, err = AppendBinaryFromHex(t.b, b, true); chk.E(err) {
		return
	}
	return
}

func (t *T) MarshalBinary() (data []byte, err error) { return t.b, nil }

func (t *T) UnmarshalBinary(data []byte) (err error) {
	if len(data) < schnorr.PubKeyBytesLen {
		return errorf.E("pubkey: not enough bytes got %d required %d",
			len(data), schnorr.PubKeyBytesLen)
	}
	t.b = data
	return
}
