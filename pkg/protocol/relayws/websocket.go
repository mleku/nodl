package relayws

import (
	"crypto/rand"
	"net/http"
	"sync"
	"time"

	"ec.mleku.dev/v2/bech32"
	"git.replicatr.dev/pkg/codec/bech32encoding"
	"git.replicatr.dev/pkg/codec/envelopes/enveloper"
	"git.replicatr.dev/pkg/util/atomic"
	"git.replicatr.dev/pkg/util/qu"
	w "github.com/fasthttp/websocket"
)

type MessageType int

// The message types are defined in RFC 6455, section 11.8.
//
// Repeating here for shorter names.
const (
	// TextMessage denotes a text data message. The text message payload is interpreted as UTF-8 encoded text data.
	TextMessage MessageType = w.TextMessage

	// BinaryMessage denotes a binary data message.
	BinaryMessage MessageType = w.BinaryMessage

	// CloseMessage denotes a close control message. The optional message payload contains a numeric code and text. Use
	// the FormatCloseMessage function to format a close message payload.
	CloseMessage MessageType = w.CloseMessage

	// PingMessage denotes a ping control message. The optional message payload is UTF-8 encoded text.
	PingMessage MessageType = w.PingMessage

	// PongMessage denotes a pong control message. The optional message payload is UTF-8 encoded text.
	PongMessage MessageType = w.PongMessage
)

// WS is a wrapper around a gorilla/websocket with mutex locking and
// NIP-42 IsAuthed support
type WS struct {
	Conn         *w.Conn
	remote       atomic.String
	mutex        sync.Mutex
	Request      *http.Request // original request
	challenge    atomic.String // nip42
	Pending      atomic.Value  // for DM CLI authentication
	authPubKey   atomic.Value
	Authed       chan struct{}
	OffenseCount atomic.Uint32 // when client does dumb stuff, increment this
}

func New(conn *w.Conn, req *http.Request, authed qu.C) (ws *WS) {
	// authPubKey must be initialized with a zero length slice so it can be detected when it hasn't been loaded.
	var authPubKey atomic.Value
	authPubKey.Store(B{})
	return &WS{Conn: conn, Request: req, Authed: authed, authPubKey: authPubKey}
}

func (ws *WS) Pong() (err E) {
	return ws.write(w.PongMessage, nil)
}
func (ws *WS) Ping() (err E) {
	return ws.write(w.PingMessage, nil)
}

// write writes a message with a given websocket type specifier
func (ws *WS) write(t MessageType, b B) (err E) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if len(b) != 0 {
		log.T.F("sending message to %s %0x\n%s", ws.Remote(), ws.AuthPub(), string(b))
	}
	chk.E(ws.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5)))
	return ws.Conn.WriteMessage(int(t), b)
}

// WriteTextMessage writes a text (binary?) message
func (ws *WS) WriteTextMessage(b B) (err E) {
	return ws.write(w.TextMessage, b)
}

// WriteEnvelope writes a message with a given websocket type specifier
func (ws *WS) WriteEnvelope(env enveloper.I) (err error) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	var b B
	if b, err = env.MarshalJSON(b); chk.E(err) {
		return
	}
	chk.E(ws.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5)))
	return ws.Conn.WriteMessage(int(TextMessage), b)
}

const ChallengeLength = 16
const ChallengeHRP = "nch"

// GenerateChallenge gathers new entropy to generate a new challenge, stores it and returns it.
func (ws *WS) GenerateChallenge() (challenge S) {
	var err error
	// create a new challenge for this connection
	cb := make([]byte, ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		// i never know what to do for this case, panic? usually just ignore, it should never happen
		panic(err)
	}
	var b5 B
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded B
	if encoded, err = bech32.Encode(bech32.B(ChallengeHRP), b5); chk.E(err) {
		return
	}
	challenge = S(encoded)
	ws.challenge.Store(challenge)
	return
}

// Challenge returns the current challenge on a websocket.
func (ws *WS) Challenge() (challenge B) { return B(ws.challenge.Load()) }

// Remote returns the current real remote.
func (ws *WS) Remote() (remote S)     { return ws.remote.Load() }
func (ws *WS) SetRealRemote(remote S) { ws.remote.Store(remote) }

// SetAuthPubKey loads the outhPubKey atomic of the websocket. Note that []byte is a reference so the caller should not
// mutate it. Calls to access it are copied as above.
func (ws *WS) SetAuthPubKey(a B) { ws.authPubKey.Store(a) }

// AuthPub returns the current authed Pubkey.
func (ws *WS) AuthPub() (a B) {
	b := ws.authPubKey.Load().(B)
	// make a copy because bytes are references
	a = append(a, b...)
	return
}

func (ws *WS) HasAuth() bool {
	b := ws.authPubKey.Load().(B)
	return len(b) > 0
}
