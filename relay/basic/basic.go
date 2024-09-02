package basic

import (
	"git.replicatr.dev/eventstore"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"util.mleku.dev/appdata"
	"util.mleku.dev/context"
)

const DefaultProfile = ".replicatr"
const DefaultListener = "0.0.0.0"
const DefaultPort = 3334

type Relay struct {
	Profile, Listener S
	Port              N
	Store             eventstore.I
}

func New() *Relay {
	return &Relay{Profile: DefaultProfile, Listener: DefaultListener, Port: DefaultPort}
}

func (r *Relay) Name() S               { return "BasicRelay" }
func (r *Relay) Storage() eventstore.I { return r.Store }
func (r *Relay) Path() S               { return r.Profile }
func (r *Relay) Init() (err E) {
	r.Profile = appdata.Dir(r.Profile, false)
	err = r.Store.Init(r.Path())
	return
}
func (r *Relay) AcceptEvent(c context.T, evt *event.T) (ok bool) {
	return
}
