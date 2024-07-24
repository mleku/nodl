package keys

import (
	"strings"

	"github.com/mleku/btcec"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/nodl/pkg/crypto/p256k"
	"github.com/mleku/nodl/pkg/util/hex"
)

var GeneratePrivateKey = func() string { return GenerateSecretKey() }

func GenerateSecretKey() (sks S) {
	var err E
	var sb B
	if sb, err = p256k.GenSecBytes(); chk.E(err) {
		return
	}
	sks = hex.Enc(sb)
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
