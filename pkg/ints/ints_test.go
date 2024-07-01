package ints

import (
	"math"
	"strconv"
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	b := make(B, 0, 8)
	var rem B
	var n T
	var err error
	for _ = range 10000000 {
		n = T(frand.Intn(math.MaxInt64))
		b, err = n.MarshalJSON(b)
		var mi any
		if mi, rem, err = New().UnmarshalJSON(b); chk.E(err) {
			t.Fatal(err)
		}
		m := mi.(T)
		if n != m {
			t.Fatalf("failed to convert to int64 at %d %s %d", n, b, m)
		}
		if len(rem) > 0 {
			t.Fatalf("leftover bytes after converting back: '%s'", rem)
		}
		b = b[:0]
	}
}

func BenchmarkByteStringToInt64(bb *testing.B) {
	b := make([]byte, 0, 19)
	var i int
	testInts := make([]T, 10000)
	for i = range 10000 {
		testInts[i] = T(frand.Intn(math.MaxInt64))
	}
	bb.Run("MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i = 0; i < bb.N; i++ {
			n := testInts[i%10000]
			b, _ = n.MarshalJSON(b)
			b = b[:0]
		}
	})
	bb.Run("MarshalJSONUnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := testInts[i%10000]
			b, _ = n.MarshalJSON(b)
			_, _, _ = New().UnmarshalJSON(b)
			b = b[:0]
		}
	})
	bb.Run("Itoa", func(bb *testing.B) {
		bb.ReportAllocs()
		var s string
		for i = 0; i < bb.N; i++ {
			n := testInts[i%10000]
			s = strconv.Itoa(int(n))
			_ = s
		}
	})
	bb.Run("ItoaAtoi", func(bb *testing.B) {
		bb.ReportAllocs()
		var s string
		for i = 0; i < bb.N; i++ {
			n := testInts[i%10000]
			s = strconv.Itoa(int(n))
			_, _ = strconv.Atoi(s)
		}
	})
}
