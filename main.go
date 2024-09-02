package main

import (
	"os"
	"sync"

	"git.replicatr.dev/eventstore/ratel"
	"git.replicatr.dev/relay"
	"git.replicatr.dev/relay/basic"
	"github.com/kelseyhightower/envconfig"
	. "nostr.mleku.dev"
	"util.mleku.dev/context"
	"util.mleku.dev/interrupt"
	"util.mleku.dev/lol"
	"util.mleku.dev/units"
)

const envPrefix = "replicatr"

func main() {
	dbg := os.Getenv("DEBUG")
	if dbg == "" {
		dbg = "info"
	}
	lol.SetLogLevel(dbg)
	var err E
	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()
	r := basic.New()
	if err = envconfig.Process(envPrefix, r); Chk.E(err) {
		Log.F.F("failed to read from env: %v", err)
		return
	}
	c, cancel := context.Cancel(context.Bg())
	wg := &sync.WaitGroup{}
	r.Store = ratel.GetBackend(c, wg, r.Path(), false, units.Gb*4,
		int(lol.Level.Load()), 512)
	var server *relay.Server
	if err = r.Init();Chk.E(err){
		os.Exit(1)
	}
	server, err = relay.NewServer(r)
	if err != nil {
		Log.F.F("failed to create server: %v", err)
	}
	interrupt.AddHandler(func() {
		Log.I.Ln("interrupt")
		server.Shutdown(c)
		// cancel()
		_ = cancel
	})
	Log.I.F("server listening on %s:%d", r.Listener, r.Port)
	if err = server.Start(r.Listener, r.Port); err != nil {
		Log.F.F("server terminated: %v", err)
	}
}
