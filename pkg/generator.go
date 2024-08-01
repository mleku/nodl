package pkg

// Generator is an interface for nost BIP-340 key mining.
type Generator interface {
	// Generate gathers entropy and derives pubkey bytes for matching, this returns the 33 byte compressed form for
	// checking the oddness of the Y coordinate.
	Generate() (pubBytes B, err E)
	// Negate flips the public key Y coordinate between odd and even.
	Negate()
	// KeyPairBytes returns the raw bytes of the secret and public key, this returns the 32 byte X-only pubkey.
	KeyPairBytes() (secBytes, cmprPubBytes B)
}
