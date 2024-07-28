package keys

import (
	"strings"

	"ec.mleku.dev/v2"
	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/util/hex"
)

var GeneratePrivateKey = func() B { return GenerateSecretKeyHex() }

func GenerateSecretKeyHex() (sks B) {
	var err E
	var sb B
	if sb, _, err = p256k.GenSecBytes(); chk.E(err) {
		return
	}
	sks = B(hex.Enc(sb))
	return
}

func GetPublicKey(sk S) (pk S, err E) {
	if !IsValid32ByteHex(sk) {
		err = errorf.E("invalid key %s", sk)
		return
	}
	var b B
	if b, err = hex.Dec(sk); chk.E(err) {
		return
	}
	_, pkk := btcec.SecKeyFromBytes(b)
	return hex.Enc(schnorr.SerializePubKey(pkk)), nil
}

func IsValid32ByteHex(pk string) bool {
	if strings.ToLower(pk) != pk {
		return false
	}
	dec, _ := hex.Dec(pk)
	return len(dec) == 32
}
