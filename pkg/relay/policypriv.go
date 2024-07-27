package relay

import (
	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/tag"
	// "git.replicatr.dev/pkg/codec/tag"
)

// FilterPrivileged interacts between filters and the privileges of the
// users according to the access control settings of the relay, checking whether
// the request is authorised, if not, requesting authorisation.
//
// If there is an ACL configured, it acts as a whitelist, no access without
// being on the ACL.
//
// If the message is a private message, only authenticated users may get these
// events who also match one of the parties in the conversation.
func (rl *R) FilterPrivileged(c Ctx, id SubID, f *filter.T) (reject bool, msg B) {

	ws := GetConnection(c)
	authRequired := rl.Info.Limitation.AuthRequired
	if !authRequired {
		return
	}
	var allow bool
	for _, v := range rl.Config.AllowIPs {
		if ws.Remote() == v {
			allow = true
			break
		}
	}
	if allow {
		return
	}
	// check if the request filter kinds are privileged
	if !f.Kinds.IsPrivileged() {
		return
	}
	if !rl.IsAuthed(c, "privileged") {
		return
	}
	rtags := f.Tags.GetAll(tag.New("#p"))
	receivers := tag.NewWithCap(rtags.Len())
	for i := range rtags.T {
		receivers.Field = append(receivers.Field, rtags.T[i].Value())
	}
	parties := tag.NewWithCap(receivers.Len() + f.Authors.Len())
	copy(parties.Field[:f.Authors.Len()], f.Authors.Field)
	copy(parties.Field[f.Authors.Len():], receivers.Field)
	log.D.Ln(ws.Remote(), "parties", parties, "querant", ws.AuthPub())
	switch {
	case !ws.HasAuth():
		// not authenticated
		return true,
			B("restricted: this relay does not serve privileged events to unauthenticated users, " +
				"does your client implement NIP-42?")
	case parties.Contains(ws.AuthPub()):
		// if the authed key is party to the messages, either as authors or
		// recipients then they are permitted to see the message.
		return
	default:
		// restricted filter: do not return any events, even if other elements
		// in filters array were not restricted). client should know better.
		return true, B("restricted: authenticated user does not match either party in privileged message type")
	}
}
