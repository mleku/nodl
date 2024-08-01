package encryption

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"

	"ec.mleku.dev/v2/secp256k1"
	"git.replicatr.dev/pkg/crypto/p256k"
	"github.com/minio/sha256-simd"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/hkdf"
)

const (
	version          byte = 2
	MinPlaintextSize      = 0x0001 // 1b msg => padded to 32b
	MaxPlaintextSize      = 0xffff // 65535 (64kb-1) => padded to 64kb
)

type Options struct {
	err   error
	nonce []byte
}

// Deprecated: use WithCustomNonce instead of WithCustomSalt, so the naming is less confusing
var WithCustomSalt = WithCustomNonce

func WithCustomNonce(nonce B) func(opts *Options) {
	return func(opts *Options) {
		if len(nonce) != 32 {
			opts.err = errors.New("salt must be 32 bytes")
		}
		opts.nonce = nonce
	}
}

// notBetween returns true if x is not between a and b, inclusive.
func notBetween(x, a, b int) bool {
	return x < a || x > b
}

func Encrypt(plain B, key B, opts ...func(opts *Options)) (b64ciphertext S, err E) {
	eo := Options{}
	for _, apply := range opts {
		if apply(&eo); chk.E(eo.err) {
			err = eo.err
			return
		}
	}
	if eo.nonce == nil {
		eo.nonce = make(B, 32)
		if _, err = rand.Read(eo.nonce); chk.E(err) {
			return
		}
	}
	var enc, cc20nonce, auth B
	if enc, cc20nonce, auth, err = messageKeys(key, eo.nonce); chk.E(err) {
		return
	}
	l := len(plain)
	if notBetween(l, MinPlaintextSize, MaxPlaintextSize) {
		err = errorf.E("plaintext must be between %d and %d", MinPlaintextSize, MaxPlaintextSize)
		return
	}
	padding := make([]byte, calcPadding(l))
	// 2 byte length prefix for ext
	binary.BigEndian.PutUint16(padding, uint16(l))
	copy(padding[2:], plain)
	var ciphertext B
	if ciphertext, err = ChaCha20Encipher(enc, cc20nonce, padding); chk.E(err) {
		return
	}
	var mac B
	if mac, err = sha256Hmac(auth, ciphertext, eo.nonce); chk.E(err) {
		return
	}
	msg := make([]byte, 0, 1+32+len(ciphertext)+32)
	msg = append(msg, version)
	msg = append(msg, eo.nonce...)
	msg = append(msg, ciphertext...)
	msg = append(msg, mac...)
	return base64.StdEncoding.EncodeToString(msg), nil
}

func Decrypt(b64ciphertext S, key B) (plaintext S, err E) {
	cLen := len(b64ciphertext)
	if notBetween(cLen, 132, 87472) {
		err = errorf.E("invalid payload length: %d", cLen)
		return
	}
	if b64ciphertext[:1] == "#" {
		return "", errors.New("unknown version")
	}
	var decoded B
	if decoded, err = base64.StdEncoding.DecodeString(b64ciphertext); chk.E(err) {
		return
	}
	if decoded[0] != version {
		err = errorf.E("unknown version %d", decoded[0])
		return
	}
	dLen := len(decoded)
	if notBetween(dLen, 99, 65603) {
		err = errorf.E(fmt.Sprintf("invalid data length: %d", dLen))
		return
	}
	nonce, ciphertext, givenMac := decoded[1:33], decoded[33:dLen-32], decoded[dLen-32:]
	var enc, cc20nonce, auth B
	if enc, cc20nonce, auth, err = messageKeys(key, nonce); chk.E(err) {
		return
	}
	var expectedMac B
	if expectedMac, err = sha256Hmac(auth, ciphertext, nonce); chk.E(err) {
		return
	}
	if !bytes.Equal(givenMac, expectedMac) {
		err = errorf.E("invalid hmac")
		return
	}
	var padded B
	if padded, err = ChaCha20Encipher(enc, cc20nonce, ciphertext); chk.E(err) {
		return
	}
	unpaddedLen := int(binary.BigEndian.Uint16(padded[:2]))
	if notBetween(unpaddedLen, MinPlaintextSize, MaxPlaintextSize) || len(padded) != calcPadding(unpaddedLen) {
		err = errorf.E("invalid padding")
		return
	}
	unpadded := padded[2:][:unpaddedLen]
	if len(unpadded) == 0 || len(unpadded) != int(unpaddedLen) {
		return "", errors.New("invalid padding")
	}
	return string(unpadded), nil
}

func MustBytes(s S) (b B) {
	if len(s)%2 != 0 {
		panic("hex string must be even")
	}
	b = make(B, len(s)/2)
	var err E
	if _, err = hex.Decode(b, B(s)); chk.E(err) {
		panic(err)
	}
	return
}

var secp256k1OrderBytes = MustBytes("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141")
var zeroBytes = make(B, secp256k1.SecKeyBytesLen)

func GenerateConversationKeyFromBytes(pkb, skb B) (key B, err E) {
	if bytes.Compare(skb, secp256k1OrderBytes) >= 0 || bytes.Equal(skb, zeroBytes) {
		err = errorf.E("invalid private key: x coordinate %s is not on the secp256k1 curve", skb)
		return
	}
	signer := &p256k.Signer{}
	if err = signer.InitSec(skb); chk.E(err) {
		return
	}
	var secret B
	if secret, err = signer.ECDH(pkb); chk.E(err) {
		return
	}
	return hkdf.Extract(sha256.New, secret, []byte("nip44-v2")), nil
}

func ChaCha20Encipher(key, nonce, message B) (dst B, err E) {
	var cipher *chacha20.Cipher
	if cipher, err = chacha20.NewUnauthenticatedCipher(key, nonce); chk.E(err) {
		return
	}
	dst = make(B, len(message))
	cipher.XORKeyStream(dst, message)
	return
}

func sha256Hmac(key, ciphertext, nonce B) (hash B, err E) {
	if len(nonce) != 32 {
		return nil, errors.New("nonce aad must be 32 bytes")
	}
	h := hmac.New(sha256.New, key)
	h.Write(nonce)
	h.Write(ciphertext)
	return h.Sum(nil), nil
}

func messageKeys(key, nonce B) (enc, cc20nonce, auth B, err E) {
	if len(key) != 32 {
		err = errorf.E("conversation key must be 32 bytes\n%0x", key)
		return
	}
	if len(nonce) != 32 {
		err = errorf.E("nonce must be 32 bytes")
		return
	}
	r := hkdf.Expand(sha256.New, key, nonce)
	enc = make(B, 32)
	if _, err = io.ReadFull(r, enc); err != nil {
		return nil, nil, nil, err
	}
	cc20nonce = make(B, 12)
	if _, err = io.ReadFull(r, cc20nonce); err != nil {
		return nil, nil, nil, err
	}
	auth = make(B, 32)
	if _, err = io.ReadFull(r, auth); err != nil {
		return nil, nil, nil, err
	}
	return enc, cc20nonce, auth, nil
}

func calcPadding(sLen int) int {
	if sLen <= 32 {
		return 32 + 2
	}
	nextPower := 1 << int(math.Floor(math.Log2(float64(sLen-1)))+1)
	chunk := int(math.Max(32, float64(nextPower/8)))
	return chunk*int(math.Floor(float64((sLen-1)/chunk))+1) + 2 // 2 byte length prefix
}
