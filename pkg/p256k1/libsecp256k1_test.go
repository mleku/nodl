//go:build !btcec

package p256k1_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/event/examples"
	"github.com/mleku/nodl/pkg/p256k1"
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
		if err = p256k1.VerifyFromBytes(id, ev.Sig, ev.PubKey); chk.E(err) {
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
	var sec1 *p256k1.Sec
	if sec1, err = p256k1.GenSec(); chk.E(err) {
		t.Fatal(err)
	}
	var pub1 *p256k1.Pub
	if pub1, err = sec1.Pub(); chk.E(err) {
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
	var pb B
	if pb, err = pub1.ToBytes(); chk.E(err) {
		t.Fatal(err)
	}
	for _, ev := range evs {
		ev.PubKey = pb
		var uid *p256k1.Uchar
		if uid, err = p256k1.Msg(ev.GetIDBytes()); chk.E(err) {
			t.Fatal(err)
		}
		if sig, err = p256k1.Sign(uid, sec1.Sec()); chk.E(err) {
			t.Fatal(err)
		}
		ev.Sig = sig
		var usig *p256k1.Uchar
		if usig, err = p256k1.Sig(sig); chk.E(err) {
			t.Fatal(err)
		}
		if !p256k1.Verify(uid, usig, pub1.Pub()) {
			t.Errorf("invalid signature")
		}
	}
	p256k1.Zero(&sec1.Key)
}
