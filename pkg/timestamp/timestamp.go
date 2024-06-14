package timestamp

import (
	"encoding/binary"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/mleku/nodl/pkg/utils/ints"
	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

// T is the value type which is used where
type T int64

// Now returns the current UNIX timestamp of the current second.
func Now() *T {
	t := T(time.Now().Unix())
	return &t
}

// U64 returns the current UNIX timestamp of the current second as uint64.
func (t *T) U64() uint64 { return uint64(*t) }

// I64 returns the current UNIX timestamp of the current second as int64.
func (t *T) I64() int64 { return int64(*t) }

// Int returns the timestamp as an int.
func (t *T) Int() int { return int(*t) }

func (t *T) Bytes() (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(*t))
	return
}

func (t *T) String() string { return strconv.Itoa(int(*t)) }

// FromTime returns a T from a time.Time
func FromTime(t time.Time) T { return T(t.Unix()) }

// FromUnix converts from a standard int64 unix timestamp.
func FromUnix(t int64) T { return T(t) }

// FromBytes converts from a string of raw bytes.
func FromBytes(b []byte) T { return T(binary.BigEndian.Uint64(b)) }

// MarshalJSON encodes a timestamp as ASCII decimal form as required for JSON.
func (t *T) MarshalJSON() (b []byte, err error) {
	// math.MaxInt64 has 19 ciphers in decimal form
	b = ints.Int64AppendToByteString(make([]byte, 0, 19), int64(*t))
	return
}

// UnmarshalJSON converts a timestamp, which is a decimal encoded as ASCII, by
// generating a place counter, multiplying by the ascii byte, minus 48 ('0') and
// then multiplying the place counter by 10, except for the last place.
//
// There is no byte slice equivalent and this function is a lot faster.
func (t *T) UnmarshalJSON(b []byte) (err error) {
	var n int64
	n, err = ints.ByteStringToInt64(b)
	*t = T(n)
	return
}

func (t *T) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0, 8)
	data = binary.AppendVarint(data, int64(*t))
	return
}

func (t *T) UnmarshalBinary(data []byte) (err error) {
	var n int
	var v int64
	v, n = binary.Varint(data)
	if n < 1 {
		return errors.New("failed to decode varint timestamp")
	}
	*t = T(v)
	return
}