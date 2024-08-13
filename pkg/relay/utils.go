package relay

import (
	"git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	"git.replicatr.dev/pkg/protocol/relayws"
	"git.replicatr.dev/pkg/util/context"
)

const (
	wsKey = iota
	subscriptionIdKey
	nip86HeaderAuthKey
)

func RequestAuth(c context.T) {
	ws := GetConnection(c)
	ws.AuthLock.Lock()
	if ws.Authed == nil {
		ws.Authed = make(chan struct{})
	}
	ws.AuthLock.Unlock()
	ws.WriteEnvelope(authenvelope.NewChallengeWith(ws.Challenge()))
}

func GetConnection(ctx context.T) *relayws.WS {
	wsi := ctx.Value(wsKey)
	if wsi != nil {
		return wsi.(*relayws.WS)
	}
	return nil
}

func GetAuthed(c context.T) B {
	if conn := GetConnection(c); conn != nil {
		return conn.AuthPubKey.Load().(B)
	}
	if nip86Auth := c.Value(nip86HeaderAuthKey); nip86Auth != nil {
		return nip86Auth.(B)
	}
	return B{}
}

func GetIP(c context.T) S {
	return GetIPFromRequest(GetConnection(c).Request)
}

func GetSubscriptionID(c context.T) S {
	return c.Value(subscriptionIdKey).(S)
}
