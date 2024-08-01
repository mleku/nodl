//go:build !btcec

package p256k_test

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/event/examples"
	"git.replicatr.dev/pkg/crypto/p256k"
	"github.com/minio/sha256-simd"
)

func TestSigner_Generate(t *testing.T) {
	for _=range 100{
		var err error
		signer := &p256k.Signer{}
		var skb B
		if err = signer.Generate(); chk.E(err) {
			t.Fatal(err)
		}
		skb = signer.Sec()
		if err = signer.InitSec(skb); chk.E(err) {
			t.Fatal(err)
		}
	}
}

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
	var skb, pkb B
	if skb, pkb, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	if err = signer.InitSec(skb); chk.E(err) {
		t.Fatal(err)
	}
	verifier := &p256k.Signer{}
	if err = verifier.InitPub(pkb[1:]); chk.E(err) {
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

func TestECDH(t *testing.T) {
	var err error
	var s1, s2 pkg.Signer
	var counter int
	const total = 10000
	s1, s2 = &p256k.Signer{}, &p256k.Signer{}
	if err = s1.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	if err = s2.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	n := time.Now()
	for _ = range total {
		var secret1, secret2 B
		if secret1, err = s1.ECDH(s2.ECPub()); chk.E(err) {
			t.Fatal(err)
		}
		if secret2, err = s2.ECDH(s1.ECPub()); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(secret1, secret2) {
			counter++
			t.Errorf("ECDH generation failed to work in both directions, %x %x", secret1, secret2)
		}
	}
	a := time.Now()
	duration := a.Sub(n)
	log.I.Ln("errors", counter, "total", total, "time", duration, "time/op", int(duration/total),
		"ops/sec", int(time.Second)/int(duration/total))
}
