//go:build !badger

package relay

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"git.replicatr.dev/pkg/codec/timestamp"
)

func GetDefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			Listen:       []string{"0.0.0.0:3334"},
			Profile:      ".replicatr",
			Name:         "replicatr icp relay",
			Icon:         "https://i.nostr.build/n8vM.png",
			AuthRequired: false,
			Public:       true,
			DBLowWater:   86,
			DBHighWater:  92,
			GCFrequency:  300,
			MaxProcs:     4,
			LogLevel:     "info",
			GCRatio:      100,
			MemLimit:     500000000,
		},
		PollFrequency: 5 * time.Second,
		PollOverlap:   4,
	}
}

type PubKey struct{}
type AddRelay struct {
	PubKey string `arg:"--addpubkey" help:"public key of client to add"`
	Admin  bool   `arg:"--admin"  help:"set client as admin"`
}
type RemoveRelay struct {
	PubKey string `arg:"--removepubkey" help:"public key of client to remove"`
}
type GetPermission struct {
}

type Config struct {
	BaseConfig
	EventStore       string         `arg:"-e,--eventstore" json:"eventstore" help:"select event store backend [ic,iconly]"`
	AddRelayCmd      *AddRelay      `arg:"subcommand:addrelay" json:"-" help:"add a relay to the cluster"`
	RemoveRelayCmd   *RemoveRelay   `arg:"subcommand:removerelay" json:"-" help:"remove a relay from the cluster"`
	PubKeyCmd        *PubKey        `arg:"subcommand:pubkey" json:"-" help:"print relay canister public key"`
	GetPermissionCmd *GetPermission `arg:"subcommand:getpermission" json:"-" help:"get permission of a relay"`
	CanisterAddr     string         `arg:"-C,--canisteraddr" json:"canister_addr" help:"IC canister address to use (for local, use http://127.0.0.1:<port number>)"`
	CanisterId       string         `arg:"-I,--canisterid" json:"canister_id" help:"IC canister ID to use"`
	// PollFrequency is how often the L2 is queried for recent events
	PollFrequency time.Duration `arg:"--pollfrequency" help:"if a level 2 event store is enabled how often it polls"`
	// PollOverlap is the multiple of the PollFrequency within which polling the L2
	// is done to ensure any slow synchrony on the L2 is covered (2-4 usually)
	PollOverlap timestamp.T `arg:"--polloverlap" help:"if a level 2 event store is enabled, multiple of poll freq overlap to account for latency"`
}

func (c *Config) Save(filename string) (err error) {
	if c == nil {
		err = errors.New("cannot save nil relay config")
		log.E.Ln(err)
		return
	}
	var b []byte
	if b, err = json.MarshalIndent(c, "", "    "); chk.E(err) {
		return
	}
	if err = os.WriteFile(filename, b, 0600); chk.E(err) {
		return
	}
	return
}

func (c *Config) Load(filename string) (err error) {
	if c == nil {
		err = errors.New("cannot load into nil config")
		chk.E(err)
		return
	}
	var b []byte
	if b, err = os.ReadFile(filename); chk.E(err) {
		return
	}
	// log.D.F("configuration\n%s", string(b))
	if err = json.Unmarshal(b, c); chk.E(err) {
		return
	}
	return
}
