package libsecp256k1

import (
	"github.com/mleku/btcec"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/btcec/secp256k1"
)

type BTCECSigner struct {
	Sec *secp256k1.SecretKey
	Pub *secp256k1.PublicKey
}

func (B *BTCECSigner) InitSec(sec B) (err error) {
	return
}

func (B *BTCECSigner) InitPub(pub B) (err error) {
	return
}

func (B *BTCECSigner) Sign(msg B) (sig B, err error) {
	return
}

func (B *BTCECSigner) Verify(msg, sig B) (valid bool, err error) {
	return
}

func (s *BTCECSigner) SecBytes() (skb B) {
	return
}

func (s *BTCECSigner) PubBytes() (pkb B) {
	return
}

func (B *BTCECSigner) Zero() {
}

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

func GenKeyPairBytes() (skb, pkb B, err error) {
	// just use the btcec key gen because the performance difference will be
	// nearly zero
	var sk *btcec.SecretKey
	if sk, err = btcec.NewSecretKey(); chk.E(err) {
		return
	}
	skb = sk.Serialize()
	pkb = schnorr.SerializePubKey(sk.PubKey())
	return
}
