package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"

	k1 "github.com/mleku/nodl/pkg/ec/secp256k1"
)

//go:embed tenthousand.jsonl
var eventCache []byte

func TestTMarshal_Unmarshal(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
	var ev *T
	var rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(B, 0, len(b))
		c = append(c, b...)
		if ev, rem, err = Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		if out, err = ev.Marshal(out); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		out = out[:0]
	}
}
func TestT_CheckSignature(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
	var ev *T
	var rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(B, 0, len(b))
		c = append(c, b...)
		if ev, rem, err = Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var valid bool
		if valid, err = ev.CheckSignature(); chk.E(err) {
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
			b, _ := ev.Marshal(nil)
			t.Fatalf("invalid signature\n%s", b)
		}
	}
}

func BenchmarkUnmarshalMarshal(bb *testing.B) {
	var i int
	var out B
	var ev *T
	var err error
	evts := make([]*T, 0, 10000)
	bb.Run("Unmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
		for i = 0; i < bb.N; i++ {
			if !scanner.Scan() {
				scanner = bufio.NewScanner(bytes.NewBuffer(eventCache))
				scanner.Scan()
			}
			b := scanner.Bytes()
			if ev, _, err = Unmarshal(b); chk.E(err) {
				bb.Fatal(err)
			}
			evts = append(evts, ev)
		}
	})
	bb.Run("Marshal", func(bb *testing.B) {
		bb.ReportAllocs()
		var counter int
		for i = 0; i < bb.N; i++ {
			out, _ = evts[counter].Marshal(out)
			out = out[:0]
			counter++
			if counter != len(evts) {
				counter = 0
			}
		}
	})
}
