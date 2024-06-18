package tag

import (
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b B
	for _ = range 1000 {
		n := frand.Intn(64) + 2
		tg := make(T, 0, n)
		for _ = range n {
			b1 := make(B, frand.Intn(128)+2)
			_, _ = frand.Read(b1)
			tg = append(tg, b1)
		}
		b = tg.Marshal(b)
		tg2, rem, err := Unmarshal(b)
		if chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("len(rem)!=0:\n%s", rem)
		}
		if !tg.Equal(tg2) {
			t.Fatalf("got\n%s\nwant\n%s", tg2, tg)
		}
		b = b[:0]
	}
}

func BenchmarkMarshalUnmarshal(bb *testing.B) {
	b := make(B, 0, 40000000)
	tg := make(T, 0, 2048)
	n := 4096
	for _ = range n {
		b1 := make(B, 128)
		_, _ = frand.Read(b1)
		tg = append(tg, b1)
	}
	bb.Run("tag.Marshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			b = tg.Marshal(b)
			b = b[:0]
			tg = tg[:0]
		}
	})
	bb.Run("tag.MarshalUnmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			b = tg.Marshal(b)
			_, _, _ = Unmarshal(b)
			b = b[:0]
			tg = tg[:0]
		}
	})
}
