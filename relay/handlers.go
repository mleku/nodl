package relay

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"git.replicatr.dev/relay/types"
	"git.replicatr.dev/relay/web"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/envelopes/authenvelope"
	"nostr.mleku.dev/codec/envelopes/countenvelope"
	"nostr.mleku.dev/codec/envelopes/eventenvelope"
	"nostr.mleku.dev/codec/envelopes/noticeenvelope"
	"nostr.mleku.dev/codec/envelopes/okenvelope"
	"nostr.mleku.dev/codec/event"
	"nostr.mleku.dev/codec/filters"
	"nostr.mleku.dev/codec/kind"
	"nostr.mleku.dev/protocol/auth"
	"nostr.mleku.dev/protocol/relayinfo"
	"util.mleku.dev/hex"
	"util.mleku.dev/normalize"

	"git.replicatr.dev/eventstore"
	"github.com/fasthttp/websocket"
	"github.com/nbd-wtf/go-nostr"
	"golang.org/x/time/rate"
	"nostr.mleku.dev/codec/filter"
	"nostr.mleku.dev/codec/tag"
	"util.mleku.dev/context"
)

func challenge(conn *websocket.Conn) *web.Socket {
	// NIP-42 challenge
	challengeBytes := make(B, 8)
	c := make(B, 16)
	rand.Read(challengeBytes)
	hex.EncBytes(c, challengeBytes)
	return &web.Socket{
		Conn:      conn,
		Challenge: c,
	}
}

func (s *Server) doEvent(c context.T, ws *web.Socket, req []json.RawMessage,
	store eventstore.I) S {
	advancedDeleter, _ := store.(types.AdvancedDeleter)
	latestIndex := len(req) - 1

	// it's a new event
	var evt event.T
	if _, err := evt.UnmarshalJSON(req[latestIndex]); err != nil {
		return "failed to decode event: " + err.Error()
	}

	// check id
	hash := sha256.Sum256(evt.Serialize())
	if !Equals(hash[:], evt.ID) {
		reason := "event id is computed incorrectly"
		okenvelope.NewFrom(evt.ID, false, normalize.Invalid.F(reason)).Write(ws)
		return ""
	}

	// check signature
	if ok, err := evt.CheckSignature(); err != nil {

		reason := "failed to verify signature"
		okenvelope.NewFrom(evt.ID, false, normalize.Error.F(reason)).Write(ws)
		return ""
	} else if !ok {
		reason := "signature is invalid"
		okenvelope.NewFrom(evt.ID, false, normalize.Invalid.F(reason))
		return ""
	}

	if evt.Kind.Equal(kind.Deletion) {
		// event deletion -- nip09
		for _, t := range evt.Tags.T {
			if len(t.Field) >= 2 && t.ToStringSlice()[0] == "e" {
				ctx, cancel := context.Timeout(c, time.Millisecond*200)
				defer cancel()

				// fetch event to be deleted
				res, err := s.relay.Storage().QueryEvents(ctx,
					&filter.T{IDs: tag.New(B(t.Field[1]))})
				if err != nil {
					reason := "failed to query for target event"
					okenvelope.NewFrom(evt.ID, false, normalize.Error.F(reason)).Write(ws)
					return ""
				}
				if len(res) < 1 {
					continue
				}
				target := res[0]
				// check if this can be deleted
				if !Equals(target.PubKey, evt.PubKey) {
					okenvelope.NewFrom(evt.ID, false,
						normalize.Blocked.F("insufficient permissions")).Write(ws)
					return ""
				}

				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(ctx, t.ToStringSlice()[1],
						hex.Enc(evt.PubKey))
				}

				if err = store.DeleteEvent(ctx, target.EventID()); err != nil {
					okenvelope.NewFrom(evt.ID, false,
						normalize.Error.F("error: %s", err.Error())).Write(ws)
					return ""
				}

				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.ToStringSlice()[1],
						hex.Enc(evt.PubKey))
				}
			}
		}

		notifyListeners(&evt)
		okenvelope.NewFrom(evt.ID, true).Write(ws)
		return ""
	}

	ok, reason := AddEvent(c, s.relay, &evt)
	okenvelope.NewFrom(evt.ID, ok, reason).Write(ws)
	return ""
}

