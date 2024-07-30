//go:build !btcec

package p256k

import "git.replicatr.dev/pkg"

// Signer implements the pkg.Signer interface.
//
// Either the Sec or Pub must be populated, the former is for generating
// signatures, the latter is for verifying them.
//
// When using this library only for verification, a constructor that converts
// from bytes to PubKey is needed prior to calling Verify.
type Signer struct {
	SecretKey *SecKey
	PublicKey *PubKey
	pkb, skb  B
}

var _ pkg.Signer = &Signer{}

func (s *Signer) Generate() (err E) {
	return
}

func (s *Signer) InitSec(sec B) (err error) {
	var us *Sec
	if us, err = SecFromBytes(sec); chk.E(err) {
		return
	}
	s.skb = sec
	s.SecretKey = &us.Key
	var up *Pub
	if up, err = us.Pub(); chk.E(err) {
		return
	}
	s.PublicKey = &up.Key
	s.pkb = up.PubB()
	return
}

func (s *Signer) InitPub(pub B) (err error) {
	var up *Pub
	if up, err = PubFromBytes(pub); chk.E(err) {
		return
	}
	s.PublicKey = &up.Key
	s.pkb = up.PubB()
	return
}

func (s *Signer) Sec() (b B) { return s.skb }
func (s *Signer) Pub() (b B) { return s.pkb }

func (s *Signer) Sign(msg B) (sig B, err error) {
	if s.SecretKey == nil {
		err = errorf.E("p256k: Signer secret not initialized")
		return
	}
	u := ToUchar(msg)
	if sig, err = Sign(u, s.SecretKey); chk.E(err) {
		return
	}
	return
}

func (s *Signer) Verify(msg, sig B) (valid bool, err error) {
	if s.PublicKey == nil {
		err = errorf.E("p256k: PubKey not initialized")
		return
	}
	var uMsg, uSig *Uchar
	if uMsg, err = Msg(msg); chk.E(err) {
		return
	}
	if uSig, err = Sig(sig); chk.E(err) {
		return
	}
	valid = Verify(uMsg, uSig, s.PublicKey)
	if !valid {
		err = errorf.E("p256k: invalid signature")
	}
	return
}

func (s *Signer) Zero() { Zero(s.SecretKey) }

func (s *Signer) ECDH(pp B) (secret B, err error) {
	var pub *ECPub
	if pub, err = ECPubFromSchnorrBytes(pp); chk.E(err) {
		return
	}
	return ECDH(ToUchar(s.skb), pub), nil
}
