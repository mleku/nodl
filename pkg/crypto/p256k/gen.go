package p256k

import (
	btcec "ec.mleku.dev/v2"
	"git.replicatr.dev/pkg"
)

func GenSecBytes() (skb, pkb B, err error) {
	// just use the btcec key gen because the performance difference will be nearly zero... just make sure we don't make
	// odd pub keys (BIP-340 doesn't allow them)
	for {
		var sk *btcec.SecretKey
		if sk, err = btcec.NewSecretKey(); chk.E(err) {
			return
		}
		pkb = sk.PubKey().SerializeCompressed()
		if pkb[0] == 2 {
			pkb = pkb[1:]
			skb = sk.Serialize()
			break
		}
	}
	return
}

func NewSigner(s pkg.Signer) (signer pkg.Signer, err error) {
	var skb B
	if skb, _, err = GenSecBytes(); chk.E(err) {
		return
	}
	// log.I.S(skb)
	if err = s.InitSec(skb); chk.E(err) {
		return
	}
	signer = s
	return
}
