package p256k

import (
	"git.replicatr.dev/pkg"
	"github.com/mleku/btcec/v2"
)

func GenSecBytes() (skb B, err error) {
	// just use the btcec key gen because the performance difference will be
	// nearly zero
	var sk *btcec.SecretKey
	if sk, err = btcec.NewSecretKey(); chk.E(err) {
		return
	}
	skb = sk.Serialize()
	return
}

func NewSigner(s pkg.Signer) (signer pkg.Signer, err error) {
	var skb B
	if skb, err = GenSecBytes(); chk.E(err) {
		return
	}
	if err = s.InitSec(skb); chk.E(err) {
		return
	}
	signer = s
	return
}
