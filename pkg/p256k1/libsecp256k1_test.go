//go:build !btcec

package p256k1

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/event/examples"
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
		if err = VerifyFromBytes(id, ev.Sig, ev.PubKey); chk.E(err) {
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
	var sec1 *Sec
	if sec1, err = GenSec(); chk.E(err) {
		t.Fatal(err)
	}
	var pub1 *Pub
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
		var uid *Uchar
		if uid, err = Msg(ev.GetIDBytes()); chk.E(err) {
			t.Fatal(err)
		}
		if sig, err = Sign(uid, sec1.Sec()); chk.E(err) {
			t.Fatal(err)
		}
		ev.Sig = sig
		var usig *Uchar
		if usig, err = Sig(sig); chk.E(err) {
			t.Fatal(err)
		}
		if !Verify(uid, usig, pub1.Pub()) {
			t.Errorf("invalid signature")
		}
	}
	Zero(&sec1.Key)
}
