package main

import (
	"os"

	"git.replicatr.dev/pkg/replicatr"
	"git.replicatr.dev/pkg/util/context"
	"git.replicatr.dev/pkg/util/lol"
)

func main() {
	lol.SetLogLevel("trace")
	c, cancel := context.Cancel(context.Bg())
	replicatr.Main(os.Args, c, cancel)
}
