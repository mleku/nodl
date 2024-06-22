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
	var rem B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		if ev, rem, err = Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		_, _ = ev, rem
	}
}
