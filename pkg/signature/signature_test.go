package signature

import (
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/utils/bytestring"
	"github.com/mleku/nodl/pkg/utils/ec"
	"github.com/mleku/nodl/pkg/utils/ec/schnorr"
	"lukechampine.com/frand"
)

func TestAppendFromBinaryAppendFromHex(t *testing.T) {
	in := make([]byte, schnorr.SignatureSize)
	out := make([]byte, 0, schnorr.SignatureSize)
	hx := make([]byte, 0, schnorr.SignatureSize*2)
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

func TestAppendFromBinaryAppendFromHexQuote(t *testing.T) {
	in := make([]byte, schnorr.SignatureSize)
	out := make([]byte, 0, schnorr.SignatureSize)
	hx := make([]byte, 0, schnorr.SignatureSize+2+2)
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
	var err error
	var sk *ec.SecretKey
	var sig *T
	var pk *ec.PublicKey
	if sk, err = ec.NewSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	pk = sk.PubKey()
	in := make([]byte, sha256.Size)
	var j []byte
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		if sig, err = Sign(sk, in); chk.E(err) {
			t.Fatal(err)
		}
		if j, err = sig.MarshalJSON(); chk.E(err) {
			t.Fatal(err)
		}
		if err = sig.UnmarshalJSON(j); chk.E(err) {
			t.Fatal(err)
		}
		if err = sig.Verify(in, pk); chk.E(err) {
			t.Fatal(err)
		}
		j = j[:0]
	}
}

func BenchmarkT(b *testing.B) {
	b.Run("AppendHexFromBinary", func(b *testing.B) {
		in := make([]byte, schnorr.SignatureSize)
		hx := make([]byte, 0, schnorr.SignatureSize*2)
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
		in := make([]byte, schnorr.SignatureSize)
		out := make([]byte, 0, schnorr.SignatureSize)
		hx := make([]byte, 0, schnorr.SignatureSize*2)
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
		in := make([]byte, schnorr.SignatureSize)
		hx := make([]byte, 0, schnorr.SignatureSize*2)
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
		in := make([]byte, schnorr.SignatureSize)
		out := make([]byte, 0, schnorr.SignatureSize)
		hx := make([]byte, 0, schnorr.SignatureSize*2+2)
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
	b.Run("SignMarshalJSON", func(b *testing.B) {
		var err error
		var sk *ec.SecretKey
		var sig *T
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			b.Fatal(err)
		}
		in := make([]byte, sha256.Size)
		var j []byte
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			if sig, err = Sign(sk, in); chk.E(err) {
				b.Fatal(err)
			}
			if j, err = sig.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			j = j[:0]
		}
	})
	b.Run("SignMarshalJSONUnmarshalJSON", func(b *testing.B) {
		var err error
		var sk *ec.SecretKey
		var sig *T
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			b.Fatal(err)
		}
		in := make([]byte, sha256.Size)
		var j []byte
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			if sig, err = Sign(sk, in); chk.E(err) {
				b.Fatal(err)
			}
			if j, err = sig.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			if err = sig.UnmarshalJSON(j); chk.E(err) {
				b.Fatal(err)
			}
			j = j[:0]
		}
	})
	b.Run("SignMarshalJSONUnmarshalJSONVerify", func(b *testing.B) {
		var err error
		var sk *ec.SecretKey
		var sig *T
		var pk *ec.PublicKey
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			b.Fatal(err)
		}
		pk = sk.PubKey()
		in := make([]byte, sha256.Size)
		var j []byte
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			if sig, err = Sign(sk, in); chk.E(err) {
				b.Fatal(err)
			}
			if j, err = sig.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			if err = sig.UnmarshalJSON(j); chk.E(err) {
				b.Fatal(err)
			}
			if err = sig.Verify(in, pk); chk.E(err) {
				b.Fatal(err)
			}
			j = j[:0]
		}
	})
}
