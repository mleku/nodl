package filter

import (
	"testing"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err error
	dst := make([]byte, 0, 4000000)
	dst1 := make(B, 0, len(dst))
	dst2 := make(B, 0, len(dst))
	for _ = range 5 {
		var f *T
		if f, err = GenFilter(); chk.E(err) {
			t.Fatal(err)
		}
		dst, _ = f.MarshalJSON(dst)
		dst1 = append(dst1, dst...)
		// now unmarshal
		var rem B
		fa := New()
		if rem, err = fa.UnmarshalJSON(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2, _ = fa.MarshalJSON(dst2)
		if !equals(dst1, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst1, dst2)
		}
		dst, dst1, dst2 = dst[:0], dst1[:0], dst2[:0]
	}
}
