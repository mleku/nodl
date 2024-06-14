package tag

import (
	"encoding/binary"
	"os"

	"github.com/mleku/nodl/pkg/text"
	"github.com/mleku/nodl/pkg/utils/bytestring"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

// The tag position meanings so they are clear when reading.
const (
	Key = iota
	Value
	Relay
)

type T []*text.T

func NewFromByteStrings(b []bytestring.T) (t *T) {
	tt := make(T, 0, len(b))
	for i := range b {
		tt = append(tt, text.NewFromBytes(b[i]))
	}
	return &tt
}

func (t *T) Len() int                   { return len(*t) }
func (t *T) Element(n int) (el *text.T) { return (*t)[n] }

func (t *T) Equal(other *T) bool {
	if t.Len() != other.Len() {
		log.E.Ln("not same length", "expected", t.Len(), "actual", other.Len())
		return false
	}
	for i := range *t {
		if !t.Element(i).Equal(other.Element(i)) {
			return false
		}
	}
	return true
}

func (t *T) HasPrefix(prefix *T) bool {
	if t.Len() < prefix.Len() {
		return false
	}
	for i, p := range *prefix {
		if !p.Equal(t.Element(i)) {
			return false
		}
	}
	return true
}

func (t *T) Contains(b *text.T) bool {
	for i := range *t {
		if t.Element(i).Equal(b) {
			return true
		}
	}
	return false
}

func (t *T) Key() (tt *text.T) {
	if len(*t) > Key {
		tt = t.Element(Key)
	}
	return
}
func (t *T) Value() (tt *text.T) {
	if len(*t) > Value {
		tt = t.Element(Value)
	}
	return
}
func (t *T) Relay() (tt *text.T) {
	if len(*t) > Relay {
		tt = t.Element(Relay)
	}
	return
}

func (t *T) Clone() (tt *T) {
	tp := make(T, 0, len(*t))
	for i := range *t {
		tp = append(tp, t.Element(i).Clone())
	}
	return &tp
}

func (t *T) Slice() (s []bytestring.T) {
	for i := range *t {
		s = append(s, t.Element(i).Bytes())
	}
	return
}

func EstimateBinarySize(src *T) (size int) {
	// first data is the number of elements in the tag, 16 bits should be enough
	size += binary.MaxVarintLen16
	// next the lengths of each element including the length prefix
	for i := range *src {
		size += binary.MaxVarintLen32 + src.Element(i).Len()
	}
	return
}

func AppendBinary(dst []byte, src *T) (b []byte) {
	// if existing capacity is insufficient, allocate new and copy
	if cap(dst) < EstimateBinarySize(src)+len(dst) {
		tmp := make([]byte, EstimateBinarySize(src)+len(dst))
		copy(tmp, dst)
	}
	// first the tag element count
	dst = binary.AppendUvarint(dst, uint64(len(*src)))
	// then each field with its uvarint length prefix
	for i := range *src {
		dst = bytestring.Append(dst, (*src)[i].Bytes())
	}
	return dst
}

// ExtractBinary decodes the data based on the length prefix and returns a the the
// remaining data from the provided slice.
func ExtractBinary(b []byte) (t *T, rem []byte, err error) {
	bl := len(b)
	nl, read := binary.Uvarint(b)
	if read < 1 {
		err = errorf.E("failed to read uvarint length prefix")
		return
	}
	t = &T{}
	cur := read
	var l uint64
	for i := range nl {
		l, read = binary.Uvarint(b[cur:])
		if read < 1 {
			err = errorf.E("failed to read uvarint length prefix of field %d",
				i)
			return
		}
		cur += read
		if bl < cur+int(l) {
			err = errorf.E("insufficient data in buffer, require %d have %d",
				int(l)+read, len(b))
			return
		}
		*t = append(*t, text.NewFromBytes(b[cur:cur+int(l)]))
		cur += int(l)
	}
	rem = b[cur:]
	return
}
