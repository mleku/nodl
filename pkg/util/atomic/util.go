package atomic

import (
	"os"

	"git.replicatr.dev/pkg/util/lol"
)

type (
	B = []byte
	S = string
)

var (
	log, chk, errorf = lol.New(os.Stderr)
)
