package hex

import (
	"encoding/hex"
	"os"

	"github.com/mleku/nodl/pkg/utils/lol"
)

var log, chk, errorf = lol.New(os.Stderr)

type InvalidByteError = hex.InvalidByteError

var Enc = hex.EncodeToString
var Dec = hex.DecodeString
var DecLen = hex.DecodedLen
var Decode = hex.Decode

const (
	// if the 4 bytes are below 10, the ascii cipher is '0' + value
	n byte = '0'
	// if the 4 bytes are above 9, the ascii cipher is 'a' - 10 + value
	l byte = 'a' - 10
	// for converting back, the offset is reversed
	l2 byte = 'a' + 10
)

// AppendHexToByteString appends the hex ASCII encoding of lower-case
// hexadecimal to a given slice.
//
// if the slice capacity is insufficient, a new slice is allocated and the old
// data copied in.
//
// Rather than use a lookup table, the values are simply derived by appending
// offsets from the relevant ASCII cipher code, for 0-9, add the base of '0',
// for the values from a-f, add the base of 'a'-10.
//
// This should all be compiled to assembly code that performs the operation on
// each byte within the limited 4 registers of an x86 processor, eliminating any
// need for indirections to the stack.
//
// The encoding is for lower case only, as is the counterpart
// ByteStringToBytes, because nostr standard encoding for hex strings is
// lower case.
func AppendHexToByteString(dst, src []byte) (b []byte) {
	lb, cb := len(dst), cap(dst)
	if cb-lb < len(src)/2 {
		tmp := make([]byte, len(dst)+len(src)*2)
		copy(tmp, dst)
		dst = tmp
	}
	for i := range src {
		var v [2]byte
		v[1] = src[i] >> 4
		v[0] = src[i] & 0xf
		if v[0] < 10 {
			v[0] += n
		} else {
			v[0] += l
		}
		if v[1] < 10 {
			v[1] += n
		} else {
			v[1] += l
		}
		dst = append(dst, v[1], v[0])
	}
	return dst
}

// ByteStringToBytes performs the decoding of hex encoded ASCII bytes
// in-place and returns the decoded raw bytes, avoiding a memory allocation
// altogether.
func ByteStringToBytes(h []byte) (b []byte, err error) {
	if len(h)%2 != 0 {
		err = errorf.E("invalid length of hex, got %d, must be even", len(h))
		return
	}
	for i := 0; i < len(h); i += 2 {
		var v, bb byte
		// check that both characters are valid hex ciphers
		for j := 0; j < 2; j++ {
			bb = h[i+j]
			if bb >= '0' && bb <= '9' {
				// log.I.F("%c %x", bb, bb-n<<(j*4))
				v += (bb - n) << ((1 - j) * 4)
			} else if bb >= 'a' && bb <= 'f' {
				// log.I.F("%c %x", bb, (bb-l)<<(j*4))
				v += (bb - l) << ((1 - j) * 4)
			} else {
				err = errorf.E("invalid hex char %c at %d in %s", h[i], i+j,
					string(h))
				return
			}
		}
		h[i/2] = v
	}
	// truncate the slice
	b = h[:len(h)/2]
	return
}
