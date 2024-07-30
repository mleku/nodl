//go:build btcec

package p256k

import (
	"git.replicatr.dev/pkg/crypto/p256k/btcec"
)

// BTCECSigner is always available but enabling it disables the use of
// github.com/bitcoin-core/secp256k1 CGO signature implementation.

type Signer = btcec.Signer
