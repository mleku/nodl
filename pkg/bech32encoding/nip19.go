package bech32encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/minio/sha256-simd"

	"github.com/mleku/nodl/pkg/bech32"
	"github.com/mleku/nodl/pkg/ec/schnorr"
	"github.com/mleku/nodl/pkg/eventid"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/pointers"
)

var (
	NoteHRP     = B("note")
	NsecHRP     = B("nsec")
	NpubHRP     = B("npub")
	NprofileHRP = B("nprofile")
	NeventHRP   = B("nevent")
	NentityHRP  = B("naddr")
)

func DecodeToString(bech32String B) (prefix, value B, err error) {
	var s any
	if prefix, s, err = Decode(bech32String); chk.D(err) {
		return
	}
	var ok bool
	if value, ok = s.(B); ok {
		return
	}
	err = log.E.Err("value was not decoded to a string, found type %s",
		reflect.TypeOf(s))
	return
}

func Decode(bech32string B) (prefix B, value any, err error) {
	var bits5 []byte
	if prefix, bits5, err = bech32.DecodeNoLimit(bech32string); chk.D(err) {
		return
	}
	var data []byte
	if data, err = bech32.ConvertBits(bits5, 5, 8, false); chk.D(err) {
		return prefix, nil, fmt.Errorf("failed translating data into 8 bits: %s",
			err.Error())
	}
	switch {
	case equals(prefix, NpubHRP) ||
		equals(prefix, NsecHRP) ||
		equals(prefix, NoteHRP):
		if len(data) < 32 {
			return prefix, nil, fmt.Errorf("data is less than 32 bytes (%d)",
				len(data))
		}
		b := make(B, schnorr.PubKeyBytesLen*2)
		hex.EncBytes(b, data[:32])
		return prefix, b, nil
	case equals(prefix, NprofileHRP):
		var result pointers.Profile
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if len(result.PublicKey) < 1 {
					return prefix, result, fmt.Errorf("no pubkey found for nprofile")
				}
				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				if len(v) < 32 {
					return prefix, nil, fmt.Errorf("pubkey is less than 32 bytes (%d)",
						len(v))
				}
				result.PublicKey = make(B, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			default:
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	case equals(prefix, NeventHRP):
		var result pointers.Event
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if result.ID.Len() == 0 {
					return prefix, result, fmt.Errorf("no id found for nevent")
				}
				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				if len(v) < 32 {
					return prefix, nil, fmt.Errorf("id is less than 32 bytes (%d)",
						len(v))
				}
				result.ID, err = eventid.NewFromBytes(v)
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			case TLVAuthor:
				if len(v) < 32 {
					return prefix, nil, fmt.Errorf("author is less than 32 bytes (%d)",
						len(v))
				}
				result.Author = make(B, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.Author, v)
			case TLVKind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	case equals(prefix, NentityHRP):
		var result pointers.Entity
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if result.Kind.ToUint16() == 0 ||
					len(result.Identifier) < 1 ||
					len(result.PublicKey) < 1 {

					return prefix, result, fmt.Errorf("incomplete naddr")
				}

				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				result.Identifier = v
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			case TLVAuthor:
				if len(v) < 32 {
					return prefix, nil, fmt.Errorf("author is less than 32 bytes (%d)",
						len(v))
				}
				result.PublicKey = make(B, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case TLVKind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				log.D.Ln("got a bogus TLV type code", t)
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	}
	return prefix, data, fmt.Errorf("unknown tag %s", prefix)
}

func EncodeNote(eventIDHex B) (s B, err error) {
	var b []byte
	if _, err = hex.DecBytes(b, eventIDHex); chk.D(err) {
		err = log.E.Err("failed to decode event id hex: %w", err)
		return
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(b, 8, 5, true); chk.D(err) {
		return
	}
	return bech32.Encode(NoteHRP, bits5)
}

func EncodeProfile(publicKeyHex B, relays []B) (s B, err error) {
	buf := &bytes.Buffer{}
	pb := make(B, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, publicKeyHex); chk.D(err) {
		err = log.E.Err("invalid pubkey '%s': %w", publicKeyHex, err)
		return
	}
	writeTLVEntry(buf, TLVDefault, pb)
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}
	return bech32.Encode(NprofileHRP, bits5)
}

func EncodeEvent(eventIDHex *eventid.T, relays []B, author B) (s B, err error) {
	buf := &bytes.Buffer{}
	id := make(B, sha256.Size)
	if _, err = hex.DecBytes(id, eventIDHex.ByteString()); chk.D(err) ||
		len(id) != 32 {

		return nil, fmt.Errorf("invalid id %d '%s': %v", len(id), eventIDHex,
			err)
	}
	writeTLVEntry(buf, TLVDefault, id)
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	pubkey := make(B, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pubkey, author); len(pubkey) == 32 {
		writeTLVEntry(buf, TLVAuthor, pubkey)
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}

	return bech32.Encode(NeventHRP, bits5)
}

func EncodeEntity(pk B, k *kind.T, id B, relays []B) (s B, err error) {
	buf := &bytes.Buffer{}
	writeTLVEntry(buf, TLVDefault, []byte(id))
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	pb := make(B, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, pk); chk.D(err) {
		return nil, fmt.Errorf("invalid pubkey '%s': %w", pb, err)
	}
	writeTLVEntry(buf, TLVAuthor, pb)
	kindBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(kindBytes, uint32(k.ToUint16()))
	writeTLVEntry(buf, TLVKind, kindBytes)
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		return nil, fmt.Errorf("failed to convert bits: %w", err)
	}
	return bech32.Encode(NentityHRP, bits5)
}
