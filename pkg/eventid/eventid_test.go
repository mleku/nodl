package eventid

import (
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/utils/bytestring"
	"lukechampine.com/frand"
)

func TestAppendHexFromBinaryAppendBinaryFromHex(t *testing.T) {
	in := make([]byte, sha256.Size)
	out := make([]byte, 0, sha256.Size)
	hx := make([]byte, 0, sha256.Size*2)
	var err error
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		hx = bytestring.AppendHexFromBinary(hx, in, false)
		if out, err = bytestring.AppendBinaryFromHex(out, hx,
			false); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(in, out) {
			log.I.S(in, out)
			t.Fatalf("AppendHexFromBinary returned wrong bytes:\n%0x\n%0x", in,
				out)
		}
		hx, out = hx[:0], out[:0]
	}
}

func TestAppendHexFromBinaryAppendBinaryFromHexQuote(t *testing.T) {
	in := make([]byte, sha256.Size)
	out := make([]byte, 0, sha256.Size)
	hx := make([]byte, 0, sha256.Size+2+2)
	var err error
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		hx = bytestring.AppendHexFromBinary(hx, in, true)
		if out, err = bytestring.AppendBinaryFromHex(out, hx,
			true); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(in, out) {
			t.Fatalf("AppendHexFromBinary returned wrong bytes:\n%0x\n%0x", in,
				out)
		}
		hx, out = hx[:0], out[:0]
	}
}

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	in := make([]byte, sha256.Size)
	var err error
	if _, err = frand.Read(in); chk.E(err) {
		t.Fatal(err)
	}
	var eid *T
	if eid, err = NewFromBytes(in); chk.E(err) {
		t.Fatal(err)
	}
	var j []byte
	if j, err = eid.MarshalJSON(); chk.E(err) {
		t.Fatal(err)
	}
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		if err = eid.Set(in); chk.E(err) {
			t.Fatal(err)
		}
		if j, err = eid.MarshalJSON(); chk.E(err) {
			t.Fatal(err)
		}
		if err = eid.UnmarshalJSON(j); chk.E(err) {
			t.Fatal(err)
		}
	}
}

func BenchmarkT(b *testing.B) {
	b.Run("AppendHexFromBinary", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		hx := make([]byte, 0, sha256.Size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = bytestring.AppendHexFromBinary(hx, in, false)
			hx = hx[:0]
		}
	})
	b.Run("AppendHexFromBinaryAppendBinaryFromHex", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		out := make([]byte, 0, sha256.Size)
		hx := make([]byte, 0, sha256.Size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = bytestring.AppendHexFromBinary(hx, in, false)
			if out, err = bytestring.AppendBinaryFromHex(out, hx,
				false); chk.E(err) {
				b.Fatal(err)
			}
			if !bytes.Equal(in, out) {
				b.Fatalf("AppendHexFromBinary returned wrong bytes:\n%0x\n%0x",
					in,
					out)
			}
			hx, out = hx[:0], out[:0]
		}
	})
	b.Run("AppendHexFromBinaryQuote", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		hx := make([]byte, 0, sha256.Size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = bytestring.AppendHexFromBinary(hx, in, true)
			hx = hx[:0]
		}
	})
	b.Run("AppendHexFromBinaryAppendBinaryFromHexQuote", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		out := make([]byte, 0, sha256.Size)
		hx := make([]byte, 0, sha256.Size*2+2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = bytestring.AppendHexFromBinary(hx, in, true)
			if out, err = bytestring.AppendBinaryFromHex(out, hx,
				true); chk.E(err) {
				b.Fatal(err)
			}
			if !bytes.Equal(in, out) {
				b.Fatalf("AppendHexFromBinary returned wrong bytes:\n%0x\n%0x",
					in,
					out)
			}
			hx, out = hx[:0], out[:0]
		}
	})
	b.Run("AppendMarshalJSON", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		var err error
		if _, err = frand.Read(in); chk.E(err) {
			b.Fatal(err)
		}
		var eid *T
		if eid, err = NewFromBytes(in); chk.E(err) {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			if err = eid.Set(in); chk.E(err) {
				b.Fatal(err)
			}
			if _, err = eid.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			eid.Reset()
		}
	})
	b.Run("AppendMarshalJSONUnmarshalJSON", func(b *testing.B) {
		in := make([]byte, sha256.Size)
		var err error
		if _, err = frand.Read(in); chk.E(err) {
			b.Fatal(err)
		}
		var eid *T
		if eid, err = NewFromBytes(in); chk.E(err) {
			b.Fatal(err)
		}
		var j []byte
		if j, err = eid.MarshalJSON(); chk.E(err) {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			if err = eid.Set(in); chk.E(err) {
				b.Fatal(err)
			}
			if j, err = eid.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			if err = eid.UnmarshalJSON(j); chk.E(err) {
				b.Fatal(err)
			}
			eid.Reset()
		}
	})
}
