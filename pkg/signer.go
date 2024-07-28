package pkg

type Signer interface {
	// InitSec initialises the secret (signing) key from the raw bytes, and also
	// derives the public key because it can.
	InitSec(sec B) (err E)
	// InitPub initializes the public (verification) key from raw bytes.
	InitPub(pub B) (err E)
	// Pub returns the public key bytes.
	Pub() B
	// Sign creates a signature using the stored secret key.
	Sign(msg B) (sig B, err E)
	// Verify checks a message hash and signature match the stored public key.
	Verify(msg, sig B) (valid bool, err E)
	// Zero wipes the secret key to prevent memory leaks.
	Zero()
	// ECDH returns a shared secret derived using Elliptic Curve Diffie Hellman on the Signer secret and provided
	// pubkey.
	ECDH(pub B) (secret B, err E)
}
