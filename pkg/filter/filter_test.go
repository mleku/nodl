package filter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/ec/secp256k1"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/kinds"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/timestamp"
	"lukechampine.com/frand"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err error
	for _ = range 5 {
		f := &T{}
		for _ = range 1000 {
			id := make(B, sha256.Size)
			frand.Read(id)
			f.IDs = append(f.IDs, id)
		}
		for _ = range 10 {
			f.Kinds = append(f.Kinds, kind.T(frand.Intn(65535)))
		}
		for _ = range 10 {
			var sk *secp256k1.SecretKey
			if sk, err = secp256k1.GenerateSecretKey(); chk.E(err) {
				t.Fatal(err)
			}
			pk := sk.PubKey()
			f.Authors = append(f.Authors, schnorr.SerializePubKey(pk))
		}
		for i := range 10 {
			p := make(B, 0, schnorr.PubKeyBytesLen*2)
			p = hex.EncAppend(p, f.Authors[i])
			f.Tags = append(f.Tags, tag.T{B("p"), p})
			idb := make(B, sha256.Size)
			frand.Read(idb)
			id := make(B, 0, sha256.Size*2)
			id = hex.EncAppend(id, idb)
			f.Tags = append(f.Tags, tag.T{B("e"), id})
			f.Tags = append(f.Tags,
				tag.T{B("a"), B(fmt.Sprintf("%d:%s:", frand.Intn(65535), id))})
		}
		f.Since = timestamp.Now() - 100
		f.Search = B("token search text")
		dst := make([]byte, 0, 4000000)
		dst = f.Marshal(dst)
		// now unmarshal
		var f2 *T
		var rem B
		if f2, rem, err = Unmarshal(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2 := f2.Marshal(nil)
		if bytes.Equal(dst, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst, dst2)
		}
	}
}

func TestUnmarshalHexArray(t *testing.T) {
	var ha []B
	h := make(B, sha256.Size)
	frand.Read(h)
	var dst B
	for _ = range 20 {
		hh := sha256.Sum256(h)
		h = hh[:]
		ha = append(ha, h)
	}
	dst = append(dst, '[')
	for i := range ha {
		dst = text.AppendQuote(dst, ha[i], hex.EncAppend)
		if i != len(ha)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	log.I.F("%s", dst)
	var ha2 []B
	var rem B
	var err error
	if ha2, rem, err = text.UnmarshalHexArray(dst, 32); chk.E(err) {
		t.Fatal(err)
	}
	if len(ha2) != len(ha) {
		t.Fatalf("failed to unmarshal, got %d fields, expected %d", len(ha2),
			len(ha))
	}
	if len(rem) > 0 {
		t.Fatalf("failed to unmarshal, remnant afterwards '%s'", rem)
	}
	for i := range ha2 {
		if !bytes.Equal(ha[i], ha2[i]) {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, ha[i], ha2[i])
		}
	}
	log.I.F("%s", text.MarshalHexArray(nil, ha2))
}

func TestUnmarshalKindsArray(t *testing.T) {
	k := make(kinds.T, 100)
	for i := range k {
		k[i] = kind.T(frand.Intn(65535))
	}
	var dst B
	dst = text.MarshalKindsArray(dst, k)
	log.I.F("%s", dst)
	var k2 kinds.T
	var rem B
	var err error
	if k2, rem, err = text.UnmarshalKindsArray(dst); chk.E(err) {
		t.Fatal(err)
	}
	if len(rem) > 0 {
		t.Fatalf("failed to unmarshal, remnant afterwards '%s'", rem)
	}
	for i := range k {
		if k[i] != k2[i] {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, k[i], k2[i])
		}
	}
	log.I.F("%s", text.MarshalKindsArray(nil, k2))
}
