package btcec_test

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
	"git.replicatr.dev/pkg/crypto/p256k/btcec"
	"github.com/minio/sha256-simd"
)

func TestBTCECSignerVerify(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	signer := &btcec.Signer{}
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

func TestBTCECSignerSign(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error

	signer := &btcec.Signer{}
	var skb B
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	skb = signer.Sec()
	if err = signer.InitSec(skb); chk.E(err) {
		t.Fatal(err)
	}
	verifier := &btcec.Signer{}
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

func TestBTCECECDH(t *testing.T) {
	n := time.Now()
	var err error
	var s1, s2 pkg.Signer
	var counter int
	const total = 10000
	if s1, err = p256k.NewSigner(&btcec.Signer{}); chk.E(err) {
		t.Fatal(err)
	}
	if s2, err = p256k.NewSigner(&btcec.Signer{}); chk.E(err) {
		t.Fatal(err)
	}
	for _ = range total {
		var secret1, secret2 B
		if secret1, err = s1.ECDH(s2.Pub()); chk.E(err) {
			t.Fatal(err)
		}
		if secret2, err = s2.ECDH(s1.Pub()); chk.E(err) {
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
