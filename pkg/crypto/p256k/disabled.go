//go:build btcec

package p256k

// BTCECSigner is always available but enabling it disables the use of
// github.com/bitcoin-core/secp256k1 CGO signature implementation.

type Signer = BTCECSigner
