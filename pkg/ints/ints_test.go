package ints

import (
	"math"
	"strconv"
	"testing"
)

func TestByteStringToInt64(t *testing.T) {
	b := make([]byte, 0, 8)
	var i, j int64
	var err error
	_, _ = j, err
	for i = range 10000 {
		b = Int64AppendToByteString(b, i)
		if j, _, err = ExtractInt64FromByteString(b); chk.E(err) {
			t.Fatal(err)
		}
		if i != j {
			t.Fatalf("failed to convert to int64 at %d", i)
		}
		b = b[:0]
	}
}
func BenchmarkByteStringToInt64(bb *testing.B) {
	b := make([]byte, 0, 19)
	var i int
	var err error
	bb.Run("Int64AppendToByteString", func(bb *testing.B) {
		bb.ReportAllocs()
		n := int64(math.MaxInt64)
		for i = 0; i < bb.N; i++ {
			b = Int64AppendToByteString(b, n)
			b = b[:0]
			n--
		}
	})
	bb.Run("ByteStringToInt64ToByteString", func(bb *testing.B) {
		bb.ReportAllocs()
		n := int64(math.MaxInt64)
		for i := 0; i < bb.N; i++ {
			b = Int64AppendToByteString(b, int64(n))
			if _, _, err = ExtractInt64FromByteString(b); chk.E(err) {
				bb.Fatal(err)
			}
			b = b[:0]
			n--
		}
	})
	bb.Run("Itoa", func(bb *testing.B) {
		bb.ReportAllocs()
		var s string
		n := int64(math.MaxInt64)
		for i = 0; i < bb.N; i++ {
			s = strconv.Itoa(int(n))
			_ = s
			n--
		}
	})
	bb.Run("ItoaAtoi", func(bb *testing.B) {
		bb.ReportAllocs()
		var s string
		n := int64(math.MaxInt64)
		for i = 0; i < bb.N; i++ {
			s = strconv.Itoa(int(n))
			_, err = strconv.Atoi(s)
			n--
		}
	})

}
