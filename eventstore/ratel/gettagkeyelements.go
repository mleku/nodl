package ratel

import (
	"git.replicatr.dev/eventstore/ratel/keys"
	"git.replicatr.dev/eventstore/ratel/keys/arb"
	"git.replicatr.dev/eventstore/ratel/keys/createdat"
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/kinder"
	"git.replicatr.dev/eventstore/ratel/keys/pubkey"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	. "nostr.mleku.dev"
	"strconv"
	"strings"

	"ec.mleku.dev/v2/schnorr"
	"util.mleku.dev/hex"
)

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
			if pkk, err = pubkey.NewFromBytes(pkb); Chk.E(err) {
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
		if pkb, err = hex.Dec(split[1]); !Chk.E(err) {
			var kin uint64
			if kin, err = strconv.ParseUint(split[0], 10, 16); err == nil {
				k = uint16(kin)
				d = split[2]
				var pk *pubkey.T
				if pk, err = pubkey.NewFromBytes(pkb); Chk.E(err) {
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
