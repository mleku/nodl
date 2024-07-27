package relay

import (
	authEnv "git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	clEnv "git.replicatr.dev/pkg/codec/envelopes/closedenvelope"
	"git.replicatr.dev/pkg/codec/subscriptionid"
)

// AuthCheck sends out a request if auth is required (this is an OnConnects
// method). It just asks for auth if enabled, saving the client time waiting
// until after sending a req.
func (rl *R) AuthCheck(c Ctx) { rl.IsAuthed(c, "connect") }

func (rl *R) IsAuthed(c Ctx, envType S) bool {
	ws := GetConnection(c)
	if ws == nil {
		panic("how can has no websocket?")
	}
	// if access requires auth, check that auth is present.
	if rl.Info.Limitation.AuthRequired && ws.HasAuth() {
		reason := "this relay requires authentication for " + envType
		var err E
		if err = clEnv.NewFrom(&subscriptionid.T{}, Reason(AuthRequired, reason)).Write(ws); chk.E(err) {
			return false
		}
		// send out authorization request
		RequestAuth(c, envType)
		return false
	}
	return true
}

func RequestAuth(c Ctx, envType S) {
	ws := GetConnection(c)
	log.D.F("requesting auth from %s for %s", ws.Remote(), envType)
	// todo: check this works
	// ws.authLock.Lock()
	// if ws.Authed == nil {
	// 	ws.Authed = make(chan struct{})
	// }
	// ws.authLock.Unlock()
	chk.E(authEnv.NewChallengeWith(ws.Challenge()).Write(ws))
}

func GetConnection(c Ctx) WS {
	v, ok := c.Value(wsKey).(WS)
	if !ok {
		return nil
	}
	return v
}
