//go:build !btcec

package p256k_test

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	btcec "ec.mleku.dev/v2"
	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/event/examples"
	"git.replicatr.dev/pkg/crypto/p256k"
	"github.com/minio/sha256-simd"
)

func TestVerify(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	for scanner.Scan() {
		var valid bool
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		} else {
			if valid, err = ev.Verify(); chk.E(err) || !valid {
				t.Errorf("btcec: invalid signature\n%s", b)
				continue
			}
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		if err = p256k.VerifyFromBytes(id, ev.Sig, ev.PubKey); chk.E(err) {
			t.Error(err)
			continue
		}
		evs = append(evs, ev)
	}
}

func TestSign(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	var sec1 *p256k.Sec
	var pub1 *p256k.XPublicKey
	var pb B
	if _, pb, sec1, pub1, _, err = p256k.Generate(); chk.E(err) {
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
	sig := make(B, schnorr.SignatureSize)
	for _, ev := range evs {
		ev.PubKey = pb
		var uid *p256k.Uchar
		if uid, err = p256k.Msg(ev.GetIDBytes()); chk.E(err) {
			t.Fatal(err)
		}
		if sig, err = p256k.Sign(uid, sec1.Sec()); chk.E(err) {
			t.Fatal(err)
		}
		ev.Sig = sig
		var usig *p256k.Uchar
		if usig, err = p256k.Sig(sig); chk.E(err) {
			t.Fatal(err)
		}
		if !p256k.Verify(uid, usig, pub1.Key) {
			t.Errorf("invalid signature")
		}
	}
	p256k.Zero(&sec1.Key)
}

func TestOnlyECDH(t *testing.T) {
	var err error
	var counter int
	const total = 10000
	var skb1, pkb1, skb2, pkb2 B
	if skb1, pkb1, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	if skb2, pkb2, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	pk1, pk2 := pkb1[1:], pkb2[1:]
	n := time.Now()
	for _ = range total {
		var secret1, secret2 B
		if secret1, err = p256k.ECDH(skb1, pk2); chk.E(err) {
			t.Fatal(err)
		}
		if secret2, err = p256k.ECDH(skb2, pk1); chk.E(err) {
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

func TestECDHAgainstBTCECDH(t *testing.T) {
	var err error
	var skb1, pkb1, skb2, pkb2 B
	if skb1, pkb1, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	if skb2, pkb2, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	pk1, pk2 := pkb1[1:], pkb2[1:]
	var secret1, secret2 B
	if secret1, err = p256k.ECDH(skb1, pk2); chk.E(err) {
		t.Fatal(err)
	}
	if secret2, err = p256k.ECDH(skb2, pk1); chk.E(err) {
		t.Fatal(err)
	}
	if !equals(secret1, secret2) {
		t.Errorf("ECDH generation failed to work in both directions, %x %x", secret1, secret2)
	}
	bs1, bp1 := btcec.PrivKeyFromBytes(skb1)
	bs2, bp2 := btcec.PrivKeyFromBytes(skb2)
	secret1b := btcec.GenerateSharedSecret(bs1, bp2)
	secret2b := btcec.GenerateSharedSecret(bs2, bp1)
	log.I.S(secret1, secret2, secret1b, secret2b)
}
