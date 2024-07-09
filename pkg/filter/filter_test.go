package filter

import (
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
		f := New()
		for _ = range 1000 {
			id := make(B, sha256.Size)
			frand.Read(id)
			f.IDs.T = append(f.IDs.T, id)
		}
		for _ = range 10 {
			f.Kinds.K = append(f.Kinds.K, kind.New(frand.Intn(65535)))
		}
		for _ = range 10 {
			var sk *secp256k1.SecretKey
			if sk, err = secp256k1.GenerateSecretKey(); chk.E(err) {
				t.Fatal(err)
			}
			pk := sk.PubKey()
			f.Authors.T = append(f.Authors.T, schnorr.SerializePubKey(pk))
		}
		for i := range 10 {
			p := make(B, 0, schnorr.PubKeyBytesLen*2)
			p = hex.EncAppend(p, f.Authors.T[i])
			f.Tags.T = append(f.Tags.T, tag.New(B("p"), p))
			idb := make(B, sha256.Size)
			frand.Read(idb)
			id := make(B, 0, sha256.Size*2)
			id = hex.EncAppend(id, idb)
			f.Tags.T = append(f.Tags.T, tag.New(B("e"), id))
			f.Tags.T = append(f.Tags.T,
				tag.New(B("a"),
					B(fmt.Sprintf("%d:%s:", frand.Intn(65535), id))))
		}
		tn := *timestamp.Now() - 100
		f.Since = &tn
		f.Search = B("token search text")
		dst := make([]byte, 0, 4000000)
		dst, _ = f.MarshalJSON(dst)
		// now unmarshal
		f2 := &T{}
		var rem B
		fa := New()
		if rem, err = fa.UnmarshalJSON(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2, _ := f2.MarshalJSON(nil)
		if equals(dst, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst, dst2)
		}
	}
}
