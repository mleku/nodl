package relay

import (
	"encoding/json"
	"net/http"

	. "nostr.mleku.dev"

	"nostr.mleku.dev/protocol/relayinfo"
	"util.mleku.dev/number"
)

var nips = number.List{
	relayinfo.BasicProtocol.Number,            // NIP1 events, envelopes and filters
	relayinfo.FollowList.Number,               // NIP2 contact list and pet names
	relayinfo.EncryptedDirectMessage.Number,   // NIP4 encrypted DM
	relayinfo.MappingNostrKeysToDNS.Number,    // NIP5 DNS
	relayinfo.EventDeletion.Number,            // NIP9 event delete
	relayinfo.RelayInformationDocument.Number, // NIP11 relay information document
	relayinfo.GenericTagQueries.Number,        // NIP12 generic tag queries
	relayinfo.NostrMarketplace.Number,         // NIP15 marketplace
	relayinfo.EventTreatment.Number,           // NIP16
	relayinfo.Reposts.Number,                  // NIP18 reposts
	relayinfo.Bech32EncodedEntities.Number,    // NIP19 bech32 encodings
	relayinfo.CommandResults.Number,           // NIP20
	// relayinfo.CreatedAtLimits.Number,                // NIP22 (probably never gonna happen)
	relayinfo.LongFormContent.Number,                // NIP23 long form
	relayinfo.PublicChat.Number,                     // NIP28 public chat
	relayinfo.ParameterizedReplaceableEvents.Number, // NIP33
	relayinfo.ExpirationTimestamp.Number,            // NIP40
	relayinfo.VersionedEncryption.Number,
	relayinfo.UserStatuses.Number,    // NIP38 user statuses
	relayinfo.Authentication.Number,  // NIP42 auth
	relayinfo.CountingResults.Number, // NIP45 count requests
}

// GetInfo returns a default relay info based on configurations
func GetInfo() *relayinfo.T {
	return &relayinfo.T{
		Name:        "replicatr",
		Description: "nostr relay",
		PubKey:      "",
		Contact:     "me@mleku.dev",
		Nips:        nips,
		Software:    "replicatr",
		Version:     "v0.0.1",
		Limitation: relayinfo.Limits{
			MaxMessageLength: MaxMessageSize,
			Oldest:           1640305962,
			AuthRequired:     false,
			PaymentRequired:  false,
			RestrictedWrites: false,
			MaxSubscriptions: 50,
		},
		// Retention:      ,
		// RelayCountries: tag.T{},
		// LanguageTags:   tag.T{},
		// Tags:           tag.T{},
		PostingPolicy: "",
		// PaymentsURL:    "https://gfy.mleku.dev",
		// Fees: relayinfo.Fees{
		// 	Admission: []relayinfo.Admission{
		// 		{Amount: 100000000, Unit: "satoshi"},
		// 	},
		// },
		Icon: "https://raw.githubusercontent.com/Hubmakerlabs/replicatr/main/doc/logo.jpg",
	}
}

func (rl *T) HandleRelayInfo(w http.ResponseWriter, r *http.Request) {
	var err E
	Log.T.Ln("HandleRelayInfo")
	// info := relayinfo.NewInfo(&relayinfo.T{})
	info := GetInfo()
	w.Header().Set("Content-Type", "application/nostr+json")
	var b []byte
	if b, err = json.Marshal(info); Chk.E(err) {
		return
	}
	Log.T.F("%s", b)
	if _, err = w.Write(b); Chk.E(err) {
		return
	}
}
