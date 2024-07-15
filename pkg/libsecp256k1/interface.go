//go:build !btcec

package libsecp256k1

import "github.com/mleku/nodl/pkg"

// Signer implements the pkg.Signer interface.
//
// Either the Sec or Pub must be populated, the former is for generating
// signatures, the latter is for verifying them.
//
// When using this library only for verification, a constructor that converts
// from bytes to PubKey is needed prior to calling Verify.
type Signer struct {
	Sec *SecKey
	Pub *PubKey
	b   B
}

var _ pkg.Signer = &Signer{}

func (s *Signer) InitSec(sec B) (err error) {
	var us *Sec
	if us, err = SecFromBytes(sec); chk.E(err) {
		return
	}
	s.Sec = &us.Key
	var up *Pub
	if up, err = us.Pub(); chk.E(err) {
		return
	}
	s.Pub = &up.Key
	s.b = up.PubB()
	return
}

func (s *Signer) InitPub(pub B) (err error) {
	var up *Pub
	if up, err = PubFromBytes(pub); chk.E(err) {
		return
	}
	s.Pub = &up.Key
	s.b = up.PubB()
	return
}

func (s *Signer) PubB() (b B) {
	return s.b
}

func (s *Signer) Sign(msg B) (sig B, err error) {
	if s.Sec == nil {
		err = errorf.E("libsecp256k1: Signer not initialized")
		return
	}
	u := ToUchar(msg)
	if sig, err = Sign(u, s.Sec); chk.E(err) {
		return
	}
	return
}

func (s *Signer) Verify(msg, sig B) (valid bool, err error) {
	if s.Pub == nil {
		err = errorf.E("libsecp256k1: PubKey not initialized")
		return
	}
	var uMsg, uSig *Uchar
	if uMsg, err = Msg(msg); chk.E(err) {
		return
	}
	if uSig, err = Sig(sig); chk.E(err) {
		return
	}
	valid = Verify(uMsg, uSig, s.Pub)
	if !valid {
		err = errorf.E("libsecp256k1: invalid signature")
	}
	return
}

func (s *Signer) Zero() { Zero(s.Sec) }
