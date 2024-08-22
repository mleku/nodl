package badger

import (
	"git.replicatr.dev/pkg/relay/eventstore"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/arb"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/kinder"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/pubkey"
	"util.mleku.dev/hex"
)

// GetTagKeyPrefix returns tag index prefixes based on the initial field of a
// tag.
//
// There is 3 types of index tag keys:
//
// - TagAddr:   [ 8 ][ 2b Kind ][ 8b Pubkey ][ address/URL ][ 8b Serial ]
// - Tag32:     [ 7 ][ 8b Pubkey ][ 8b Serial ]
// - Tag:       [ 6 ][ address/URL ][ 8b Serial ]
//
// This function produces the initial bytes without the index.
func GetTagKeyPrefix(tagValue string) (key []byte, err error) {
	if k, pkb, d := eventstore.GetAddrTagElements(tagValue); len(pkb) == 32 {
		// store value in the new special "a" tag index
		var pk *pubkey.T
		if pk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		els := []keys.Element{kinder.New(k), pk}
		if len(d) > 0 {
			els = append(els, arb.NewFromString(d))
		}
		key = index.TagAddr.Key(els...)
	} else if pkb, _ := hex.Dec(tagValue); len(pkb) == 32 {
		// store value as bytes
		var pkk *pubkey.T
		if pkk, err = pubkey.NewFromBytes(pkb); chk.E(err) {
			return
		}
		key = index.Tag32.Key(pkk)
	} else {
		// store whatever as utf-8
		if len(tagValue) > 0 {
			var a *arb.T
			a = arb.NewFromString(tagValue)
			key = index.Tag.Key(a)
		}
		key = index.Tag.Key()
	}
	return
}
