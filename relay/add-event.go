package relay

import (
	"context"
	"regexp"

	"git.replicatr.dev/eventstore"
	"git.replicatr.dev/relay/types"
	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"util.mleku.dev/normalize"
)

var nip20prefixmatcher = regexp.MustCompile(`^\w+: `)

// AddEvent has a business rule to add an event to the relayer
func AddEvent(ctx context.Context, relay types.Relayer, evt *event.T) (accepted bool,
	message B) {
	if evt == nil {
		return false, nil
	}

	store := relay.Storage()
	wrapper := &eventstore.RelayWrapper{I: store}
	advancedSaver, _ := store.(types.AdvancedSaver)

	if !relay.AcceptEvent(ctx, evt) {
		return false, normalize.Blocked.F("event blocked by relay")
	}
	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(ctx, evt)
		}

		if saveErr := wrapper.Publish(ctx, evt); saveErr != nil {
			switch saveErr {
			case eventstore.ErrDupEvent:
				return true, B(saveErr.Error())
			default:
				errmsg := saveErr.Error()
				if nip20prefixmatcher.MatchString(errmsg) {
					return false, B(errmsg)
				} else {
					return false, normalize.Error.F("failed to save(%s)", errmsg)
					// B(fmt.Sprintf("error: failed to save (%s)", errmsg))
				}
			}
		}

		if advancedSaver != nil {
			advancedSaver.AfterSave(evt)
		}
	}

	notifyListeners(evt)

	return true, nil
}
