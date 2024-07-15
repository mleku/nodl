package libsecp256k1

import (
	"fmt"

	"github.com/mleku/btcec"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/btcec/secp256k1"
	"github.com/mleku/nodl/pkg"
)

type BTCECSigner struct {
	Sec *secp256k1.SecretKey
	Pub *secp256k1.PublicKey
	b   B
}

var _ pkg.Signer = &Signer{}

func (bs *BTCECSigner) InitSec(sec B) (err error) {
	if len(sec) != secp256k1.SecKeyBytesLen {
		return fmt.Errorf("sec key must be %d bytes", secp256k1.SecKeyBytesLen)
	}
	bs.Sec = secp256k1.SecKeyFromBytes(sec)
	bs.Pub = bs.Sec.PubKey()
	bs.b = schnorr.SerializePubKey(bs.Pub)
	return
}

func (bs *BTCECSigner) InitPub(pub B) (err error) {
	if bs.Pub, err = schnorr.ParsePubKey(pub); chk.E(err) {
		return
	}
	bs.b = pub
	return
}
func (bs *BTCECSigner) PubB() (b B) { return bs.b }

func (bs *BTCECSigner) Sign(msg B) (sig B, err error) {
	if bs.Sec == nil {
		err = errorf.E("btcec: Signer not initialized")
		return
	}
	var s *schnorr.Signature
	if s, err = schnorr.Sign(bs.Sec, msg); chk.E(err) {
		return
	}
	sig = s.Serialize()
	return
}

func (bs *BTCECSigner) Verify(msg, sig B) (valid bool, err error) {
	var s *schnorr.Signature
	if s, err = schnorr.ParseSignature(sig); chk.D(err) {
		err = errorf.E("failed to parse signature:\n%d %s\n%v", len(sig),
			sig, err)
		return
	}
	valid = s.Verify(msg, bs.Pub)
	return
}

func (bs *BTCECSigner) Zero() { bs.Sec.Key.Zero() }

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
