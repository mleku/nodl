package tags

import (
	"testing"

	"github.com/mleku/nodl/pkg/tag"
	"lukechampine.com/frand"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b B
	for _ = range 10 {
		n := frand.Intn(40) + 2
		tgs := make(T, 0, n)
		for _ = range n {
			n1 := frand.Intn(40) + 2
			tg := make(tag.T, 0, n)
			for _ = range n1 {
				b1 := make(B, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg = append(tg, b1)
			}
			tgs = append(tgs, tg)
		}
		b = tgs.Marshal(b)
		tgs2, rem, err := Unmarshal(b)
		if chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("len(rem)!=0:\n%s", rem)
		}
		if !tgs.Equal(tgs2) {
			t.Fatalf("got\n%s\nwant\n%s", tgs2, tgs)
		}
		b = b[:0]
	}
}

func BenchmarkMarshalUnmarshal(bb *testing.B) {
	var b B
	bb.Run("tag.Marshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := make(T, 0, n)
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := make(tag.T, 0, n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg = append(tg, b1)
				}
				tgs = append(tgs, tg)
			}
			b = tgs.Marshal(b)
			b = b[:0]
		}
	})
	bb.Run("tag.MarshalUnmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := make(T, 0, n)
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := make(tag.T, 0, n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg = append(tg, b1)
				}
				tgs = append(tgs, tg)
			}
			b = tgs.Marshal(b)
			tgs2, rem, err := Unmarshal(b)
			if chk.E(err) {
				bb.Fatal(err)
			}
			if len(rem) != 0 {
				bb.Fatalf("len(rem)!=0:\n%s", rem)
			}
			if !tgs.Equal(tgs2) {
				bb.Fatalf("got\n%s\nwant\n%s", tgs2, tgs)
			}
			b = b[:0]
		}
	})
}
