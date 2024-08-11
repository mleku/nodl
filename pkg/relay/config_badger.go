//go:build badger

package relay

import (
	"encoding/json"
	"errors"
	"os"
)

const DefaultListener = "0.0.0.0:3334"

func GetDefaultConfig() *Config {
	return &Config{
		BaseConfig: BaseConfig{
			Listen:        []string{DefaultListener},
			Profile:       ".replicatr",
			Name:          "replicatr relay",
			Icon:          "https://i.nostr.build/n8vM.png",
			AuthRequired:  false,
			Public:        true,
			DBLowWater:    86,
			DBHighWater:   92,
			GCFrequency:   300,
			MaxProcs:      4,
			// LogLevel:      "trace",
			GCRatio:       100,
			MemLimit:      500000000,
		},
	}
}

type Config struct {
	BaseConfig
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
