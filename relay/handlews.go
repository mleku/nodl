package relay

import (
	"net/http"
	"time"

	. "nostr.mleku.dev"
	"nostr.mleku.dev/protocol/ws"

	W "github.com/fasthttp/websocket"
	"nostr.mleku.dev/codec/envelopes"
	"nostr.mleku.dev/codec/envelopes/authenvelope"
	"nostr.mleku.dev/codec/envelopes/closeenvelope"
	"nostr.mleku.dev/codec/envelopes/countenvelope"
	"nostr.mleku.dev/codec/envelopes/eventenvelope"
	"nostr.mleku.dev/codec/envelopes/okenvelope"
	"nostr.mleku.dev/codec/envelopes/reqenvelope"
	C "util.mleku.dev/context"
	"util.mleku.dev/normalize"
)

func (rl *T) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	Log.T.F("HandleWebsocket inbound connection from %s", r.RemoteAddr)
	var err E
	var conn *W.Conn
	conn, err = rl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	c, cancel := C.Cancel(C.Bg())
	serv := ws.New(c, conn, r, MaxMessageSize)
	Log.T.Ln("upgraded websocket", serv.Remote())
	rl.Tracker.Do(func() { rl.AddWS(serv) })
	Log.T.F("established websocket connection with %s", serv.Remote())
	go rl.wsReadMessages(serv, cancel)
	rl.wsWatcher(serv, cancel)
}

func (rl *T) wsWatcher(ws *ws.Serv, cancel C.F) {
	for {
		select {
		case <-rl.Ctx.Done():
			Log.T.F("relay listener context done, closing websocket %s", ws.Remote())
			cancel()
			Log.W.Ln("removing ws")
			rl.RemoveWS(ws)
			return
		case <-ws.Ctx.Done():
			Log.T.Ln("websocket %s context done", ws.Remote())
			Log.W.Ln("removing ws")
			rl.RemoveWS(ws)
			return
		}
	}
}

func (rl *T) wsReadMessages(ws *ws.Serv, cancel C.F) {
	Chk.E(ws.Conn.SetReadDeadline(time.Now().Add(PongWait)))
	ws.Conn.SetPongHandler(func(S) (err E) {
		if err = ws.Conn.SetReadDeadline(time.Now().Add(PongWait)); Chk.E(err) {
		}
		return
	})
	for {
		select {
		case <-rl.Ctx.Done():
			Log.T.Ln("relay listener context done, closing websocket %s", ws.Remote())
			cancel()
			return
		case <-ws.Ctx.Done():
			Log.T.Ln("websocket %s context done", ws.Remote())
			return
		default:
		}
		var err E
		var typ N
		var msg B
		if typ, msg, err = ws.Conn.ReadMessage(); err != nil {
			if W.IsUnexpectedCloseError(err,
				W.CloseNormalClosure, W.CloseGoingAway,
				W.CloseNoStatusReceived, W.CloseAbnormalClosure,
			) {
				Log.E.F("unexpected close error from %s: %v", ws.Remote(), err)
			}
			rl.Tracker.Do(func() { rl.Tracker.RemoveWS(ws) })
			return
		}
		if typ == W.PingMessage {
			Chk.E(ws.Pong())
			continue
		}
		Log.T.F("received message from %s: \n%s", ws.Remote(), msg)
		var sentinel S
		var rem B
		if sentinel, rem, err = envelopes.Identify(msg); Chk.E(err) {
			continue
		}
		Log.T.F("received %s envelope from %s\n%s", sentinel, ws.Remote(), rem)
		switch sentinel {
		case authenvelope.L:
			env := authenvelope.NewResponse()
			if rem, err = env.UnmarshalJSON(rem); Chk.E(err) {
				return
			}
			Log.I.S(env)
		case closeenvelope.L:
			env := closeenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); Chk.E(err) {
				return
			}
			rl.Tracker.Do(func() { rl.RemoveSub(ws, env.ID) })
		case countenvelope.L:
			env := countenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); Chk.E(err) {
				return
			}
			Log.I.S(env)
		case eventenvelope.L:
			env := eventenvelope.NewSubmission()
			if rem, err = env.UnmarshalJSON(rem); Chk.E(err) {
				if err = okenvelope.NewFrom(env.T.EventID(), false,
					normalize.Error.Message(err.Error())).Write(ws); Chk.E(err) {
					continue
				}
				continue
			}
			rl.handleEvent(ws, env)
		case reqenvelope.L:
			env := reqenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); Chk.E(err) {
				return
			}
			rl.Tracker.Do(func() { rl.AddSub(ws, env.Subscription, env.Filters) })
			rl.handleReq(ws, env.Subscription, env.Filters)
		}
	}
}
