package main

import (
	"net/http"
	"sync"

	"git.replicatr.dev/pkg/relay"
	"git.replicatr.dev/pkg/util/interrupt"
	"git.replicatr.dev/pkg/util/lol"
)

const DefaultListener = "0.0.0.0:3334"

func main() {
	lol.SetLogLevel("trace")
	rl := relay.T{ListenAddresses: []S{DefaultListener}}.Init()
	var wg sync.WaitGroup
	for _, l := range rl.ListenAddresses {
		wg.Add(1)
		go func(l S) {
			log.I.F("listening on %s\n", l)
			srv := http.Server{Addr: l, Handler: rl}
			interrupt.AddHandler(func() { chk.E(srv.Close()) })
			_ = srv.ListenAndServe()
			wg.Done()
		}(l)
	}
	wg.Wait()
}
