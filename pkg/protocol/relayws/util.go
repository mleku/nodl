package relayws

import (
	"bytes"
	"os"

	"github.com/mleku/nodl/pkg/util/lol"
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