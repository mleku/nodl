package filters

import (
	"testing"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err error
	dst := make([]byte, 0, 4000000)
	for _ = range 1000 {
		var ff *T
		ff, err = GenFilters(5)
		// now unmarshal
		if dst, err = ff.MarshalJSON(dst); chk.E(err) {
			t.Fatal(err)
		}
		fa := New()
		var rem B
		if rem, err = fa.UnmarshalJSON(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		f2 := New()
		dst2, _ := f2.MarshalJSON(nil)
		if equals(dst, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst, dst2)
		}
		dst = dst[:0]
	}
}
