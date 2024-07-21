package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"

	"github.com/mleku/nodl/pkg"
	"github.com/mleku/nodl/pkg/codec/event/examples"
	"github.com/mleku/nodl/pkg/crypto/p256k"
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
		if valid, err = ea.Verify(); chk.E(err) {
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
	var signer pkg.Signer
	if signer, err = p256k.NewSigner(&p256k.Signer{}); chk.E(err) {
		t.Fatal(err)
	}
	var ev *T
	for _ = range 1000 {
		if ev, err = GenerateRandomTextNoteEvent(signer, 1000); chk.E(err) {
			t.Fatal(err)
		}
		var valid bool
		if valid, err = ev.Verify(); chk.E(err) {
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
	evts := make([]*T, 9999)
	bb.Run("UnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
		buf := make(B, 1_000_000)
		scanner.Buffer(buf, len(buf))
		var counter I
		for i = 0; i < bb.N; i++ {
			if !scanner.Scan() {
				scanner = bufio.NewScanner(bytes.NewBuffer(examples.Cache))
				scanner.Scan()
			}
			b := scanner.Bytes()
			ea := New()
			if b, err = ea.UnmarshalJSON(b); chk.E(err) {
				bb.Fatal(err)
			}
			evts[counter] = ea
			b = b[:0]
			if counter > 9999 {
				counter = 0
			}
		}
	})
	bb.Run("MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		var counter int
		out = out[:0]
		for i = 0; i < bb.N; i++ {
			out, _ = evts[counter].MarshalJSON(out)
			out = out[:0]
			counter++
			if counter != len(evts) {
				counter = 0
			}
		}
	})
	bins := make([]B, len(evts))
	bb.Run("MarshalBinary", func(bb *testing.B) {
		bb.ReportAllocs()
		var counter int
		for i = 0; i < bb.N; i++ {
			var b B
			b, _ = evts[counter].MarshalBinary(b)
			bins[counter] = b
			counter++
			if counter != len(evts) {
				counter = 0
			}
		}
	})
	bb.Run("UnmarshalBinary", func(bb *testing.B) {
		bb.ReportAllocs()
		var counter int
		ev := New()
		out = out[:0]
		for i = 0; i < bb.N; i++ {
			out = append(out, bins[counter]...)
			out, _ = ev.UnmarshalBinary(out)
			out = out[:0]
			counter++
			if counter != len(evts) {
				counter = 0
			}
		}
	})

}

func TestBinaryEvents(t *testing.T) {
	var err error
	var ev, ev2 *T
	var orig B
	b2, b3 := make(B, 0, 1_000_000), make(B, 0, 1_000_000)
	j2, j3 := make(B, 0, 1_000_000), make(B, 0, 1_000_000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	ev, ev2 = New(), New()
	for !scanner.Scan() {
		orig = scanner.Bytes()
		if orig, err = ev.UnmarshalJSON(orig); chk.E(err) {
			t.Fatal(err)
		}
		if len(orig) > 0 {
			t.Fatalf("remainder after end of event: %s", orig)
		}
		if b2, err = ev.MarshalBinary(b2); chk.E(err) {
			t.Fatal(err)
		}
		// copy for verification
		b3 = append(b3, b2...)
		if b2, err = ev2.UnmarshalBinary(b2); chk.E(err) {
			t.Fatal(err)
		}
		if len(b2) > 0 {
			t.Fatalf("remainder after end of event: %s", orig)
		}
		// bytes should be identical to b3
		if b2, err = ev2.MarshalBinary(b2); chk.E(err) {
			t.Fatal(err)
		}
		if equals(b2, b3) {
			log.E.S(ev, ev2)
			t.Fatalf("failed to remarshal\n%0x\n%0x",
				b3, b2)
		}
		j2, j3 = j2[:0], j3[:0]
		b2, b3 = b2[:0], b3[:0]
	}
}
