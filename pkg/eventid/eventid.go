package eventid

import (
	"crypto/sha1"
	"encoding/hex"
	"os"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk = lol.New(os.Stderr)

type T struct {
	b []byte
}

func New(b []byte) *T { return &T{b: make([]byte, 0, sha256.Size)} }

func AppendFromBinary(h []byte, quote bool) (b []byte) {
	if quote {
		h = append(h, '"')
		b = hex.AppendEncode(b, h)
		h = append(h, '"')
	} else {
		b = hex.AppendEncode(b, h)
	}
	return
}

func (t *T) MarshalJSON() (b []byte, err error) {
	b = make([]byte, 0, sha256.Size*2+2)
	b = AppendFromBinary(t.b, true)
	return
}

func (t *T) UnmarshalJSON(b []byte) (err error) {
	if len(b) < sha1.Size*2+2 {
		return log.E.Err("eventid: not enough bytes got %d required %d",
			len(b), sha1.Size+2)
	}
	// reset the slice
	t.b = t.b[:0]
	if t.b, err = hex.AppendDecode(t.b, b[1:len(b)-1]); chk.E(err) {
		return
	}
	return
}

func (t *T) MarshalBinary() (data []byte, err error) { return t.b, nil }

func (t *T) UnmarshalBinary(data []byte) (err error) {
	if len(data) < sha1.Size {
		return log.E.Err("eventid: not enough bytes got %d required %d",
			len(data), sha1.Size)
	}
	t.b = data
	return
}
