package relay

import (
	"net/http"
	"time"

	"git.replicatr.dev/pkg/codec/envelopes"
	"git.replicatr.dev/pkg/codec/envelopes/authenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/closeenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/countenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eventenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/protocol/relayws"
	C "git.replicatr.dev/pkg/util/context"
	W "github.com/fasthttp/websocket"
)

func (rl *T) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	log.T.F("HandleWebsocket inbound connection from %s", r.RemoteAddr)
	var err E
	var conn *W.Conn
	conn, err = rl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	c, cancel := C.Cancel(C.Bg())
	ws := relayws.New(c, conn, r, MaxMessageSize)
	log.T.Ln("upgraded websocket", ws.Remote())
	rl.Tracker.Do(func() { rl.AddWS(ws) })
	log.T.F("established websocket connection with %s", ws.Remote())
	go rl.wsReadMessages(ws, cancel)
	rl.wsWatcher(ws, cancel)
}

func (rl *T) wsWatcher(ws *relayws.WS, cancel C.F) {
	for {
		select {
		case <-rl.Ctx.Done():
			log.T.F("relay listener context done, closing websocket %s", ws.Remote())
			cancel()
			log.W.Ln("removing ws")
			rl.RemoveWS(ws)
			return
		case <-ws.Ctx.Done():
			log.T.Ln("websocket %s context done", ws.Remote())
			log.W.Ln("removing ws")
			rl.RemoveWS(ws)
			return
		}
	}
}

func (rl *T) wsReadMessages(ws *relayws.WS, cancel C.F) {
	chk.E(ws.Conn.SetReadDeadline(time.Now().Add(PongWait)))
	ws.Conn.SetPongHandler(func(S) (err E) {
		if err = ws.Conn.SetReadDeadline(time.Now().Add(PongWait)); chk.E(err) {
		}
		return
	})
	for {
		select {
		case <-rl.Ctx.Done():
			log.T.Ln("relay listener context done, closing websocket %s", ws.Remote())
			cancel()
			return
		case <-ws.Ctx.Done():
			log.T.Ln("websocket %s context done", ws.Remote())
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
				log.E.F("unexpected close error from %s: %v", ws.Remote(), err)
			}
			rl.Tracker.Do(func() { rl.Tracker.RemoveWS(ws) })
			return
		}
		if typ == W.PingMessage {
			chk.E(ws.Pong())
			continue
		}
		log.T.F("received message from %s: \n%s", ws.Remote(), msg)
		var sentinel S
		var rem B
		if sentinel, rem, err = envelopes.Identify(msg); chk.E(err) {
			continue
		}
		log.T.F("received %s envelope from %s\n%s", sentinel, ws.Remote(), rem)
		switch sentinel {
		case authenvelope.L:
			env := authenvelope.NewResponse()
			if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			log.I.S(env)
		// case closedenvelope.L:
		// 	env := closedenvelope.New()
		// 	if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
		// 		return
		// 	}
		// 	log.I.S(env)
		case closeenvelope.L:
			env := closeenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			rl.Tracker.Do(func() { rl.RemoveSub(ws, env.ID) })
		case countenvelope.L:
			env := countenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			log.I.S(env)
		// case eoseenvelope.L:
		// 	env := eoseenvelope.New()
		// 	if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
		// 		return
		// 	}
		// 	log.I.S(env)
		case eventenvelope.L:
			env := eventenvelope.NewSubmission()
			if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			rl.handleEvent(ws, env)
		// case noticeenvelope.L:
		// 	env := noticeenvelope.New()
		// 	if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
		// 		return
		// 	}
		// 	log.I.S(env)
		// case okenvelope.L:
		// 	env := okenvelope.New()
		// 	if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
		// 		return
		// 	}
		// 	log.I.S(env)
		case reqenvelope.L:
			env := reqenvelope.New()
			if rem, err = env.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			rl.Tracker.Do(func() { rl.AddSub(ws, env.Subscription, env.Filters) })
			rl.handleReq(ws, env.Filters, env.Subscription)
		}
	}
}
