package authenvelope

import (
	"bytes"
	"testing"

	"github.com/mleku/nodl/pkg/auth"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	k1 "github.com/mleku/nodl/pkg/ec/secp256k1"
	"github.com/mleku/nodl/pkg/envelopes/sentinel"
)

const relayURL = "wss://example.com"

func TestChallenge(t *testing.T) {
	var err error
	var sec *k1.SecretKey
	if sec, err = k1.GenerateSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	pk := schnorr.SerializePubKey(sec.PubKey())
	var b B
	for _ = range 10000 {
		ch := auth.GenerateChallenge()
		chal := Challenge{Challenge: ch}
		if b, err = chal.Marshal(b); chk.E(err) {
			t.Fatal(err)
		}
		var chal2 *Challenge
		var rem B
		var l string
		if l, b, err = sentinel.Identify(b); chk.E(err) {
			t.Fatal(err)
		}
		orig := make(B, len(b))
		copy(orig, b)
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if chal2, rem, err = UnmarshalChallenge(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatal("remainder should be empty")
		}
		if !bytes.Equal(chal.Challenge, chal2.Challenge) {
			t.Fatalf("challenge mismatch\n%s\n%s",
				chal.Challenge, chal2.Challenge)
		}
		resp := Response{Event: auth.CreateUnsigned(pk, ch, relayURL)}
		if err = resp.Event.SignWithSecKey(sec); chk.E(err) {
			t.Fatal(err)
		}
		var b2 B
		if b2, err = resp.Marshal(b2); chk.E(err) {
			t.Fatal(err)
		}
		orig2 := make(B, len(b2))
		copy(orig2, b2)
		var resp2 *Response
		if resp2, _, err = UnmarshalResponse(b2); chk.E(err) {
			t.Fatal(err)
		}
		if b2, err = resp2.Marshal(b2); chk.E(err) {
			t.Fatal(err)
		}
	}
}
