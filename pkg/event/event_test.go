package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"
)

//go:embed tenthousand.jsonl
var eventCache []byte

func TestUnmarshal(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
	var ev T
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
		out = ev.Marshal(out)
		if !bytes.Equal(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		out = out[:0]
		_, _, _ = ev, rem, out
	}
}

func BenchmarkUnmarshalMarshal(bb *testing.B) {
	var i int
	var ev T
	var out B
	var err error
	bb.Run("Unmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
		for i = 0; i < bb.N; i++ {
			if !scanner.Scan() {
				scanner = bufio.NewScanner(bytes.NewBuffer(eventCache))
				scanner.Scan()
			}
			b := scanner.Bytes()
			if _, _, err = Unmarshal(b); chk.E(err) {
				bb.Fatal(err)
			}
		}
	})
	bb.Run("UnmarshalMarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		scanner := bufio.NewScanner(bytes.NewBuffer(eventCache))
		for i = 0; i < bb.N; i++ {
			if !scanner.Scan() {
				scanner = bufio.NewScanner(bytes.NewBuffer(eventCache))
				scanner.Scan()
			}
			b := scanner.Bytes()
			if _, _, err = Unmarshal(b); chk.E(err) {
				bb.Fatal(err)
			}
			out = ev.Marshal(out)
			out = out[:0]
		}
	})
}
