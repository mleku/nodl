package pubkey

import (
	"bytes"
	. "nostr.mleku.dev"
	"testing"

	"ec.mleku.dev/v2/schnorr"
	"lukechampine.com/frand"
)

func TestT(t *testing.T) {
	for _ = range 10000000 {
		fakePubkeyBytes := frand.Bytes(schnorr.PubKeyBytesLen)
		v, err := New(fakePubkeyBytes)
		if Chk.E(err) {
			t.FailNow()
		}
		buf := new(bytes.Buffer)
		v.Write(buf)
		buf2 := bytes.NewBuffer(buf.Bytes())
		v2, _ := New()
		el := v2.Read(buf2).(*T)
		if bytes.Compare(el.Val, v.Val) != 0 {
			t.Fatalf("expected %x got %x", v.Val, el.Val)
		}
	}
}
