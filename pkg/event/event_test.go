package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"

	k1 "github.com/mleku/btcec/secp256k1"
	"github.com/mleku/nodl/pkg/event/examples"
)

func TestTMarshal_Unmarshal(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(B, 0, len(b))
		c = append(c, b...)
		ea := New()
		if _, err = ea.UnmarshalJSON(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		if out, err = ea.MarshalJSON(out); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		out = out[:0]
	}
}

func TestT_CheckSignature(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(B, 0, len(b))
		c = append(c, b...)
		ea := New()
		if _, err = ea.UnmarshalJSON(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var valid bool
		if valid, err = ea.CheckSignature(); chk.E(err) {
			t.Fatal(err)
		}
		if !valid {
			t.Fatalf("invalid signature\n%s", b)
		}
		out = out[:0]
	}
}

func TestT_SignWithSecKey(t *testing.T) {
	var err error
	var sec *k1.SecretKey
	if sec, err = k1.GenerateSecretKey(); chk.E(err) {
		t.Fatal(err)
	}
	var ev *T
	for _ = range 1000 {
		if ev, err = GenerateRandomTextNoteEvent(sec, 1000); chk.E(err) {
			t.Fatal(err)
		}
		var valid bool
		if valid, err = ev.CheckSignature(); chk.E(err) {
			t.Fatal(err)
		}
		if !valid {
			b, _ := ev.MarshalJSON(nil)
			t.Fatalf("invalid signature\n%s", b)
		}
	}
}

func BenchmarkUnmarshalMarshal(bb *testing.B) {
	var i int
	var out B
	var err error
	evts := make([]*T, 0, 10000)
	bb.Run("UnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
		buf := make(B, 1_000_000)
		scanner.Buffer(buf, len(buf))
		for i = 0; i < bb.N; i++ {
			if !scanner.Scan() {
				scanner = bufio.NewScanner(bytes.NewBuffer(examples.Cache))
				scanner.Scan()
			}
			b := scanner.Bytes()
			ea := New()
			if _, err = ea.UnmarshalJSON(b); chk.E(err) {
				bb.Fatal(err)
			}
			evts = append(evts, ea)
		}
	})
	bb.Run("MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		var counter int
		for i = 0; i < bb.N; i++ {
			out, _ = evts[counter].MarshalJSON(out)
			out = out[:0]
			counter++
			if counter != len(evts) {
				counter = 0
			}
		}
	})
}
