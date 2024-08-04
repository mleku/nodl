package main

import (
	"os"

	"git.replicatr.dev/pkg/replicatr"
	"git.replicatr.dev/pkg/util/context"
)

func main() {
	c, cancel := context.Cancel(context.Bg())
	replicatr.Main(os.Args, c, cancel)
}
