package subscriptionid

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	for _ = range 100 {
		b := make(B, frand.Intn(48)+1)
		bc := make(B, len(b))
		_, _ = frand.Read(b)
		copy(bc, b)
		si := T(b)
		m, err := si.MarshalJSON(nil)
		if chk.E(err) {
			t.Fatal(err)
		}
		var ui any
		var rem B
		ui, rem, err = T{}.UnmarshalJSON(m)
		if chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) > 0 {
			t.Errorf("len(rem): %d, '%s'", len(rem), rem)
		}
		uu := ui.(T)
		if !bytes.Equal(uu, bc) {
			t.Fatalf("bc: %0x, uu: %0x", bc, uu)
		}
	}
}
