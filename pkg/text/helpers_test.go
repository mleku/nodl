package text

import (
	"bytes"
	"testing"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/kinds"
	"lukechampine.com/frand"
)

func TestUnmarshalKindsArray(t *testing.T) {
	k := make(kinds.T, 100)
	for i := range k {
		k[i] = kind.T(frand.Intn(65535))
	}
	var dst B
	dst = MarshalKindsArray(dst, k)
	var k2 kinds.T
	var rem B
	var err error
	if k2, rem, err = UnmarshalKindsArray(dst); chk.E(err) {
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
		dst = AppendQuote(dst, ha[i], hex.EncAppend)
		if i != len(ha)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	var ha2 []B
	var rem B
	var err error
	if ha2, rem, err = UnmarshalHexArray(dst, 32); chk.E(err) {
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
}
