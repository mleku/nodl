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
	"errors"
	"unsafe"
)

var globalSecp256k1Context *C.secp256k1_context

func init() {
	globalSecp256k1Context = C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN | C.SECP256K1_CONTEXT_VERIFY)
	if globalSecp256k1Context == nil {
		panic("failed to create secp256k1 context")
	}
}

func getUchar32(b [32]byte) (u *C.uchar) { return (*C.uchar)(unsafe.Pointer(&b)) }

func Sign(msg [32]byte, sk [32]byte) ([64]byte, error) {
	var sig [64]byte

	var keypair C.secp256k1_keypair
	if C.secp256k1_keypair_create(globalSecp256k1Context, &keypair,
		getUchar32(sk)) != 1 {
		return sig, errors.New("failed to parse private key")
	}

	var random [32]byte
	rand.Read(random[:])

	if C.secp256k1_schnorrsig_sign32(globalSecp256k1Context,
		(*C.uchar)(unsafe.Pointer(&sig[0])),
		getUchar32(msg), &keypair,
		(*C.uchar)(unsafe.Pointer(&random[0]))) != 1 {
		return sig, errors.New("failed to sign message")
	}

	return sig, nil
}

func verify(msg, sig, pk B) (valid bool, err error) {

	return
}

func Verify(msg [32]byte, sig [64]byte, pk [32]byte) bool {
	var xonly C.secp256k1_xonly_pubkey
	if C.secp256k1_xonly_pubkey_parse(globalSecp256k1Context, &xonly,
		getUchar32(pk)) != 1 {
		return false
	}

	return C.secp256k1_schnorrsig_verify(globalSecp256k1Context,
		(*C.uchar)(unsafe.Pointer(&sig[0])),
		(*C.uchar)(unsafe.Pointer(&msg[0])), 32, &xonly) == 1
}
