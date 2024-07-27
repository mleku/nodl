package keys

import (
	"strings"

	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/util/hex"
	"github.com/mleku/btcec"
	"github.com/mleku/btcec/schnorr"
)

var GeneratePrivateKey = func() B { return GenerateSecretKey() }

func GenerateSecretKey() (sks B) {
	var err E
	var sb B
	if sb, err = p256k.GenSecBytes(); chk.E(err) {
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
