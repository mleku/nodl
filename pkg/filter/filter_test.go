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
	"github.com/mleku/nodl/pkg/tag"
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
		dst, _ = f.MarshalJSON(dst)
		// now unmarshal
		var f2 *T
		var rem B
		var fa any
		if fa, rem, err = New().UnmarshalJSON(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		f2 = fa.(*T)
		dst2, _ := f2.MarshalJSON(nil)
		if bytes.Equal(dst, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst, dst2)
		}
	}
}
