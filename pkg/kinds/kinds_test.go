package kinds

import (
	"testing"

	"github.com/mleku/nodl/pkg/kind"
	"lukechampine.com/frand"
)

func TestUnmarshalKindsArray(t *testing.T) {
	k := make(T, 100)
	for i := range k {
		k[i] = kind.T(frand.Intn(65535))
	}
	var dst B
	var err error
	if dst, err = k.MarshalJSON(dst); chk.E(err) {
		t.Fatal(err)
	}
	k2 := T{}
	var rem B
	var kk any
	if kk, rem, err = k2.UnmarshalJSON(dst); chk.E(err) {
		return
	}
	k2 = kk.(T)
	if len(rem) > 0 {
		t.Fatalf("failed to unmarshal, remnant afterwards '%s'", rem)
	}
	for i := range k {
		if k[i] != k2[i] {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, k[i], k2[i])
		}
	}
}
