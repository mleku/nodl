//go:build libsecp256k1

package libsecp256k1

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
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("btcec: failed to marshal\n%s", b)
			continue
		}
		var valid bool
		if valid, err = ev.CheckSignature(); chk.E(err) || !valid {
			t.Errorf("btcec: invalid signature\n%s", b)
			continue
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		var ida [32]byte
		copy(ida[:], id)
		if len(ev.Sig) != 64 {
			t.Errorf("sig should be 64 bytes, got %d", len(ev.Sig))
			continue
		}
		var siga [64]byte
		copy(siga[:], ev.Sig)
		if len(ev.PubKey) != schnorr.PubKeyBytesLen {
			t.Errorf("pubkey should be 32 bytes got %d", len(ev.PubKey))
			continue
		}
		var puba [32]byte
		copy(puba[:], ev.PubKey)
		if !Verify(ida, siga, puba) {
			t.Errorf("libsecp256k1: invalid signature\n%s", b)
			continue
		}
		evs = append(evs, ev)
	}
}
