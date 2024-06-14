package eventid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/minio/sha256-simd"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type T struct {
	b []byte
}

func New() *T                 { return &T{b: make([]byte, 0, sha256.Size)} }
func (t *T) Bytes() []byte    { return t.b }
func (t *T) String() string   { return hex.EncodeToString(t.Bytes()) }
func (t *T) Reset()           { t.b = t.b[:0] }
func (t *T) Equal(t2 *T) bool { return bytes.Equal(t.b, t2.b) }

func NewFromBytes(b []byte) (t *T, err error) {
	if len(b) != sha256.Size {
		err = fmt.Errorf("eventid.NewFromBytes: invalid length %d require %d",
			len(b), sha256.Size)
	}
	t = &T{b: b}
	return
}

func (t *T) Set(b []byte) (err error) {
	if len(b) != sha256.Size {
		err = fmt.Errorf("eventid.NewFromBytes: invalid length %d require %d",
			len(b), sha256.Size)
	}
	t.b = b
	return
}
