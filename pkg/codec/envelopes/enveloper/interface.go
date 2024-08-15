package enveloper

import (
	"git.replicatr.dev/pkg/codec"
	"git.replicatr.dev/pkg/protocol/relayws"
)

type I interface {
	Label() string
	Write(ws *relayws.WS) (err E)
	codec.JSON
}
