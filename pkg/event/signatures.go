package event

import (
	sch "github.com/mleku/btcec/schnorr"
	k1 "github.com/mleku/btcec/secp256k1"
)

// SignWithSecKey signs an event with a given *secp256xk1.SecretKey.
func (ev *T) SignWithSecKey(sk *k1.SecretKey,
	so ...sch.SignOption) (err error) {

	// sign the event.
	var sig *sch.Signature
	ev.ID = ev.GetIDBytes()
	if sig, err = sch.Sign(sk, ev.ID, so...); chk.D(err) {
		return
	}
	// we know secret key is good so we can generate the public key.
	ev.PubKey = sch.SerializePubKey(sk.PubKey())
	ev.Sig = sig.Serialize()
	return
}

func (ev *T) CheckSignature() (valid bool, err error) {
	// parse pubkey bytes.
	var pk *k1.PublicKey
	if pk, err = sch.ParsePubKey(ev.PubKey); chk.D(err) {
		err = errorf.E("event has invalid pubkey '%0x': %v", ev.PubKey, err)
		return
	}
	// parse signature bytes.
	var sig *sch.Signature
	if sig, err = sch.ParseSignature(ev.Sig); chk.D(err) {
		err = errorf.E("failed to parse signature:\n%d %s\n%v", len(ev.Sig),
			ev.Sig, err)
		return
	}
	// check signature.
	valid = sig.Verify(ev.GetIDBytes(), pk)
	return
}