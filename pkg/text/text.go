package text

import (
	"bytes"
	"encoding/binary"
	"os"
	"unsafe"

	text "github.com/mleku/nodl/pkg/text/escape"
	"github.com/mleku/nodl/pkg/utils/bytestring"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type T struct {
	b []byte
}

func New() *T { return &T{} }

func NewFromBytes(b []byte) *T { return &T{b: b} }
func (t *T) Bytes() []byte     { return t.b }
func (t *T) String() string    { return string(t.b) }

// SetStringImmutable uses unsafe to turn the input string to a byte slice with
// the limitation that the resultant state of the text.T cannot be mutated.
//
// The slice can be copied freely of course. It is possible to use MarshalJSON,
// for example, after calling this to pull in a string.
func (t *T) SetStringImmutable(s string) {
	d := unsafe.StringData(s)
	t.b = unsafe.Slice(d, len(s))
}

// SetString copies a string into the bytes, which can then be mutated, in
// contrast to SetStringImmutable.
func (t *T) SetString(s string) {
	t.b = make([]byte, len(s))
	copy(t.b, s)
}

func (t *T) Set(b []byte) { t.b = b }

func AppendFromBinary(dst, src []byte) (b []byte) {
	dst = append(dst, src...)
	return dst
}

func (t *T) Equal(t2 *T) bool { return bytes.Equal(t.b, t2.b) }

func (t *T) MarshalJSON() (b []byte, err error) {
	// a reasonable estimate of how much the escaping will increase is about 30
	// characters between line breaks, but with quotes, and embedded JSON, maybe
	// better to go as much as 10% over to avoid reallocations.
	est := len(b)*11/10 + 2
	if cap(t.b) < est {
		t.b = make([]byte, 0, est)
	}
	b = bytestring.AppendQuote(b, t.b, text.NostrEscape)
	return
}

func (t *T) UnmarshalJSON(b []byte) (err error) {
	t.b = text.NostrUnescape(t.b, bytestring.Unquote(b))
	return
}

func (t *T) Len() int { return len(t.b) }

func (t *T) Append(dst []byte) (b []byte) {
	dst = binary.AppendUvarint(dst, uint64(len(t.b)))
	dst = append(dst, t.b...)
	b = dst
	return
}

// MarshalBinary appends a varint prefix length prefix and then the text bytes.
func (t *T) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0, t.Len()+4) // should never be more than 268Mb
	data = t.Append(data)
	return
}

// UnmarshalBinary expects a uvarint length prefix and then the specified amount
// of bytes.
func (t *T) UnmarshalBinary(data []byte) (err error) {
	l, read := binary.Uvarint(data)
	if read < 1 {
		return errorf.E("failed to read uvarint length prefix")
	}
	if int(l)+read > len(data) {
		return errorf.E("insufficient data in buffer, expect %d have %d",
			int(l)+read, len(data))
	}
	t.b = data[read : read+int(l)]
	return
}