func (s *Server) doCount(ctx context.T, ws *web.Socket, request []json.RawMessage,
	store eventstore.I) S {
	counter, ok := store.(types.EventCounter)
	if !ok {
		return "restricted: this relay does not support NIP-45"
	}

	var id string
	json.Unmarshal(request[1], &id)
	if id == "" {
		return "COUNT has no <id>"
	}

	var total N
	filters := filters.Make(len(request) - 2)
	for i, filterReq := range request[2:] {
		if err := json.Unmarshal(filterReq, filters.F[i]); err != nil {
			return "failed to decode filter"
		}

		filter := filters.F[i]

		// prevent kind-4 events from being returned to unauthed users,
		//   only when authentication is a thing
		if _, ok := s.relay.(types.Authenticator); ok {
			if filter.Kinds.Contains(kind.EncryptedDirectMessage) {
				senders := filter.Authors
				receivers := filter.Tags.GetFirst(tag.New("p"))
				switch {
				case len(ws.Authed) == 0:
					// not authenticated
					return "restricted: this relay does not serve kind-4 to unauthenticated users, does your client implement NIP-42?"
				case senders.Len() == 1 && receivers.Len() < 2 && Equals(senders.Field[0], ws.Authed):
					// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
				case receivers.Len() == 1 && senders.Len() < 2 && Equals(receivers.Field[0], ws.Authed):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
				default:
					// restricted filter: do not return any events,
					//   even if other elements in filters array were not restricted).
					//   client should know better.
					return "restricted: authenticated user does not have authorization for requested filters."
				}
			}
		}

		count, err := counter.CountEvents(ctx, filter)
		if err != nil {
			Log.E.F("store: %v", err)
			continue
		}
		total += count
	}
	countenvelope.NewResponseFrom(id, total).Write(ws)
	ws.WriteJSON([]interface{}{"COUNT", id, map[S]N{"count": total}})
	return ""
}

func (s *Server) doReq(ctx context.T, ws *web.Socket, request []json.RawMessage,
	store eventstore.I) string {
	var id string
	json.Unmarshal(request[1], &id)
	if id == "" {
		return "REQ has no <id>"
	}
	filters := filters.Make(len(request) - 2)
	for i, filterReq := range request[2:] {
		if err := json.Unmarshal(
			filterReq,
			filters.F[i],
		); err != nil {
			return "failed to decode filter"
		}
	}

	if accepter, ok := s.relay.(types.RequestAcceptor); ok {
		if !accepter.AcceptReq(ctx, id, filters, ws.Authed) {
			return "REQ filters are not accepted"
		}
	}

	for _, filter := range filters.F {

		// prevent kind-4 events from being returned to unauthed users,
		//   only when authentication is a thing
		if _, ok := s.relay.(types.Authenticator); ok {
			if filter.Kinds.Contains(kind.EncryptedDirectMessage) {
				senders := filter.Authors
				receivers := filter.Tags.GetFirst(tag.New("p"))
				switch {
				case len(ws.Authed) == 0:
					// not authenticated
					return "restricted: this relay does not serve kind-4 to unauthenticated users, does your client implement NIP-42?"
				case senders.Len() == 1 && receivers.Len() < 2 && Equals(senders.Field[0], ws.Authed):
				// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
				case receivers.Len() == 1 && senders.Len() < 2 && Equals(receivers.Field[0], ws.Authed):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
				default:
					// restricted filter: do not return any events,
					//   even if other elements in filters array were not restricted).
					//   client should know better.
					return "restricted: authenticated user does not have authorization for requested filters."
				}
			}
		}

		events, err := store.QueryEvents(ctx, filter)
		if err != nil {
			Log.E.F("store: %v", err)
			continue
		}

		// ensures the client won't be bombarded with events in case Storage doesn't do limits right
		if filter.Limit == 0 {
			filter.Limit = 9999999999
		}
		i := 0
		for e := range events {
			if s.Options.skipEventFunc != nil && s.Options.skipEventFunc(events[e]) {
				continue
			}
			eventenvelope.NewResultWith(id, events[e])
			i++
			if i > filter.Limit {
				break
			}
		}

		// exhaust the channel (in case we broke out of it early) so it is closed by the storage
		for range events {
		}
	}

	ws.WriteJSON(nostr.EOSEEnvelope(id))
	setListener(id, ws, filters)
	return ""
}

func (s *Server) doClose(ctx context.T, ws *web.Socket, request []json.RawMessage,
	store eventstore.I) string {
	var id string
	json.Unmarshal(request[1], &id)
	if id == "" {
		return "CLOSE has no <id>"
	}

	removeListenerId(ws, id)
	return ""
}

func (s *Server) doAuth(ctx context.T, ws *web.Socket, request []json.RawMessage,
	store eventstore.I) (note S) {
	if auther, ok := s.relay.(types.Authenticator); ok {
		var err E
		evt, b := &event.T{}, B{}
		if b, err = evt.MarshalJSON(b); Chk.E(err) {
			return "failed to decode auth event: " + err.Error()
		}

		if ok, err = auth.Validate(evt, ws.Challenge, auther.ServiceURL()); !ok || Chk.E(err) {
			note = fmt.Sprintf("failed to auth: %s", err)
			okenvelope.NewFrom(evt.ID, false,
				normalize.Error.F(note)).Write(ws)
		} else {
			ctx = context.Value(ctx, AuthContextKey, evt.PubKey)
			okenvelope.NewFrom(evt.ID, true).Write(ws)
		}
	}
	return
}

