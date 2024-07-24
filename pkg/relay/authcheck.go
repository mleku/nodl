package relay

import (
	authEnv "github.com/mleku/nodl/pkg/codec/envelopes/authenvelope"
	clEnv "github.com/mleku/nodl/pkg/codec/envelopes/closedenvelope"
	"github.com/mleku/nodl/pkg/protocol/auth"
	"github.com/mleku/nodl/pkg/util/normalize"
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
	if rl.Info.Limitation.AuthRequired && len(ws.AuthPubKey()) == 0 {
		reason := "this relay requires authentication for " + envType
		log.I.Ln(reason)
		chk.E(ws.WriteEnvelope(&clEnv.T{
			Reason: normalize.Reason(auth.Required, reason),
		}))
		// send out authorization request
		RequestAuth(c, envType)
		return false
	}
	return true
}

func RequestAuth(c Ctx, envType S) {
	ws := GetConnection(c)
	log.D.F("requesting auth from %s for %s", ws.RealRemote(), envType)
	// todo: check this works
	// ws.authLock.Lock()
	// if ws.Authed == nil {
	// 	ws.Authed = make(chan struct{})
	// }
	// ws.authLock.Unlock()
	chk.E(ws.WriteEnvelope(&authEnv.Challenge{Challenge: ws.Challenge()}))
}

func GetConnection(c Ctx) WS {
	v, ok := c.Value(wsKey).(WS)
	if !ok {
		return nil
	}
	return v
}
