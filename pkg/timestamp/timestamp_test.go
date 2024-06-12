package timestamp

import (
	"testing"
	"time"
)

func TestT(t *testing.T) {
	for i := 0; i < 1000; i++ {
		ts := Now()
		ti64 := int64(*ts)
		timey := time.Unix(ti64, 0)
		t1 := FromUnix(ti64)
		if t1.I64() != ti64 {
			t.Fatal("failed to convert timestamp to int64")
		}
		if t1.U64() != uint64(ti64) {
			t.Fatal("failed to convert timestamp to uint64")
		}
		if timey.Unix() != ti64 {
			t.Fatal("failed to convert timestamp to time.Time")
		}
		var err error
		var b []byte
		if b, err = t1.MarshalJSON(); chk.E(err) {
			t.Fatal(err)
		}
		var t2 T
		if err = t2.UnmarshalJSON(b); err != nil {
			t.Fatal(err)
		}
		if b, err = t1.MarshalBinary(); chk.E(err) {
			t.Fatal(err)
		}
		var t3 T
		if err = t3.UnmarshalBinary(b); err != nil {
			t.Fatal(err)
		}
		tan := time.Now()
		tn := FromTime(tan)
		tb := tn.Bytes()
		tfb := FromBytes(tb)
		ti := tfb.Int()
		tfi := T(ti)
		if tn != T(tfi) {
			t.Fatal("failed to convert timestamp to int64")
		}
		_ = tn.String()
		err = tn.UnmarshalBinary([]byte{255, 255, 255, 255, 255, 255, 255, 255})
		if err == nil {
			t.Fatal("should fail")
		}
	}
}
func BenchmarkT(bb *testing.B) {
	bb.Run("MarshalJSON", func(bb *testing.B) {
		ts := Now()
		ti64 := int64(*ts)
		t1 := FromUnix(ti64)
		var err error
		var b []byte
		if b, err = t1.MarshalJSON(); chk.E(err) {
			bb.Fatal(err)
		}
		_ = b
	})
	bb.Run("MarshalJSONUnmarshalJSON", func(bb *testing.B) {
		ts := Now()
		ti64 := int64(*ts)
		t1 := FromUnix(ti64)
		var err error
		var b []byte
		if b, err = t1.MarshalJSON(); chk.E(err) {
			bb.Fatal(err)
		}
		var t3 T
		if err = t3.UnmarshalBinary(b); err != nil {
			bb.Fatal(err)
		}
	})
}
