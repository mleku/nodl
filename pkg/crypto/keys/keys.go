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
	var skb B
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	sks = B(hex.Enc(skb))
	return
}

func GetPublicKey(sk string) (string, error) {
	b, err := hex.Dec(sk)
	if err != nil {
		return "", err
	}

	_, pk := btcec.PrivKeyFromBytes(b)
	return hex.Enc(schnorr.SerializePubKey(pk)), nil
}
//
// func GetPublicKey(sk S) (pk S, err E) {
// 	if !IsValid32ByteHex(sk) {
// 		err = errorf.E("invalid key %s", sk)
// 		return
// 	}
// 	var b B
// 	if b, err = hex.Dec(sk); chk.E(err) {
// 		return
// 	}
// 	return SecretBytesToPubKeyHex(b)
// }

func SecretBytesToPubKeyHex(skb B) (pk S, err E) {
	_, pkk := btcec.SecKeyFromBytes(skb)
	return hex.Enc(schnorr.SerializePubKey(pkk)), nil
}

func SecretToPubKeyBytes(skb B) (pk B, err E) {
	_, pkk := btcec.SecKeyFromBytes(skb)
	return schnorr.SerializePubKey(pkk), nil
}

func IsValid32ByteHex(pk string) bool {
	if strings.ToLower(pk) != pk {
		return false
	}
	dec, _ := hex.Dec(pk)
	return len(dec) == 32
}
