package authenvelope

import (
	"bytes"
	"testing"

	"github.com/mleku/nodl/pkg/auth"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	k1 "github.com/mleku/nodl/pkg/ec/secp256k1"
	"github.com/mleku/nodl/pkg/envelopes"
)

const relayURL = "wss://example.com"

func TestAuth(t *testing.T) {
	var err error
	var sec *k1.SecretKey
	if sec, err = k1.GenerateSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	pk := schnorr.SerializePubKey(sec.PubKey())
	var b1, b2, b3, b4 B
	for _ = range 1000 {
		ch := auth.GenerateChallenge()
		chal := Challenge{Challenge: ch}
		if b1, err = chal.MarshalJSON(b1); chk.E(err) {
			t.Fatal(err)
		}
		oChal := make(B, len(b1))
		copy(oChal, b1)
		var rem B
		var l string
		if l, b1, err = envelopes.Identify(b1); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		var c2 any
		if c2, rem, err = NewChallenge().UnmarshalJSON(b1); chk.E(err) {
			t.Fatal(err)
		}
		chal2 := c2.(*Challenge)
		if len(rem) != 0 {
			t.Fatal("remainder should be empty")
		}
		if !bytes.Equal(chal.Challenge, chal2.Challenge) {
			t.Fatalf("challenge mismatch\n%s\n%s",
				chal.Challenge, chal2.Challenge)
		}
		if b2, err = c2.(*Challenge).MarshalJSON(b2); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(oChal, b2) {
			t.Fatalf("challenge mismatch\n%s\n%s", oChal, b2)
		}
		resp := Response{Event: auth.CreateUnsigned(pk, ch, relayURL)}
		if err = resp.Event.SignWithSecKey(sec); chk.E(err) {
			t.Fatal(err)
		}
		if b3, err = resp.MarshalJSON(b3); chk.E(err) {
			t.Fatal(err)
		}
		oResp := make(B, len(b3))
		copy(oResp, b3)
		if l, b3, err = envelopes.Identify(b3); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		var r2 any
		if r2, _, err = NewResponse().UnmarshalJSON(b3); chk.E(err) {
			t.Fatal(err)
		}
		resp2 := r2.(*Response)
		if b4, err = resp2.MarshalJSON(b4); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(oResp, b4) {
			t.Fatalf("challenge mismatch\n%s\n%s", oResp, b4)
		}
		b1, b2, b3, b4 = b1[:0], b2[:0], b3[:0], b4[:0]
		oChal, oResp = oChal[:0], oChal[:0]
	}
}
