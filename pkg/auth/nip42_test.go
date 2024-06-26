package auth

import (
	"testing"

	"github.com/mleku/nodl/pkg/ec/schnorr"
	k1 "github.com/mleku/nodl/pkg/ec/secp256k1"
	"github.com/mleku/nodl/pkg/hex"
	"lukechampine.com/frand"
)

func TestCreateUnsigned(t *testing.T) {
	var err error
	var sec *k1.SecretKey
	if sec, err = k1.GenerateSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	var ok bool
	pk := schnorr.SerializePubKey(sec.PubKey())
	const relayURL = "wss://example.com"
	for _ = range 100 {
		challenge := hex.Enc(frand.Bytes(16))
		ev := CreateUnsigned(pk, challenge, relayURL)
		if err = ev.SignWithSecKey(sec); chk.E(err) {
			t.Fatal(err)
		}
		if ok, err = Validate(ev, challenge, relayURL); chk.E(err) {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("failed to validate auth event\n%s", ev.Marshal(nil))
		}
	}
}
