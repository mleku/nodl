package setup

import (
	"bytes"

	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/lol"
)

type (
	B   = []byte
	S   = string
	E   = error
	N   = int
	Ctx = context.T
)

var (
	log, chk, errorf = lol.Main.Log, lol.Main.Check, lol.Main.Errorf
	equals           = bytes.Equal
)
