//go:build !btcec

package p256k

import "C"
import (
	"crypto/rand"
	"unsafe"

	"ec.mleku.dev/v2/schnorr"
	"github.com/minio/sha256-simd"
)

/*
#cgo LDFLAGS: -lsecp256k1
#include <secp256k1.h>
#include <secp256k1_schnorrsig.h>
#include <secp256k1_extrakeys.h>
*/
import "C"

type (
	Context = C.secp256k1_context
	Uchar   = C.uchar
	SecKey  = C.secp256k1_keypair
	PubKey  = C.secp256k1_xonly_pubkey
)

var (
	ctx *Context
)

func CreateContext() *Context {
	return C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN |
		C.SECP256K1_CONTEXT_VERIFY)
}

func GetRandom() (u *Uchar) {
	rnd := make([]byte, 32)
	_, _ = rand.Read(rnd)
	return ToUchar(rnd)
}

func AssertLen(b B, length int, name string) (err error) {
	if len(b) != length {
		err = errorf.E("%s should be %d bytes, got %d", name, length, len(b))
	}
	return
}

func RandomizeContext(ctx *C.secp256k1_context) {
	C.secp256k1_context_randomize(ctx, GetRandom())
	return
}

func CreateRandomContext() (c *Context) {
	c = CreateContext()
	RandomizeContext(c)
	return
}

func init() {
	if ctx = CreateContext(); ctx == nil {
		panic("failed to create secp256k1 context")
	}
}

func ToUchar(b B) (u *Uchar) { return (*Uchar)(unsafe.Pointer(&b[0])) }

type Sec struct {
	Key SecKey
}

func GenSec() (sec *Sec, err error) {
	var skb B
	if skb, err = GenSecBytes(); chk.E(err) {
		return
	}
	return SecFromBytes(skb)
}

func SecFromBytes(sk B) (sec *Sec, err error) {
	sec = new(Sec)
	if C.secp256k1_keypair_create(ctx, &sec.Key, ToUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	return
}

func (s *Sec) Sec() *SecKey { return &s.Key }

func (s *Sec) Pub() (p *Pub, err error) {
	p = new(Pub)
	if C.secp256k1_keypair_xonly_pub(ctx, &p.Key, nil, s.Sec()) != 1 {
		err = errorf.E("pubkey derivation failed")
		return
	}
	return
}

type Pub struct {
	Key PubKey
}

func PubFromBytes(pk B) (pub *Pub, err error) {
	if err = AssertLen(pk, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = new(Pub)
	if C.secp256k1_xonly_pubkey_parse(ctx, &pub.Key, ToUchar(pk)) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pk)
		return
	}
	return
}

func (p *Pub) PubB() (b B) {
	b = make(B, schnorr.PubKeyBytesLen)
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), &p.Key)
	return
}

func (p *Pub) Pub() *PubKey { return &p.Key }

func (p *Pub) ToBytes() (b B, err error) {
	b = make(B, schnorr.PubKeyBytesLen)
	if C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), p.Pub()) != 1 {
		err = errorf.E("pubkey serialize failed")
		return
	}
	return
}

func Sign(msg *Uchar, k *SecKey) (sig B, err error) {
	sig = make(B, schnorr.SignatureSize)
	c := CreateRandomContext()
	if C.secp256k1_schnorrsig_sign32(c, ToUchar(sig), msg, k,
		GetRandom()) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

func SignFromBytes(msg, sk B) (sig B, err error) {
	var umsg *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	var sec *Sec
	if sec, err = SecFromBytes(sk); chk.E(err) {
		return
	}
	return Sign(umsg, sec.Sec())
}

func Msg(b B) (id *Uchar, err error) {
	if err = AssertLen(b, sha256.Size, "id"); chk.E(err) {
		return
	}
	id = ToUchar(b)
	return
}

func Sig(b B) (sig *Uchar, err error) {
	if err = AssertLen(b, schnorr.SignatureSize, "sig"); chk.E(err) {
		return
	}
	sig = ToUchar(b)
	return
}

func Verify(msg, sig *Uchar, pk *PubKey) (valid bool) {
	return C.secp256k1_schnorrsig_verify(ctx, sig, msg, 32, pk) == 1
}

func VerifyFromBytes(msg, sig, pk B) (err error) {
	var umsg, usig *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	if usig, err = Sig(sig); chk.E(err) {
		return
	}
	var pub *Pub
	if pub, err = PubFromBytes(pk); chk.E(err) {
		return
	}
	valid := Verify(umsg, usig, pub.Pub())
	if !valid {
		err = errorf.E("failed to verify signature")
	}
	return
}

func Zero(s *SecKey) {
	b := (*[96]byte)(unsafe.Pointer(s))[:96]
	for i := range b {
		b[i] = 0
	}
}
