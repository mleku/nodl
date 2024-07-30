package p256k

import (
	"git.replicatr.dev/pkg"
)

func NewSigner(s pkg.Signer) (signer pkg.Signer, err error) {
	var skb B
	// todo: this really should generate the cgo structs directly
	if skb, _, _, _, err = GenSecBytes(); chk.E(err) {
		return
	}
	if err = s.InitSec(skb); chk.E(err) {
		return
	}
	signer = s
	return
}
