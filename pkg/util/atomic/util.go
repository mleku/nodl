package atomic

import (
	"os"

	"github.com/mleku/nodl/pkg/util/lol"
)

type (
	B = []byte
	S = string
)

var (
	log, chk, errorf = lol.New(os.Stderr)
)
