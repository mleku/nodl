package relay

import (
	bdb "git.replicatr.dev/pkg/relay/eventstore/badger"
)

// Wipe clears the badgerDB local event store/cache.
func (rl *R) Wipe(store *bdb.Backend) (err error) {
	if err = store.Wipe(); chk.E(err) {
		return
	}
	return
}
