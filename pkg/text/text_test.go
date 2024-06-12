package text

import (
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	const MaxBytes = 4000000 // 4Mb
	b := make([]byte, MaxBytes)
	var err error
	var j []byte
	for _ = range 100 {
		l := frand.Intn(MaxBytes)
		original := NewFromBytes(b[:l])
		// fill it with stuff
		frand.Read(original.Bytes())
		j, _ = original.MarshalJSON()
		copyOf := New()
		if err = copyOf.UnmarshalJSON(j); chk.E(err) {
			t.Fatal(err)
		}
		if !copyOf.Equal(original) {
			t.Fatalf("failed to unmarshal JSON")
		}
	}
}

func BenchmarkT(bb *testing.B) {
	const MaxBytes = 4000000 // 4Mb
	b := make([]byte, MaxBytes)
	var j []byte
	bb.Run("MarshalJson", func(bb *testing.B) {
		l := frand.Intn(MaxBytes)
		original := NewFromBytes(b[:l])
		// fill it with stuff
		frand.Read(original.Bytes())
		j, _ = original.MarshalJSON()
	})
	bb.Run("MarshalJsonUnmarshalJSON", func(bb *testing.B) {
		l := frand.Intn(MaxBytes)
		original := NewFromBytes(b[:l])
		// fill it with stuff
		frand.Read(original.Bytes())
		j, _ = original.MarshalJSON()
		_ = New().UnmarshalJSON(j)
	})
}
