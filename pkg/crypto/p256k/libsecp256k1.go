//go:build !btcec

package p256k

import "C"
import (
	"crypto/rand"
	"unsafe"

	"ec.mleku.dev/v2/schnorr"
	"ec.mleku.dev/v2/secp256k1"
	"github.com/minio/sha256-simd"
)

/*
#cgo LDFLAGS: -lsecp256k1
#include <secp256k1.h>
#include <secp256k1_schnorrsig.h>
#include <secp256k1_extrakeys.h>
#include <secp256k1_ecdh.h>
*/
import "C"

type (
	Context  = C.secp256k1_context
	Uchar    = C.uchar
	Cint     = C.int
	SecKey   = C.secp256k1_keypair
	PubKey   = C.secp256k1_xonly_pubkey
	ECPubKey = C.secp256k1_pubkey
)

var (
	ctx *Context
)

func CreateContext() *Context {
	return C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN |
		C.SECP256K1_CONTEXT_VERIFY)
}

func GetRandom() (u *Uchar) {
	rnd := make([]byte, 32)
	_, _ = rand.Read(rnd)
	return ToUchar(rnd)
}

func AssertLen(b B, length int, name string) (err E) {
	if len(b) != length {
		err = errorf.E("%s should be %d bytes, got %d", name, length, len(b))
	}
	return
}

func RandomizeContext(ctx *C.secp256k1_context) {
	C.secp256k1_context_randomize(ctx, GetRandom())
	return
}

func CreateRandomContext() (c *Context) {
	c = CreateContext()
	RandomizeContext(c)
	return
}

func init() {
	if ctx = CreateContext(); ctx == nil {
		panic("failed to create secp256k1 context")
	}
}

func ToUchar(b B) (u *Uchar) { return (*Uchar)(unsafe.Pointer(&b[0])) }

type Sec struct {
	Key SecKey
}

func GenSec() (sec *Sec, err E) {
	var skb B
	if skb, _, err = GenSecBytes(); chk.E(err) {
		return
	}
	return SecFromBytes(skb)
}

func SecFromBytes(sk B) (sec *Sec, err E) {
	sec = new(Sec)
	if C.secp256k1_keypair_create(ctx, &sec.Key, ToUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	return
}

func (s *Sec) Sec() *SecKey { return &s.Key }

func (s *Sec) Pub() (p *Pub, err E) {
	p = new(Pub)
	if C.secp256k1_keypair_xonly_pub(ctx, &p.Key, nil, s.Sec()) != 1 {
		err = errorf.E("pubkey derivation failed")
		return
	}
	return
}

type PublicKey struct {
	pk *C.secp256k1_pubkey
}

func newPublicKey() *PublicKey {
	return &PublicKey{
		pk: &C.secp256k1_pubkey{},
	}
}

type XPublicKey struct {
	pk *C.secp256k1_xonly_pubkey
}

func newXPublicKey() *XPublicKey {
	return &XPublicKey{
		pk: &C.secp256k1_xonly_pubkey{},
	}
}

func GenSecBytes() (sk32, pkb B, err error) {
	// because we must not create pubkeys with odd/0x03 prefix we have to derive the pubkey in this process anyway, and
	// so we must use the bitcoin-core/secp256k1 library or pay a performance penalty if we want to generate a lot of
	// new keys (for network transport MAC use case)
	sk32 = make(B, secp256k1.SecKeyBytesLen)
	for {
		var kp Sec
		kpk := newPublicKey()
		kpx := newXPublicKey()
		if _, err = rand.Read(sk32); chk.E(err) {
			return
		}
		usk32 := ToUchar(sk32)
		res := C.secp256k1_keypair_create(ctx, &kp.Key, usk32)
		if res != 1 {
			err = errorf.E("failed to create secp256k1 keypair")
			return
		}
		ecpkb := make(B, secp256k1.PubKeyBytesLenCompressed)
		clen := C.size_t(secp256k1.PubKeyBytesLenCompressed)
		pkb = make(B, schnorr.PubKeyBytesLen)
		var parity Cint
		C.secp256k1_keypair_xonly_pub(ctx, kpx.pk, &parity, &kp.Key)
		C.secp256k1_keypair_pub(ctx, kpk.pk, &kp.Key)
		C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, kpk.pk, C.SECP256K1_EC_COMPRESSED)
		C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(pkb), kpx.pk)
		if ecpkb[0] == 2 {
			break
		}
	}
	return
}

