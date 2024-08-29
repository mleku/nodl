package ratel

import (
	"git.replicatr.dev/eventstore/ratel/keys/index"
	"git.replicatr.dev/eventstore/ratel/keys/serial"
	. "nostr.mleku.dev"
)

// GetCounterKey returns the proper counter key for a given event ID.
func GetCounterKey(ser *serial.T) (key B) {
	key = index.Counter.Key(ser)
	// Log.T.F("counter key %d %d", index.Counter, ser.Uint64())
	return
}
