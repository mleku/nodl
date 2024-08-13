package main

import (
	"net/http"

	"git.replicatr.dev/pkg/relay"
	"git.replicatr.dev/pkg/relay/eventstore/budger"
	"git.replicatr.dev/pkg/util/lol"
)

func main() {
	lol.SetLogLevel("trace")
	relay := relay.New()
	db := budger.BadgerBackend{Path: "/tmp/khatru-badgern-tmp"}
	if err := db.Init(); err != nil {
		panic(err)
	}
	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	log.I.Ln("running on :3334")
	chk.F(http.ListenAndServe(":3334", relay))
}
