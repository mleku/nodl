package pubkey

import (
	"bytes"
	"testing"

	"github.com/mleku/nodl/pkg/utils/ec"
	"github.com/mleku/nodl/pkg/utils/ec/schnorr"
	"lukechampine.com/frand"
)

func TestAppendFromBinaryAppendFromHex(t *testing.T) {
	in := make([]byte, schnorr.PubKeyBytesLen)
	out := make([]byte, 0, schnorr.PubKeyBytesLen)
	hx := make([]byte, 0, schnorr.PubKeyBytesLen*2)
	var err error
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		hx = AppendHexFromBinary(hx, in, false)
		if out, err = AppendBinaryFromHex(out, hx, false); chk.E(err) {
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
	in := make([]byte, schnorr.PubKeyBytesLen)
	out := make([]byte, 0, schnorr.PubKeyBytesLen)
	hx := make([]byte, 0, schnorr.PubKeyBytesLen+2+2)
	var err error
	for _ = range 100 {
		if _, err = frand.Read(in); chk.E(err) {
			t.Fatal(err)
		}
		hx = AppendHexFromBinary(hx, in, true)
		if out, err = AppendBinaryFromHex(out, hx, true); chk.E(err) {
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
	var pk *T
	var j []byte
	for _ = range 100 {
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			t.Fatal(err)
		}
		if pk, err = NewFromPubKey(sk.PubKey()); chk.E(err) {
			t.Fatal(err)
		}
		if j, err = pk.MarshalJSON(); chk.E(err) {
			t.Fatal(err)
		}
		if err = pk.UnmarshalJSON(j); chk.E(err) {
			t.Fatal(err)
		}
		j = j[:0]
	}
}

func BenchmarkT(b *testing.B) {
	b.Run("AppendHexFromBinary", func(b *testing.B) {
		in := make([]byte, schnorr.PubKeyBytesLen)
		hx := make([]byte, 0, schnorr.PubKeyBytesLen*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = AppendHexFromBinary(hx, in, false)
			hx = hx[:0]
		}
	})
	b.Run("AppendHexFromBinaryAppendBinaryFromHex", func(b *testing.B) {
		in := make([]byte, schnorr.PubKeyBytesLen)
		out := make([]byte, 0, schnorr.PubKeyBytesLen)
		hx := make([]byte, 0, schnorr.PubKeyBytesLen*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = AppendHexFromBinary(hx, in, false)
			if out, err = AppendBinaryFromHex(out, hx, false); chk.E(err) {
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
		in := make([]byte, schnorr.PubKeyBytesLen)
		hx := make([]byte, 0, schnorr.PubKeyBytesLen*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = AppendHexFromBinary(hx, in, true)
			hx = hx[:0]
		}
	})
	b.Run("AppendHexFromBinaryAppendBinaryFromHexQuote", func(b *testing.B) {
		in := make([]byte, schnorr.PubKeyBytesLen)
		out := make([]byte, 0, schnorr.PubKeyBytesLen)
		hx := make([]byte, 0, schnorr.PubKeyBytesLen*2+2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			hx = AppendHexFromBinary(hx, in, true)
			if out, err = AppendBinaryFromHex(out, hx, true); chk.E(err) {
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
		var err error
		var sk *ec.SecretKey
		var pk *T
		var j []byte
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			b.Fatal(err)
		}
		if pk, err = NewFromPubKey(sk.PubKey()); chk.E(err) {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			if j, err = pk.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			j = j[:0]
		}
	})
	b.Run("AppendMarshalJSONUnmarshalJSON", func(b *testing.B) {
		var err error
		var sk *ec.SecretKey
		var pk *T
		var j []byte
		if sk, err = ec.NewSecretKey(); chk.E(err) {
			b.Fatal(err)
		}
		if pk, err = NewFromPubKey(sk.PubKey()); chk.E(err) {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			if j, err = pk.MarshalJSON(); chk.E(err) {
				b.Fatal(err)
			}
			if err = pk.UnmarshalJSON(j); chk.E(err) {
				b.Fatal(err)
			}
			j = j[:0]
		}
	})
}
