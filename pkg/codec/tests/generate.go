package tests

import (
	"encoding/base64"

	"git.replicatr.dev/pkg"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/util/hex"
	"lukechampine.com/frand"
)

func GenerateEvent(nsec B, maxSize int) (ev *event.T, binSize int, err E) {
	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &event.T{
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   event.B(base64.StdEncoding.EncodeToString(frand.Bytes(l))),
	}
	var sec B
	if _, err = hex.DecBytes(sec, nsec); chk.E(err) {
		return
	}
	var signer pkg.Signer
	if signer, err = p256k.NewSigner(&p256k.Signer{}); chk.E(err) {
		return
	}
	if err = signer.InitSec(sec); chk.E(err) {
		return
	}
	if err = ev.Sign(signer); chk.E(err) {
		return
	}
	var bin []byte
	bin, err = ev.MarshalBinary(bin)
	binSize = len(bin)
	return
}
