package filters

import (
	"fmt"

	"github.com/minio/sha256-simd"
	"github.com/mleku/btcec/schnorr"
	"github.com/mleku/btcec/secp256k1"
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/filter"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/tag"
	"github.com/mleku/nodl/pkg/timestamp"
	"lukechampine.com/frand"
)

type T struct {
	F []*filter.T
}

func New() (f *T) { return &T{} }

func (f *T) Match(event *event.T) bool {
	for _, f := range f.F {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (f *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	end := len(f.F) - 1
	for i := range f.F {
		if b, err = f.F[i].MarshalJSON(b); chk.E(err) {
			return
		}
		if i < end {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (f *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b[:]
	for len(rem) > 0 {
		switch rem[0] {
		case '[':
			if len(rem) > 1 && rem[1] == ']' {
				rem = rem[1:]
				return
			}
			log.I.Ln("first filter")
			ffa := filter.New()
			if rem, err = ffa.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
			log.I.F("%s", rem)
			// continue
		case ',':
			rem = rem[1:]
			if len(rem) > 1 && rem[1] == ']' {
				rem = rem[1:]
				return
			}
			log.I.Ln("nth filter")
			ffa := filter.New()
			if rem, err = ffa.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
		// next
		case ']':
			rem = rem[1:]
			// the end
			return
		}
	}
	return
}

func GenFilters(n int) (ff *T, err error) {
	ff = &T{}
	for _ = range n {
		f := filter.New()
		for _ = range 5 {
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
				return
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
		tn := timestamp.Now()
		f.Since = timestamp.FromUnix(int64(*tn - 100))
		f.Search = B("token search text")
		ff.F = append(ff.F, f)
	}
	return
}
