package enveloper

import (
	"git.replicatr.dev/pkg/codec"
)

type I interface {
	Label() string
	Write(ws Writer) (err E)
	codec.JSON
}

type Writer interface {
	WriteEnvelope(env I) (err error)
}
