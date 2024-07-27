package subscriptionid

import (
	"bytes"
	"os"

	"git.replicatr.dev/pkg/util/lol"
)

type (
	B = []byte
	S = string
	E = error
)

var (
	log, chk, errorf = lol.New(os.Stderr)
	equals           = bytes.Equal
)
