package relay

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/codec/tag"
	"git.replicatr.dev/pkg/codec/tags"
	"git.replicatr.dev/pkg/codec/timestamp"
	"git.replicatr.dev/pkg/crypto/encryption"
	"git.replicatr.dev/pkg/crypto/p256k"
)

// DecryptDM decrypts a DM, kind 4, 1059 or 1060
func DecryptDM(ev *event.T, meSec, youPub B) (decryptedStr B, err E) {
	switch ev.Kind {
	case kind.EncryptedDirectMessage:
		var secret, decrypted B
		if secret, err = encryption.ComputeSharedSecret(S(meSec), S(youPub)); chk.E(err) {
			return
		}
		if decrypted, err = encryption.DecryptNip4(ev.ContentString(), secret); chk.E(err) {
			return
		}
		decryptedStr = decrypted
	case kind.GiftWrap:
	case kind.GiftWrapWithKind4:
	}
	return
}

// EncryptDM encrypts a DM, kind 4, 1059 or 1060
func EncryptDM(ev *event.T, meSec, youPub B) (evo *event.T, err E) {
	var secret []byte
	switch ev.Kind {
	case kind.EncryptedDirectMessage:
		if secret, err = encryption.ComputeSharedSecret(S(meSec), S(youPub)); chk.E(err) {
			return
		}
		if ev.Content, err = encryption.EncryptNip4(ev.ContentString(), secret); chk.E(err) {
			return
		}
		sec := &p256k.Signer{}
		if err = sec.InitSec(meSec); chk.E(err) {
			return
		}
		if err = ev.Sign(sec); chk.E(err) {
			return
		}
	case kind.GiftWrap:
	case kind.GiftWrapWithKind4:
	}
	evo = ev
	return
}

// MakeReply creates an appropriate reply event from a provided event that is
// being replied to (not quoting, just the right tags, timestamps and kind).
func MakeReply(ev *event.T, content S) (evo *event.T) {
	created := ev.CreatedAt.I64() + 2
	now := timestamp.Now().I64()
	if created < now {
		created = now
	}
	evo = &event.T{
		CreatedAt: timestamp.FromUnix(created),
		Kind:      ev.Kind,
		Tags:      tags.New(tag.New("p", ev.PubKeyString()), tag.New("e", ev.IDString())),
		Content:   event.B(content),
	}
	return
}

// Chat implements the control interface, intercepting kind 4 encrypted direct
// messages and processing them if they are for the relay's pubkey
func (rl *R) Chat(c Ctx, ev *event.T) (err E) {
	ws := GetConnection(c)
	if ws == nil {
		return
	}
	// log.T.Ln("running chat checker")
	if ev.Kind != kind.EncryptedDirectMessage {
		// log.T.Ln("not chat event", ev.Kind, kind.GetString(ev.Kind))
		return
	}
	if !ev.Tags.ContainsAny(B("p"),
		B(rl.RelayPubHex)) && ev.PubKeyString() != rl.RelayPubHex {
		// log.T.Ln("direct message not for relay chat", ev.PubKey, rl.RelayPubHex)
		return
	}
	meSec := rl.Config.SecKey
	youPub := ev.PubKey
	log.T.Ln(rl.RelayPubHex, "receiving message via DM", ev.String())
	var decryptedStr B
	decryptedStr, err = DecryptDM(ev, B(meSec), youPub)
	log.T.F("decrypted message: '%s'", decryptedStr)
	decryptedStr = bytes.TrimSpace(decryptedStr)
	var reply *event.T
	if !ws.HasAuth() {
		if bytes.HasPrefix(decryptedStr, B("AUTH_")) {
			var authed bool
			authStr := strings.Split(S(decryptedStr), "_")
			log.I.Ln(authStr, ws.Challenge())
			if len(authStr) == 3 {
				var ts int64
				if ts, err = strconv.ParseInt(authStr[1], 10, 64); chk.E(err) {
					return
				}
				now := timestamp.Now().Time().Unix()
				log.I.Ln()
				if ts < now+15 || ts > now-15 {
					if equals(B(authStr[2]), ws.Challenge()) {
						authed = true
						ws.SetAuthPubKey(ev.PubKey)
					}
				}
			}
			if !authed {
				reply = MakeReply(ev, fmt.Sprintf("access denied"))
				if reply, err = EncryptDM(reply, B(meSec), youPub); chk.E(err) {
					return
				}
				log.T.Ln("reply", reply.String())
				rl.BroadcastEvent(reply)
				ws.GenerateChallenge()
				return
			} else {
				// now process cached
				log.T.Ln("pending message:", ws.Pending.Load())
				cmd := ws.Pending.Load().(string)
				// erase
				ws.Pending.Store("")
				chk.E(rl.command(ev, cmd))
				return
			}
		} else {
			// store the input in the websocket state to process after
			// successful auth
			ws.Pending.Store(decryptedStr)
			content := fmt.Sprintf(`
received command

%s

to authorise executing this command, please reply within 15 seconds with the following text:

AUTH_%d_%v

after this you will not have to do this again unless there is a long idle, disconnect or you refresh your session

note that if you have NIP-42 enabled in the client and you are already authorised this notice will not appear
`, decryptedStr, timestamp.Now(), ws.Challenge())
			log.I.F("sending message to user\n%s", content)
			reply = MakeReply(ev, content)
			if reply, err = EncryptDM(reply, B(meSec), youPub); chk.E(err) {
				return
			}
			log.T.Ln("reply", reply.String())
			rl.BroadcastEvent(reply)
			return
		}
	} else {
		if err = rl.command(ev, S(decryptedStr)); chk.E(err) {
			return
		}
	}
	return
}

type Command struct {
	Name string
	Help string
	Func func(rl *R, prefix S, ev *event.T, cmd *Command, args ...S) (reply *event.T, err E)
}

func (rl *R) command(ev *event.T, cmd S) (err E) {
	log.T.Ln("running relay method")
	args := strings.Split(cmd, " ")
	if len(args) < 1 {
		err = log.E.Err("no command received")
		return
	}
	var reply *event.T
	for i := range Commands {
		if Commands[i].Name == args[0] {
			if reply, err = Commands[i].Func(rl, "", ev, Commands[i],
				args...); chk.E(err) {
				return
			}
			break
		}
	}
	if reply == nil {
		for i := range Commands {
			if Commands[i].Name == "help" {
				reply, err = help(rl, fmt.Sprintf("unknown command: '%s'", cmd),
					ev, Commands[i], args...)
				if chk.E(err) {
					return
				}
				break
			}
		}
	}
	if reply, err = EncryptDM(reply, B(rl.Config.SecKey), ev.PubKey); chk.E(err) {
		return
	}
	rl.BroadcastEvent(reply)
	return
}
