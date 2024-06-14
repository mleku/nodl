package text

import (
	"bytes"
	"os"
	"unsafe"

	"github.com/mleku/nodl/pkg/utils/bytestring"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type T struct {
	b bytestring.T
}

func New() *T                                      { return &T{} }
func NewFromBytes[V bytestring.T | []byte](b V) *T { return &T{b: bytestring.T(b)} }
func (t *T) Bytes() []byte                         { return t.b }
func (t *T) String() string                        { return string(t.b) }
func (t *T) Len() int                              { return len(t.b) }

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

// Clone duplicates the buffer into a new T.
func (t *T) Clone() (tt *T) {
	t.b = make([]byte, len(t.b))
	tt = &T{b: make(bytestring.T, len(t.b))}
	copy(tt.b, t.b)
	return
}

func (t *T) Set(b []byte) { t.b = b }

func (t *T) Equal(t2 *T) bool { return bytes.Equal(t.b, t2.b) }
