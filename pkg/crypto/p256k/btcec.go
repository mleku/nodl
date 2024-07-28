package p256k

import (
	btcec "ec.mleku.dev/v2"
	"ec.mleku.dev/v2/schnorr"
	"ec.mleku.dev/v2/secp256k1"
	"git.replicatr.dev/pkg"
)

type BTCECSigner struct {
	SecretKey *secp256k1.SecretKey
	PublicKey *secp256k1.PublicKey
	b         B
}

var _ pkg.Signer = &Signer{}

func (bs *BTCECSigner) InitSec(sec B) (err error) {
	if len(sec) != secp256k1.SecKeyBytesLen {
		err = errorf.E("sec key must be %d bytes", secp256k1.SecKeyBytesLen)
		return
	}
	bs.SecretKey = secp256k1.SecKeyFromBytes(sec)
	bs.PublicKey = bs.SecretKey.PubKey()
	bs.b = schnorr.SerializePubKey(bs.PublicKey)
	return
}

func (bs *BTCECSigner) InitPub(pub B) (err error) {
	if bs.PublicKey, err = schnorr.ParsePubKey(pub); chk.E(err) {
		return
	}
	bs.b = pub
	return
}
func (bs *BTCECSigner) Pub() (b B) { return bs.b }

func (bs *BTCECSigner) Sign(msg B) (sig B, err error) {
	if bs.SecretKey == nil {
		err = errorf.E("btcec: Signer not initialized")
		return
	}
	var s *schnorr.Signature
	if s, err = schnorr.Sign(bs.SecretKey, msg); chk.E(err) {
		return
	}
	sig = s.Serialize()
	return
}

func (bs *BTCECSigner) Verify(msg, sig B) (valid bool, err error) {
	if bs.PublicKey == nil {
		err = errorf.E("btcec: PubKey not initialized")
		return
	}
	var s *schnorr.Signature
	if s, err = schnorr.ParseSignature(sig); chk.D(err) {
		err = errorf.E("failed to parse signature:\n%d %s\n%v", len(sig),
			sig, err)
		return
	}
	valid = s.Verify(msg, bs.PublicKey)
	return
}

func (bs *BTCECSigner) Zero() { bs.SecretKey.Key.Zero() }

func (bs *BTCECSigner) ECDH(pubkeyBytes B) (secret B, err E) {
	var pub *secp256k1.PublicKey
	if pub, err = secp256k1.ParsePubKey(append(B{0x02},pubkeyBytes...)); chk.E(err) {
		return
	}
	secret = btcec.GenerateSharedSecret(bs.SecretKey, pub)
	return
}
