//go:build libsecp256k1

package libsecp256k1

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	k1 "github.com/mleku/btcec/secp256k1"
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
			if valid, err = ev.CheckSignature(); chk.E(err) || !valid {
				t.Errorf("btcec: invalid signature\n%s", b)
				continue
			}
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		if err = Verify(id, ev.Sig, ev.PubKey); chk.E(err) {
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
	var sec *k1.SecretKey
	if sec, err = k1.GenerateSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	sk := sec.Serialize()
	pk := sec.PubKey()
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
		ev.PubKey = schnorr.SerializePubKey(pk)
		id := ev.GetIDBytes()
		if sig, err = Sign(id, sk); chk.E(err) {
			t.Error(err)
		}
		ev.Sig = sig
		if err = Verify(id, ev.Sig, ev.PubKey); chk.E(err) {
			t.Error(err)
		}
	}
}
