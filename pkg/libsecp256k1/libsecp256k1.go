//go:build libsecp256k1

package libsecp256k1

/*
#cgo LDFLAGS: -lsecp256k1
#include <secp256k1.h>
#include <secp256k1_schnorrsig.h>
#include <secp256k1_extrakeys.h>
*/
import "C"

import (
	"crypto/rand"
	"unsafe"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
)

type (
	Context = C.secp256k1_context
	Uchar   = C.uchar
	Keypair = C.secp256k1_keypair
	PubKey  = C.secp256k1_xonly_pubkey
)

var (
	ctx *Context
)

func CreateContext() *Context {
	return C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN |
		C.SECP256K1_CONTEXT_VERIFY)
}

func init() {
	if ctx = CreateContext(); ctx == nil {
		panic("failed to create secp256k1 context")
	}
}

func getUchar(b B) (u *Uchar) { return (*Uchar)(unsafe.Pointer(&b[0])) }

type Sec struct {
	Pair Keypair
}

func GetSecFromSlice(sk B) (sec *Sec, err error) {
	sec = new(Sec)
	if C.secp256k1_keypair_create(ctx, &sec.Pair, getUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	return
}

type Pub struct {
	Key PubKey
}

func GetPubFromSlice(pk B) (pub *Pub, err error) {
	if err = AssertLen(pk, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = new(Pub)
	if C.secp256k1_xonly_pubkey_parse(ctx, &pub.Key, getUchar(pk)) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pk)
		return
	}
	return
}

func GetRandom() (u *Uchar) {
	rnd := make([]byte, 32)
	_, _ = rand.Read(rnd)
	return getUchar(rnd)
}

func AssertLen(b B, length int, name string) (err error) {
	if len(b) != length {
		err = errorf.E("%s should be %d bytes, got %d", name, length, len(b))
	}
	return
}

func RandomizeContext(ctx *C.secp256k1_context) (err error) {
	C.secp256k1_context_randomize(ctx, GetRandom())
	return
}

func sign(msg *Uchar, k *Keypair) (sig B, err error) {
	sig = make(B, schnorr.SignatureSize)
	if err = RandomizeContext(ctx); chk.E(err) {
		return
	}
	if C.secp256k1_schnorrsig_sign32(ctx, getUchar(sig), msg, k,
		GetRandom()) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

func Sign(id, sk B) (sig B, err error) {
	if err = AssertLen(id, sha256.Size, "id"); chk.E(err) {
		return
	}
	var sec *Sec
	if sec, err = GetSecFromSlice(sk); chk.E(err) {
		return
	}
	return sign(getUchar(id), &sec.Pair)
}

func verify(id, sig *Uchar, pk *PubKey) (valid bool) {
	return C.secp256k1_schnorrsig_verify(ctx, sig, id, 32, pk) == 1
}

func Verify(id, sig, pk B) (err error) {
	if err = AssertLen(id, sha256.Size, "id"); chk.E(err) {
		return
	}
	if err = AssertLen(sig, schnorr.SignatureSize, "sig"); chk.E(err) {
		return
	}
	var pub *Pub
	if pub, err = GetPubFromSlice(pk); chk.E(err) {
		err = errorf.E("failed to verify signature")
		return
	}
	valid := verify(getUchar(id), getUchar(sig), &pub.Key)
	if !valid {
		err = errorf.E("failed to verify signature")
	}
	return
}
