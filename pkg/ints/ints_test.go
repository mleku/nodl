package ints

import (
	"strconv"
	"testing"
)

func TestByteStringToInt64(t *testing.T) {
	b := make([]byte, 0, 8)
	var i, j int64
	var err error
	for i = range 100000000 {
		b = Int64AppendToByteString(b, i)
		if j, err = ByteStringToInt64(b); chk.E(err) {
			t.Fatal(err)
		}
		if i != j {
			t.Fatalf("failed to convert to int64 at %d", i)
		}
		b = b[:0]
	}
}
func BenchmarkByteStringToInt64(bb *testing.B) {
	b := make([]byte, 0, 8)
	var i int
	var err error
	bb.Run("Int64AppendToByteString", func(bb *testing.B) {
		for i = 0; i < bb.N; i++ {
			b = Int64AppendToByteString(b, int64(i))
			b = b[:0]
		}
	})
	bb.Run("ByteStringToInt64ToByteString", func(bb *testing.B) {
		for i := 0; i < bb.N; i++ {
			b = Int64AppendToByteString(b, int64(i))
			if _, err = ByteStringToInt64(b); chk.E(err) {
				bb.Fatal(err)
			}
			b = b[:0]
		}
	})
	bb.Run("Itoa", func(bb *testing.B) {
		var s string
		for i = 0; i < bb.N; i++ {
			s = strconv.Itoa(i)
			_ = s
		}
	})
	bb.Run("ItoaAtoi", func(bb *testing.B) {
		var s string
		var n int
		for i = 0; i < bb.N; i++ {
			s = strconv.Itoa(i)
			n, err = strconv.Atoi(s)
			_ = n
		}
	})

}
