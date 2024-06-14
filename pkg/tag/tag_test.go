package tag

import (
	"testing"

	"github.com/mleku/nodl/pkg/text"
	"lukechampine.com/frand"
)

func TestAppendBinaryExtractBinary(t *testing.T) {
	const MaxBytes = 4000 // 4Kb
	var err error
	for _ = range 100 {
		n := frand.Intn(100) + 4
		tg := &T{}
		for _ = range n {
			l := frand.Intn(MaxBytes)
			nb := make([]byte, l)
			for i := range nb {
				nb[i] = byte(i)
			}
			*tg = append(*tg, text.NewFromBytes(nb))
		}
		var b, rem []byte
		b = AppendBinary(b, tg)
		var tg2 *T
		if tg2, rem, err = ExtractBinary(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) > 0 {
			t.Fatalf("remaining bytes: %v", rem)
		}
		if !tg.Equal(tg2) {
			t.Fatalf("failed %0x", tg2)
		}
		*tg, *tg2 = (*tg)[:0], (*tg2)[:0]
	}
}

func BenchmarkT(t *testing.B) {
	const MaxBytes = 4000 // 4Kb
	var err error
	t.Run("AppendBinary", func(bb *testing.B) {
		for _ = range 100 {
			n := frand.Intn(100) + 4
			tg := &T{}
			for _ = range n {
				l := frand.Intn(MaxBytes)
				nb := make([]byte, l)
				for i := range nb {
					nb[i] = byte(i)
				}
				*tg = append(*tg, text.NewFromBytes(nb))
			}
			var b []byte
			b = AppendBinary(b, tg)
			*tg = (*tg)[:0]
		}
	})
	t.Run("AppendBinaryExtractBinary", func(bb *testing.B) {
		for _ = range 100 {
			n := frand.Intn(100) + 4
			tg := &T{}
			for _ = range n {
				l := frand.Intn(MaxBytes)
				nb := make([]byte, l)
				for i := range nb {
					nb[i] = byte(i)
				}
				*tg = append(*tg, text.NewFromBytes(nb))
			}
			var b []byte
			b = AppendBinary(b, tg)
			var tg2 *T
			if tg2, _, err = ExtractBinary(b); chk.E(err) {
				t.Fatal(err)
			}
			*tg, *tg2 = (*tg)[:0], (*tg2)[:0]
		}
	})
}
