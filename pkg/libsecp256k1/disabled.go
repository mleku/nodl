//go:build btcec

package libsecp256k1

// BTCECSigner is always available but enabling it disables the libsecp256k1

type Signer = BTCECSigner