func (s *Server) handleMessage(c context.T, ws *web.Socket, msg B, store eventstore.I) {
	var notice S
	defer func() {
		if notice != "" {
			noticeenvelope.NewFrom(notice).Write(ws)
		}
	}()

	var request []json.RawMessage
	if err := json.Unmarshal(msg, &request); err != nil {
		// stop silently
		return
	}

	if len(request) < 2 {
		notice = "request has less than 2 parameters"
		return
	}

	var typ string
	json.Unmarshal(request[0], &typ)

	switch typ {
	case "EVENT":
		notice = s.doEvent(c, ws, request, store)
	case "COUNT":
		notice = s.doCount(c, ws, request, store)
	case "REQ":
		notice = s.doReq(c, ws, request, store)
	case "CLOSE":
		notice = s.doClose(c, ws, request, store)
	case "AUTH":
		notice = s.doAuth(c, ws, request, store)
	default:
		if cwh, ok := s.relay.(types.CustomWebSocketHandler); ok {
			cwh.HandleUnknownType(ws, typ, request)
		} else {
			notice = "unknown message type " + typ
		}
	}
}

func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[conn] = struct{}{}
	ticker := time.NewTicker(s.PingPeriod)

	ip := conn.RemoteAddr().String()
	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP // possible to be multiple comma separated
	} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	Log.I.F("connected from %s", ip)

	ws := challenge(conn)

	if s.Options.perConnectionLimiter != nil {
		ws.Limiter = rate.NewLimiter(
			s.Options.perConnectionLimiter.Limit(),
			s.Options.perConnectionLimiter.Burst(),
		)
	}

	ctx, cancel := context.Cancel(context.Bg())

	store := s.relay.Storage()

	// reader
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			s.clientsMu.Lock()
			if _, ok := s.clients[conn]; ok {
				conn.Close()
				delete(s.clients, conn)
				removeListener(ws)
			}
			s.clientsMu.Unlock()
			Log.I.F("disconnected from %s", ip)
		}()

		conn.SetReadLimit(int64(s.MaxMessageSize))
		conn.SetReadDeadline(time.Now().Add(s.PongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(s.PongWait))
			return nil
		})

		// NIP-42 auth challenge
		if _, ok := s.relay.(types.Authenticator); ok {
			authenvelope.NewChallengeWith(ws.Challenge).Write(ws)
		}

		for {
			typ, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,        // 1001
					websocket.CloseNoStatusReceived, // 1005
					websocket.CloseAbnormalClosure,  // 1006
				) {
					Log.W.F("unexpected close error from %s: %v",
						r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}

			if ws.Limiter != nil {
				// NOTE: Wait will throttle the requests.
				// To reject requests exceeding the limit, use if !ws.limiter.Allow()
				if err := ws.Limiter.Wait(context.TODO()); err != nil {
					Log.W.F("unexpected limiter error %v", err)
					continue
				}
			}

			if typ == websocket.PingMessage {
				ws.WriteMessage(websocket.PongMessage, nil)
				continue
			}

			go s.handleMessage(ctx, ws, message, store)
		}
	}()

	// writer
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			conn.Close()
		}()

		for {
			select {
			case <-ticker.C:
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(s.WriteWait))
				if err != nil {
					Log.E.F("error writing ping: %v; closing websocket", err)
					return
				}
				Log.I.F("pinging for %s", ip)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Server) HandleNIP11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var info *relayinfo.T
	if ifmer, ok := s.relay.(types.Informer); ok {
		info = ifmer.GetNIP11InformationDocument()
	} else {
		supportedNIPs := []int{9, 11, 12, 15, 16, 20, 33}
		if _, ok := s.relay.(types.Authenticator); ok {
			supportedNIPs = append(supportedNIPs, 42)
		}
		if storage, ok := s.relay.(eventstore.I); ok && storage != nil {
			if _, ok = storage.(types.EventCounter); ok {
				supportedNIPs = append(supportedNIPs, 45)
			}
		}

		info = &relayinfo.T{
			Name:        s.relay.Name(),
			Description: "relay powered by the relayer framework",
			PubKey:      "~",
			Contact:     "~",
			Nips:        supportedNIPs,
			Software:    "https://github.com/fiatjaf/relayer",
			Version:     "~",
		}
	}

	json.NewEncoder(w).Encode(info)
}
