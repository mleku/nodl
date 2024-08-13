package budger

import (
	"encoding/binary"
	"strconv"
	"strings"

	"ec.mleku.dev/v2/schnorr"
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/arb"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/kinder"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/pubkey"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
	"git.replicatr.dev/pkg/util/hex"
	"github.com/fiatjaf/eventstore"
)

func getTagIndexPrefix(tagValue string) ([]byte, int) {
	var k []byte   // the key with full length for created_at and idx at the end, but not filled with these
	var offset int // the offset -- i.e. where the prefix ends and the created_at and idx would start

	if kind, pkb, d := eventstore.GetAddrTagElements(tagValue); len(pkb) == 32 {
		// store value in the new special "a" tag index
		k = make([]byte, 1+2+8+len(d)+4+4)
		k[0] = indexTagAddrPrefix
		binary.BigEndian.PutUint16(k[1:], kind)
		copy(k[1+2:], pkb[0:8])
		copy(k[1+2+8:], d)
		offset = 1 + 2 + 8 + len(d)
	} else if vb, _ := hex.Dec(tagValue); len(vb) == 32 {
		// store value as bytes
		k = make([]byte, 1+8+4+4)
		k[0] = indexTag32Prefix
		copy(k[1:], vb[0:8])
		offset = 1 + 8
	} else {
		// store whatever as utf-8
		k = make([]byte, 1+len(tagValue)+4+4)
		k[0] = indexTagPrefix
		copy(k[1:], tagValue)
		offset = 1 + len(tagValue)
	}

	return k, offset
}

// func getIndexKeysForEvent(evt *event.T, idx []byte) [][]byte {
// 	keys := make([][]byte, 0, 18)
//
// 	// indexes
// 	{
// 		// ~ by id
// 		idPrefix8, _ := hex.DecodeString(evt.ID[0 : 8*2])
// 		k := make([]byte, 1+8+4)
// 		k[0] = indexIdPrefix
// 		copy(k[1:], idPrefix8)
// 		copy(k[1+8:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	{
// 		// ~ by pubkey+date
// 		pubkeyPrefix8, _ := hex.DecodeString(evt.PubKey[0 : 8*2])
// 		k := make([]byte, 1+8+4+4)
// 		k[0] = indexPubkeyPrefix
// 		copy(k[1:], pubkeyPrefix8)
// 		binary.BigEndian.PutUint32(k[1+8:], uint32(evt.CreatedAt))
// 		copy(k[1+8+4:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	{
// 		// ~ by kind+date
// 		k := make([]byte, 1+2+4+4)
// 		k[0] = indexKindPrefix
// 		binary.BigEndian.PutUint16(k[1:], uint16(evt.Kind))
// 		binary.BigEndian.PutUint32(k[1+2:], uint32(evt.CreatedAt))
// 		copy(k[1+2+4:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	{
// 		// ~ by pubkey+kind+date
// 		pubkeyPrefix8, _ := hex.DecodeString(evt.PubKey[0 : 8*2])
// 		k := make([]byte, 1+8+2+4+4)
// 		k[0] = indexPubkeyKindPrefix
// 		copy(k[1:], pubkeyPrefix8)
// 		binary.BigEndian.PutUint16(k[1+8:], uint16(evt.Kind))
// 		binary.BigEndian.PutUint32(k[1+8+2:], uint32(evt.CreatedAt))
// 		copy(k[1+8+2+4:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	// ~ by tagvalue+date
// 	for i, tag := range evt.Tags {
// 		if len(tag) < 2 || len(tag[0]) != 1 || len(tag[1]) == 0 || len(tag[1]) > 100 {
// 			// not indexable
// 			continue
// 		}
// 		firstIndex := slices.IndexFunc(evt.Tags, func(t nostr.Tag) bool { return len(t) >= 2 && t[1] == tag[1] })
// 		if firstIndex != i {
// 			// duplicate
// 			continue
// 		}
//
// 		// get key prefix (with full length) and offset where to write the last parts
// 		k, offset := getTagIndexPrefix(tag[1])
//
// 		// write the last parts (created_at and idx)
// 		binary.BigEndian.PutUint32(k[offset:], uint32(evt.CreatedAt))
// 		copy(k[offset+4:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	{
// 		// ~ by date only
// 		k := make([]byte, 1+4+4)
// 		k[0] = indexCreatedAtPrefix
// 		binary.BigEndian.PutUint32(k[1:], uint32(evt.CreatedAt))
// 		copy(k[1+4:], idx)
// 		keys = append(keys, k)
// 	}
//
// 	return keys
// }

