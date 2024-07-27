package p256k_test

import (
	"bufio"
	"bytes"
	"testing"

	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/event/examples"
	"git.replicatr.dev/pkg/crypto/p256k"
	"github.com/minio/sha256-simd"
)

func TestSignerVerify(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	signer := &p256k.Signer{}
	for scanner.Scan() {
		var valid bool
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		} else {
			if valid, err = ev.Verify(); chk.E(err) || !valid {
				t.Errorf("invalid signature\n%s", b)
				continue
			}
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		if err = signer.InitPub(ev.PubKey); chk.E(err) {
			t.Errorf("failed to init pub key: %s\n%0x", err, b)
		}
		if valid, err = signer.Verify(id, ev.Sig); chk.E(err) {
			t.Errorf("failed to verify: %s\n%0x", err, b)
		}
		if !valid {
			t.Errorf("invalid signature for pub %0x %0x %0x", ev.PubKey, id,
				ev.Sig)
		}
		evs = append(evs, ev)
	}
}

func TestSignerSign(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	signer := &p256k.Signer{}
	var skb B
	if skb, err = p256k.GenSecBytes(); chk.E(err) {
		t.Fatal(err)
	}
	if err = signer.InitSec(skb); chk.E(err) {
		t.Fatal(err)
	}
	verifier := &p256k.Signer{}
	pkb := signer.Pub()
	if err = verifier.InitPub(pkb); chk.E(err) {
		t.Fatal(err)
	}
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		}
		evs = append(evs, ev)
	}
	var valid bool
	sig := make(B, schnorr.SignatureSize)
	for _, ev := range evs {
		ev.PubKey = pkb
		id := ev.GetIDBytes()
		if sig, err = signer.Sign(id); chk.E(err) {
			t.Errorf("failed to sign: %s\n%0x", err, id)
		}
		if valid, err = verifier.Verify(id, sig); chk.E(err) {
			t.Errorf("failed to verify: %s\n%0x", err, id)
		}
		if !valid {
			t.Errorf("invalid signature")
		}
	}
	signer.Zero()
}
