package hex

import (
	"bytes"
	"encoding/hex"
	"testing"

	"lukechampine.com/frand"
)

func TestAppendHexToByteString(t *testing.T) {
	bin := make([]byte, 16)
	h := make([]byte, 0, 32)
	var err error
	for _ = range 5 {
		if _, err = frand.Read(bin); chk.E(err) {
			t.Fatal(err)
		}
		h = AppendHexToByteString(h, bin)
		if h, err = ByteStringToBytes(h); chk.E(err) {
			t.Fatal(err)
		}
		if bytes.Compare(bin, h) != 0 {
			t.Fatalf("mismatch %0x mangled to %0x", bin, h)
		}
		h = h[:0]
		h = hex.AppendEncode(h, bin)
		if _, err = hex.Decode(h, h); chk.E(err) {
			t.Fatal(err)
		}
		h = h[:0]
	}
}

func BenchmarkAppendHexToByteString(b *testing.B) {
	const size = 64
	bin := make([]byte, size)
	h := make([]byte, 0, size)
	var err error
	b.Run("AppendHexToByteString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(bin); chk.E(err) {
				b.Fatal(err)
			}
			h = AppendHexToByteString(h, bin)
			// if h, err = ByteStringToBytes(h); chk.E(err) {
			// 	b.Fatal(err)
			// }
			h = h[:0]
		}
	})
	b.Run("AppendHexToByteStringToHex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(bin); chk.E(err) {
				b.Fatal(err)
			}
			h = AppendHexToByteString(h, bin)
			if h, err = ByteStringToBytes(h); chk.E(err) {
				b.Fatal(err)
			}
			h = h[:0]
		}
	})
	b.Run("hex.AppendEncode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(bin); chk.E(err) {
				b.Fatal(err)
			}
			h = hex.AppendEncode(h, bin)
			// if _, err = hex.Decode(h, h); chk.E(err) {
			// 	b.Fatal(err)
			// }
			h = h[:0]
		}
	})
	b.Run("hex.AppendEncodeDecode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(bin); chk.E(err) {
				b.Fatal(err)
			}
			h = hex.AppendEncode(h, bin)
			if _, err = hex.Decode(h, h); chk.E(err) {
				b.Fatal(err)
			}
			h = h[:0]
		}
	})

}
