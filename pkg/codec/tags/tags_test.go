package tags

import (
	"testing"

	"github.com/mleku/nodl/pkg/codec/tag"
	"lukechampine.com/frand"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b, rem B
	var err error
	for _ = range 10 {
		n := frand.Intn(40) + 2
		tgs := New()
		for _ = range n {
			n1 := frand.Intn(40) + 2
			tg := tag.NewWithCap(n)
			for _ = range n1 {
				b1 := make(B, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg.Field = append(tg.Field, b1)
			}
			tgs.T = append(tgs.T, tg)
		}
		b, _ = tgs.MarshalJSON(b)
		ta := New()
		rem, err = ta.UnmarshalJSON(b)
		if chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("len(rem)!=0:\n%s", rem)
		}
		if !tgs.Equal(ta) {
			t.Fatalf("got\n%s\nwant\n%s", ta, tgs)
		}
		b = b[:0]
	}
}

func BenchmarkMarshalJSONUnmarshalJSON(bb *testing.B) {
	var b, rem B
	var err error
	bb.Run("tag.MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg.Field = append(tg.Field, b1)
				}
				tgs.T = append(tgs.T, tg)
			}
			b, _ = tgs.MarshalJSON(b)
			b = b[:0]
		}
	})
	bb.Run("tag.MarshalJSONUnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg.Field = append(tg.Field, b1)
				}
				tgs.T = append(tgs.T, tg)
			}
			b, _ = tgs.MarshalJSON(b)
			ta := New()
			rem, err = ta.UnmarshalJSON(b)
			if chk.E(err) {
				bb.Fatal(err)
			}
			if len(rem) != 0 {
				bb.Fatalf("len(rem)!=0:\n%s", rem)
			}
			if !tgs.Equal(ta) {
				bb.Fatalf("got\n%s\nwant\n%s", ta, tgs)
			}
			b = b[:0]
		}
	})
}
