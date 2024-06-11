package text

import (
	"encoding/binary"
	"os"

	text "github.com/mleku/nodl/pkg/text/escape"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk = lol.New(os.Stderr)

type T struct {
	b []byte
}

func New() *T { return &T{} }

func (t *T) AppendEscaped(dst []byte, quote bool) (b []byte) {
	if quote {
		dst = append(dst, '"')
	}
	dst = text.EscapeByteString(dst, t.b)
	if quote {
		dst = append(dst, '"')
	}
	return dst
}

func Unquote(b []byte) []byte { return b[1 : len(b)-2] }

func (t *T) MarshalJSON() (b []byte, err error) {
	// a reasonable estimate of how much the escaping will increase is about 30
	// characters between line breaks, but with quotes, and embedded JSON, maybe
	// better to go as much as 10% over to avoid reallocations.
	est := len(b)*11/10 + 2
	if cap(t.b) < est {
		t.b = make([]byte, 0, est)
	}
	b = append(b, '"')
	b = text.EscapeByteString(b, t.b)
	b = append(b, '"')
	return
}

func (t *T) UnmarshalJSON(b []byte) (err error) {
	text.UnescapeByteString(t.b, Unquote(b))
	return
}

func (t *T) Len() int { return len(t.b) }

func (t *T) AppendBinary(data []byte) []byte {
	data = binary.AppendUvarint(data, uint64(len(t.b)))
	data = append(data, t.b...)
	return data
}

// MarshalBinary appends a varint prefix length prefix and then the text bytes.
func (t *T) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0, t.Len()+4) // should never be more than 268Mb
	data = t.AppendBinary(data)
	return
}

// UnmarshalBinary expects a uvarint length prefix and then the specified amount
// of bytes.
func (t *T) UnmarshalBinary(data []byte) (err error) {
	l, read := binary.Uvarint(data)
	if read < 1 {
		return log.E.Err("failed to read uvarint length prefix")
	}
	if int(l)+read > len(data) {
		return log.E.Err("insufficient data in buffer to ")
	}
	t.b = data[read : read+int(l)]
	return
}