func (s *Sec) ECPub() (p *ECPub) {
	p = new(ECPub)

	return
}

type ECPub struct {
	Key ECPubKey
}

// ECPubFromSchnorrBytes converts a BIP-340 public key to its even standard 33 byte encoding.
//
// This function is for the purpose of getting a key to do ECDH.
func ECPubFromSchnorrBytes(pk B) (pub *ECPub, err E) {
	if err = AssertLen(pk, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = &ECPub{}
	p := append(B{0x02}, pk...)
	if C.secp256k1_ec_pubkey_parse(ctx, &pub.Key, ToUchar(p),
		secp256k1.PubKeyBytesLenCompressed) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", p)
		log.I.S(pub)
		return
	}
	return
}

type Pub struct {
	Key PubKey
}

func PubFromBytes(pk B) (pub *Pub, err E) {
	if err = AssertLen(pk, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = new(Pub)
	if C.secp256k1_xonly_pubkey_parse(ctx, &pub.Key, ToUchar(pk)) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pk)
		return
	}
	return
}

func (p *Pub) PubB() (b B) {
	b = make(B, schnorr.PubKeyBytesLen)
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), &p.Key)
	return
}

func (p *Pub) Pub() *PubKey { return &p.Key }

func (p *Pub) ToBytes() (b B, err E) {
	b = make(B, schnorr.PubKeyBytesLen)
	if C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), p.Pub()) != 1 {
		err = errorf.E("pubkey serialize failed")
		return
	}
	return
}

func Sign(msg *Uchar, sk *SecKey) (sig B, err E) {
	sig = make(B, schnorr.SignatureSize)
	c := CreateRandomContext()
	if C.secp256k1_schnorrsig_sign32(c, ToUchar(sig), msg, sk,
		GetRandom()) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

func SignFromBytes(msg, sk B) (sig B, err E) {
	var umsg *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	var sec *Sec
	if sec, err = SecFromBytes(sk); chk.E(err) {
		return
	}
	return Sign(umsg, sec.Sec())
}

func Msg(b B) (id *Uchar, err E) {
	if err = AssertLen(b, sha256.Size, "id"); chk.E(err) {
		return
	}
	id = ToUchar(b)
	return
}

func Sig(b B) (sig *Uchar, err E) {
	if err = AssertLen(b, schnorr.SignatureSize, "sig"); chk.E(err) {
		return
	}
	sig = ToUchar(b)
	return
}

func Verify(msg, sig *Uchar, pk *PubKey) (valid bool) {
	return C.secp256k1_schnorrsig_verify(ctx, sig, msg, 32, pk) == 1
}

func VerifyFromBytes(msg, sig, pk B) (err error) {
	var umsg, usig *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	if usig, err = Sig(sig); chk.E(err) {
		return
	}
	var pub *Pub
	if pub, err = PubFromBytes(pk); chk.E(err) {
		return
	}
	valid := Verify(umsg, usig, pub.Pub())
	if !valid {
		err = errorf.E("failed to verify signature")
	}
	return
}

func Zero(sk *SecKey) {
	b := (*[96]byte)(unsafe.Pointer(sk))[:96]
	for i := range b {
		b[i] = 0
	}
}

func ECDH(sec *Uchar, pk *ECPub) (secret B) {
	secret = make(B, sha256.Size)
	uSecret := ToUchar(secret)
	C.secp256k1_ecdh(ctx, uSecret, &pk.Key, sec, nil, nil)
	return
}
