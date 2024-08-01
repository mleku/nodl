package pkg

type Signer interface {
	// Generate creates a fresh new key pair from system entropy, and ensures it is even (so ECDH works).
	Generate() (err E)
	// InitSec initialises the secret (signing) key from the raw bytes, and also
	// derives the public key because it can.
	InitSec(sec B) (err E)
	// InitPub initializes the public (verification) key from raw bytes.
	InitPub(pub B) (err E)
	// Sec returns the secret key bytes.
	Sec() B
	// Pub returns the public key bytes (x-only schnorr pubkey).
	Pub() B
	// ECPub returns the public key bytes (33 byte ecdsa pubkey). The first byte is always 2 due to ECDH and X-only
	// keys.
	ECPub() B
	// Sign creates a signature using the stored secret key.
	Sign(msg B) (sig B, err E)
	// Verify checks a message hash and signature match the stored public key.
	Verify(msg, sig B) (valid bool, err E)
	// Zero wipes the secret key to prevent memory leaks.
	Zero()
	// ECDH returns a shared secret derived using Elliptic Curve Diffie Hellman on the Signer secret and provided
	// pubkey.
	ECDH(pub B) (secret B, err E)
	// Negate flips the the secret key to change between odd and even compressed public key.
	Negate()
}
