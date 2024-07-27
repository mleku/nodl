package p256k_test

import (
	"bytes"
	"os"

	"git.replicatr.dev/pkg/util/lol"
)

type (
	B = []byte
	S = string
)

var (
	log, chk, errorf = lol.New(os.Stderr)
	equals           = bytes.Equal
)
