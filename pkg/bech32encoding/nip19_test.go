package bech32encoding

import (
	"reflect"
	"testing"

	"github.com/mleku/nodl/pkg/eventid"
	"github.com/mleku/nodl/pkg/hex"
	"github.com/mleku/nodl/pkg/kind"
	"github.com/mleku/nodl/pkg/pointers"
)

func TestEncodeNpub(t *testing.T) {
	npub, err := HexToNpub(B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d"))
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(npub,
		B("npub180cvv07tjdrrgpa0j7j7tmnyl2yr6yr7l8j4s3evf6u64th6gkwsyjh6w6")) {
		t.Error("produced an unexpected npub string")
	}
}

func TestEncodeNsec(t *testing.T) {
	nsec, err := HexToNsec(B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d"))
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(nsec,
		B("nsec180cvv07tjdrrgpa0j7j7tmnyl2yr6yr7l8j4s3evf6u64th6gkwsgyumg0")) {
		t.Error("produced an unexpected nsec string")
	}
}

func TestDecodeNpub(t *testing.T) {
	prefix, pubkey, err := Decode(B("npub180cvv07tjdrrgpa0j7j7tmnyl2yr6yr7l8j4s3evf6u64th6gkwsyjh6w6"))
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(prefix, B("npub")) {
		t.Error("returned invalid prefix")
	}
	if !equals(pubkey.(B),
		B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d")) {
		t.Error("returned wrong pubkey")
	}
}

func TestFailDecodeBadChecksumNpub(t *testing.T) {
	_, _, err := Decode(B("npub180cvv07tjdrrgpa0j7j7tmnyl2yr6yr7l8j4s3evf6u64th6gkwsyjh6w4"))
	if err == nil {
		t.Errorf("should have errored: %s", err)
	}
}

func TestDecodeNprofile(t *testing.T) {
	prefix, data, err := Decode(B(
		"nprofile1qqsrhuxx8l9ex335q7he0f09aej04zpazpl0ne2cgukyawd24mayt8gpp4mhxue69uhhytnc9e3k7mgpz4mhxue69uhkg6nzv9ejuumpv34kytnrdaksjlyr9p"))
	if err != nil {
		t.Error("failed to decode nprofile")
	}
	if !equals(prefix, B("nprofile")) {
		t.Error("what")
	}
	pp, ok := data.(pointers.Profile)
	if !ok {
		t.Error("value returned of wrong type")
	}

	if !equals(pp.PublicKey,
		B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d")) {
		t.Error("decoded invalid public key")
	}

	if len(pp.Relays) != 2 {
		t.Error("decoded wrong number of relays")
	}
	if !equals(pp.Relays[0], B("wss://r.x.com")) ||
		!equals(pp.Relays[1], B("wss://djbas.sadkb.com")) {
		t.Error("decoded relay URLs wrongly")
	}
}

func TestDecodeOtherNprofile(t *testing.T) {
	prefix, data, err := Decode(B("nprofile1qqsw3dy8cpumpanud9dwd3xz254y0uu2m739x0x9jf4a9sgzjshaedcpr4mhxue69uhkummnw3ez6ur4vgh8wetvd3hhyer9wghxuet5qyw8wumn8ghj7mn0wd68yttjv4kxz7fww4h8get5dpezumt9qyvhwumn8ghj7un9d3shjetj9enxjct5dfskvtnrdakstl69hg"))
	if err != nil {
		t.Error("failed to decode nprofile")
	}
	if !equals(prefix, B("nprofile")) {
		t.Error("what")
	}
	pp, ok := data.(pointers.Profile)
	if !ok {
		t.Error("value returned of wrong type")
	}

	if !equals(pp.PublicKey,
		B("e8b487c079b0f67c695ae6c4c2552a47f38adfa2533cc5926bd2c102942fdcb7")) {
		t.Error("decoded invalid public key")
	}

	if len(pp.Relays) != 3 {
		t.Error("decoded wrong number of relays")
	}
	if !equals(pp.Relays[0], B("wss://nostr-pub.wellorder.net")) ||
		!equals(pp.Relays[1], B("wss://nostr-relay.untethr.me")) {

		t.Error("decoded relay URLs wrongly")
	}
}

func TestEncodeNprofile(t *testing.T) {
	nprofile, err := EncodeProfile(B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d"),
		[]B{
			B("wss://r.x.com"),
			B("wss://djbas.sadkb.com"),
		})
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(nprofile,
		B("nprofile1qqsrhuxx8l9ex335q7he0f09aej04zpazpl0ne2cgukyawd24mayt8gpp4mhxue69uhhytnc9e3k7mgpz4mhxue69uhkg6nzv9ejuumpv34kytnrdaksjlyr9p")) {
		t.Error("produced an unexpected nprofile string")
	}
}

func TestEncodeDecodeNaddr(t *testing.T) {
	var naddr B
	var err error
	naddr, err = EncodeEntity(
		B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d"),
		kind.Article,
		B("banana"),
		[]B{
			B("wss://relay.nostr.example.mydomain.example.com"),
			B("wss://nostr.banana.com"),
		})
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(naddr,
		B("naddr1qqrxyctwv9hxzqfwwaehxw309aex2mrp0yhxummnw3ezuetcv9khqmr99ekhjer0d4skjm3wv4uxzmtsd3jjucm0d5q3vamnwvaz7tmwdaehgu3wvfskuctwvyhxxmmdqgsrhuxx8l9ex335q7he0f09aej04zpazpl0ne2cgukyawd24mayt8grqsqqqa28a3lkds")) {
		t.Errorf("produced an unexpected naddr string: %s", naddr)
	}
	var prefix B
	var data any
	prefix, data, err = Decode(naddr)
	// log.D.S(prefix, data, e)
	if chk.D(err) {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(prefix, NentityHRP) {
		t.Error("returned invalid prefix")
	}
	ep, ok := data.(pointers.Entity)
	if !ok {
		t.Fatalf("did not decode an entity type, got %v", reflect.TypeOf(data))
	}
	if !equals(ep.PublicKey,
		B("3bf0c63fcb93463407af97a5e5ee64fa883d107ef9e558472c4eb9aaaefa459d")) {
		t.Error("returned wrong pubkey")
	}
	if ep.Kind.ToUint16() != kind.Article.ToUint16() {
		t.Error("returned wrong kind")
	}
	if !equals(ep.Identifier, B("banana")) {
		t.Error("returned wrong identifier")
	}
	if !equals(ep.Relays[0],
		B("wss://relay.nostr.example.mydomain.example.com")) ||
		!equals(ep.Relays[1], B("wss://nostr.banana.com")) {
		t.Error("returned wrong relays")
	}
}

func TestDecodeNaddrWithoutRelays(t *testing.T) {
	prefix, data, err := Decode(B("naddr1qq98yetxv4ex2mnrv4esygrl54h466tz4v0re4pyuavvxqptsejl0vxcmnhfl60z3rth2xkpjspsgqqqw4rsf34vl5"))
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}
	if !equals(prefix, B("naddr")) {
		t.Error("returned invalid prefix")
	}
	ep := data.(pointers.Entity)
	if !equals(ep.PublicKey,
		B("7fa56f5d6962ab1e3cd424e758c3002b8665f7b0d8dcee9fe9e288d7751ac194")) {
		t.Error("returned wrong pubkey")
	}
	if ep.Kind.ToUint16() != kind.Article.ToUint16() {
		t.Error("returned wrong kind")
	}
	if !equals(ep.Identifier, B("references")) {
		t.Error("returned wrong identifier")
	}
	if len(ep.Relays) != 0 {
		t.Error("relays should have been an empty array")
	}
}

func TestEncodeDecodeNEventTestEncodeDecodeNEvent(t *testing.T) {
	aut := B("7fa56f5d6962ab1e3cd424e758c3002b8665f7b0d8dcee9fe9e288d7751abb88")
	eid := B("45326f5d6962ab1e3cd424e758c3002b8665f7b0d8dcee9fe9e288d7751ac194")
	nevent, err := EncodeEvent(
		MustDecode(eid),
		[]B{B("wss://banana.com")}, aut,
	)
	if err != nil {
		t.Errorf("shouldn't error: %s", err.Error())
	}

	prefix, res, err := Decode(nevent)
	if err != nil {
		t.Errorf("shouldn't error: %s", err)
	}

	if !equals(prefix, B("nevent")) {
		t.Errorf("should have 'nevent' prefix, not '%s'", prefix)
	}
	ep, ok := res.(pointers.Event)
	if !ok {
		t.Errorf("'%s' should be an nevent, not %v", nevent, res)
	}

	if !equals(ep.Author, aut) {
		t.Errorf("wrong author got\n%s, expect\n%s", ep.Author, aut)
	}
	id := MustDecode("45326f5d6962ab1e3cd424e758c3002b8665f7b0d8dcee9fe9e288d7751ac194")
	if !ep.ID.Equal(id) {
		log.I.S(ep.ID, id)
		t.Error("wrong id")
	}

	if len(ep.Relays) != 1 ||
		!equals(ep.Relays[0], B("wss://banana.com")) {
		t.Error("wrong relay")
	}
}

func MustDecode[V S | B](s V) (ei *eventid.T) {
	var err error
	var b []byte
	if b, err = hex.Dec(string(s)); chk.E(err) {
		panic(err)
	}
	if ei, err = eventid.NewFromBytes(b); chk.E(err) {
		panic(err)
	}
	return
}
