package main

import (
	"net/http"
	"os"

	"git.replicatr.dev/relay"
	. "nostr.mleku.dev"

	"util.mleku.dev/interrupt"
	"util.mleku.dev/lol"
)

const DefaultListener = "0.0.0.0:3334"
const Path = ".replicatr"

func main() {
	lol.SetLogLevel("trace")
	var err E
	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()
	// var path S
	// if path, err = os.MkdirTemp("", "replicatr"); Chk.E(err) {
	// 	return
	// }
	path := Path
	var rl *relay.T
	rl, err = relay.T{
		ListenAddresses: []S{
			DefaultListener,
			// "10.0.0.2:4869",
		},
		Tracker: relay.Tracker{

		},
	}.Init(path)
	rl.WG.Add(1)
	for _, l := range rl.ListenAddresses {
		rl.WG.Add(1)
		go func(l S) {
			Log.I.F("listening on %s", l)
			srv := http.Server{Addr: l, Handler: rl}
			interrupt.AddHandler(func() { Chk.E(srv.Close()) })
			_ = srv.ListenAndServe()
			rl.WG.Done()
		}(l)
	}
	interrupt.AddHandler(func() {
		rl.Cancel()
		rl.WG.Done()
		if err = os.RemoveAll(path); Chk.E(err) {
			return
		}
	})
	rl.WG.Wait()
}