// GetIndexKeysForEvent generates all the index keys required to filter for
// events. evtSerial should be the output of Serial() which gets a unique,
// monotonic counter value for each new event.
func GetIndexKeysForEvent(ev *event.T, ser *serial.T) (keyz [][]byte) {

	var err error
	keyz = make([][]byte, 0, 18)
	ID := id.New(eventid.NewWith(ev.ID))
	CA := createdat.New(ev.CreatedAt)
	K := kinder.New(ev.Kind.ToU16())
	PK, _ := pubkey.New(ev.PubKey)
	// indexes
	{ // ~ by id
		k := index.Id.Key(ID, ser)
		log.T.F("id key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+date
		k := index.Pubkey.Key(PK, CA, ser)
		log.T.F("pubkey + date key: %x %0x %0x %0x",
			k[0], k[1:9], k[9:17], k[17:])
		keyz = append(keyz, k)
	}
	{ // ~ by kind+date
		k := index.Kind.Key(K, CA, ser)
		log.T.F("kind + date key: %x %0x %0x %0x",
			k[0], k[1:3], k[3:11], k[11:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+kind+date
		k := index.PubkeyKind.Key(PK, K, CA, ser)
		log.T.F("pubkey + kind + date key: %x %0x %0x %0x %0x",
			k[0], k[1:9], k[9:11], k[11:19], k[19:])
		keyz = append(keyz, k)
	}
	// ~ by tag value + date
	for i, t := range ev.Tags.T {
		if len(t.Field) < 2 || // there is no value field
			// the tag is not a-zA-Z probably (this would permit arbitrary other
			// single byte chars)
			len(t.Field[0]) != 1 ||
			// the second field is zero length
			len(t.Field[1]) == 0 ||
			// the second field is more than 100 characters long
			len(t.Field[1]) > 100 {
			// any of the above is true then the tag is not indexable
			continue
		}
		var firstIndex int
		for firstIndex = range ev.Tags.T {
			if len(t.Field) >= 2 &&
				len(ev.Tags.T[firstIndex].Field) >= 2 &&
				len(ev.Tags.T) > 1 &&
				equals(t.Field[1],
					ev.Tags.T[1].Value()) {
				break
			}
		}
		// firstIndex := slices.IndexFunc(ev.Tags.T,
		// 	func(ti tag.T) bool {
		// 		return len(t) >= 2 && len(ti) >= 2 &&
		// 			ti[1] == t[1]
		// 	})
		if firstIndex != i {
			// duplicate
			continue
		}
		// get key prefix (with full length) and offset where to write the last
		// parts
		prf, elems := index.P(0), []keys.Element(nil)
		if prf, elems, err = GetTagKeyElements(S(t.Field[1]), CA, ser); chk.E(err) {
			return
		}
		k := prf.Key(elems...)
		log.T.F("tag '%s': %v key %x", t.Field[0], t.Field[1:], k)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := index.CreatedAt.Key(CA, ser)
		log.T.F("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}

func GetTagKeyElements(tagValue string, CA *createdat.T,
	ser *serial.T) (prf index.P,
	elems []keys.Element, err error) {

	var pkb []byte
	// first check if it might be a public key, fastest test
	if len(tagValue) == 2*schnorr.PubKeyBytesLen {
		// this could be a pubkey
		pkb, err = hex.Dec(tagValue)
		if err == nil {
			// it's a pubkey
			var pkk keys.Element
			if pkk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
				return
			}
			prf, elems = index.Tag32, keys.Make(pkk, ser)
			return
		}
	}
	// check for a tag
	if strings.Count(tagValue, ":") == 2 {
		// this means we will get 3 pieces here
		split := strings.Split(tagValue, ":")
		// middle element should be a public key so must be 64 hex ciphers
		if len(split[1]) != schnorr.PubKeyBytesLen*2 {
			return
		}
		var k uint16
		var d string
		if pkb, err = hex.Dec(split[1]); !chk.E(err) {
			var kin uint64
			if kin, err = strconv.ParseUint(split[0], 10, 16); err == nil {
				k = uint16(kin)
				d = split[2]
				var pk *pubkey.T
				if pk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
					return
				}
				prf = index.TagAddr
				elems = keys.Make(kinder.New(k), pk, arb.NewFromString(d), CA,
					ser)
				return
			}
		}
	}
	// store whatever as utf-8
	prf = index.Tag
	elems = keys.Make(arb.NewFromString(tagValue), CA, ser)
	return
}
