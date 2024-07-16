package auth

import (
	"testing"

	"github.com/mleku/nodl/pkg"
	"github.com/mleku/nodl/pkg/p256k1"
)

func TestCreateUnsigned(t *testing.T) {
	var err error
	var signer pkg.Signer
	if signer, err = p256k1.NewSigner(&p256k1.Signer{}); chk.E(err) {
		t.Fatal(err)
	}
	var ok bool
	const relayURL = "wss://example.com"
	for _ = range 100 {
		challenge := GenerateChallenge()
		ev := CreateUnsigned(signer.Pub(), challenge, relayURL)
		if err = ev.Sign(signer); chk.E(err) {
			t.Fatal(err)
		}
		if ok, err = Validate(ev, challenge, relayURL); chk.E(err) {
			t.Fatal(err)
		}
		if !ok {
			bb, _ := ev.MarshalJSON(nil)
			t.Fatalf("failed to validate auth event\n%s", bb)
		}
	}
}
