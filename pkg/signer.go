package pkg

type Signer interface {
	// InitSec initialises the secret (signing) key from the raw bytes, and also
	// derives the public key because it can.
	InitSec(sec B) (err error)
	// InitPub initializes the public (verification) key from raw bytes.
	InitPub(pub B) (err error)
	// Pub returns the public key bytes.
	Pub() B
	// Sign creates a signature using the stored secret key.
	Sign(msg B) (sig B, err error)
	// Verify checks a message hash and signature match the stored public key.
	Verify(msg, sig B) (valid bool, err error)
	// Zero wipes the secret key to prevent memory leaks.
	Zero()
}
