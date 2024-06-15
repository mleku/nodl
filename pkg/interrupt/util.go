package interrupt

import (
	"os"

	"github.com/mleku/nodl/pkg/lol"
	"github.com/mleku/nodl/pkg/util"
)

var (
	log, chk, errorf = lol.New(os.Stderr)
	B                = util.B
)
