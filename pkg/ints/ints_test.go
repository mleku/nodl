package ints

import (
	"math"
	"strconv"
	"testing"

	"lukechampine.com/frand"
)

func TestByteStringToInt64(t *testing.T) {
	b := make([]byte, 0, 8)
	var m, n int64
	var err error
	_, _ = m, err
	for _ = range 10000000 {
		n = int64(frand.Intn(math.MaxInt64))
		b = Int64AppendToByteString(b, n)
		if m, _, err = ExtractInt64FromByteString(b); chk.E(err) {
			t.Fatal(err)
		}
		if n != m {
			t.Fatalf("failed to convert to int64 at %d %s %d", n, b, m)
		}
		b = b[:0]
	}
}

func BenchmarkByteStringToInt64(bb *testing.B) {
	b := make([]byte, 0, 19)
	var i int
	var err error
	testInts := make([]int64, 10000)
	for i = range 10000 {
		testInts[i] = int64(frand.Intn(math.MaxInt64))
	}

	bb.Run("Int64AppendToByteString", func(bb *testing.B) {
		bb.ReportAllocs()
		for i = 0; i < bb.N; i++ {
			n := testInts[i%10000]
			b = Int64AppendToByteString(b, n)
			b = b[:0]
		}
	})
	bb.Run("ByteStringToInt64ToByteString", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := testInts[i%10000]
			b = Int64AppendToByteString(b, int64(n))
			if _, _, err = ExtractInt64FromByteString(b); chk.E(err) {
				bb.Fatal(err)
			}
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
			_, err = strconv.Atoi(s)
		}
	})

}
