//go:build btcec

package p256k1

// BTCECSigner is always available but enabling it disables the p256k1

type Signer = BTCECSigner
