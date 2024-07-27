package badger

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/eventid"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/createdat"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/id"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/kinder"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/pubkey"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
)

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
		// log.T.F("id key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+date
		k := index.Pubkey.Key(PK, CA, ser)
		// log.T.F("pubkey + date key: %x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:17], k[17:])
		keyz = append(keyz, k)
	}
	{ // ~ by kind+date
		k := index.Kind.Key(K, CA, ser)
		// log.T.F("kind + date key: %x %0x %0x %0x",
		// 	k[0], k[1:3], k[3:11], k[11:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+kind+date
		k := index.PubkeyKind.Key(PK, K, CA, ser)
		// log.T.F("pubkey + kind + date key: %x %0x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:11], k[11:19], k[19:])
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
				equals(t.Field[1], ev.Tags.T[1].Value()) {
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
		if prf, elems, err = GetTagKeyElements(S(t.Field[1]), CA, ser); err != nil {
			return
		}
		k := prf.Key(elems...)
		// log.T.F("tag '%s': %v key %x", t[0], t[1:], k)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := index.CreatedAt.Key(CA, ser)
		// log.T.F("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}
