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

var ctx *C.secp256k1_context

func init() {
	ctx = C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN | C.SECP256K1_CONTEXT_VERIFY)
	if ctx == nil {
		panic("failed to create secp256k1 context")
	}
}

func getUchar(b B) (u *C.uchar) { return (*C.uchar)(unsafe.Pointer(&b[0])) }

type Pub struct {
	Key C.secp256k1_xonly_pubkey
}

func GetPubFromSlice(pk B) (pub *Pub, err error) {
	pub = new(Pub)
	if C.secp256k1_xonly_pubkey_parse(ctx, &pub.Key, getUchar(pk)) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pk)
		return
	}
	return
}

func Sign(msg, sk B) (sig B, err error) {
	var keypair C.secp256k1_keypair
	if C.secp256k1_keypair_create(ctx, &keypair, getUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	random := make([]byte, 32)
	if _, err = rand.Read(random); chk.E(err) {
		return
	}
	sig = make(B, schnorr.SignatureSize)
	if C.secp256k1_schnorrsig_sign32(ctx,
		getUchar(sig), getUchar(msg), &keypair, getUchar(random)) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

func Verify(id, sig, pk B) (valid bool, err error) {
	if len(id) != sha256.Size {
		err = errorf.E("id should be 32 bytes, got %d", len(id))
		return
	}
	if len(sig) != 64 {
		err = errorf.E("sig should be 64 bytes, got %d", len(sig))
		return
	}
	if len(pk) != schnorr.PubKeyBytesLen {
		err = errorf.E("pubkey should be 32 bytes got %d", len(pk))
		return
	}
	var pub *Pub
	if pub, err = GetPubFromSlice(pk); chk.E(err) {
		err = errorf.E("failed to verify signature")
		return
	}
	valid = C.secp256k1_schnorrsig_verify(ctx,
		getUchar(sig), getUchar(id), 32, &pub.Key) == 1
	if !valid {
		err = errorf.E("failed to verify signature")
	}
	return
}
