package keys

import (
	k1 "github.com/mleku/btcec/secp256k1"
	"github.com/mleku/nodl/pkg/hex"
)

func SecHexToSecretKey(skh string) (sk *k1.SecretKey, err error) {
	// secret key hex must be 64 characters.
	if len(skh) != 64 {
		err = errorf.E("invalid secret key length, 64 required, got %d: %s",
			len(skh), skh)
		return
	}
	// decode secret key hex to bytes
	var skBytes []byte
	if skBytes, err = hex.Dec(skh); chk.D(err) {
		err = errorf.E("sign called with invalid secret key '%s': %w", skh,
			err)
		return
	}
	// parse bytes to get secret key (size checks have been done).
	sk = k1.SecKeyFromBytes(skBytes)
	return
}
