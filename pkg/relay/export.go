package relay

import (
	"os"
	"sync"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/relay/eventstore/badger"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/index"
	bdb "github.com/dgraph-io/badger/v4"
)

// Export prints the JSON of all events or writes them to a file.
func (rl *R) Export(db *badger.Backend, filename string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	log.D.Ln("running export subcommand")
	b := make([]byte, MaxMessageSize)
	o := make([]byte, MaxMessageSize)
	var fh *os.File
	var err error
	if filename != "" {
		fh, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
		if chk.F(err) {
			os.Exit(1)
		}
	} else {
		fh = os.Stdout
	}
	prf := []byte{index.Event.B()}
	// first gather the last accessed timestamps
	chk.E(db.View(func(txn *bdb.Txn) (err error) {
		it := txn.NewIterator(bdb.IteratorOptions{Prefix: prf})
		var ev *event.T
		for it.Rewind(); it.ValidForPrefix(prf); it.Next() {
			select {
			case <-rl.Ctx.Done():
				return
			default:
			}
			// get the event
			if b, err = it.Item().ValueCopy(b); chk.E(err) {
				continue
			}
			if o, err = ev.UnmarshalBinary(o); chk.E(err) {
				continue
			}
			if _, err = fh.Write(ev.Serialize()); chk.E(err) {
				continue
			}
			if _, err = fh.Write([]byte("\n")); chk.E(err) {
				continue
			}
		}
		it.Close()
		return nil
	}))
}
