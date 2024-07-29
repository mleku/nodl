//go:build btcec

package p256k

// BTCECSigner is always available but enabling it disables the use of
// github.com/bitcoin-core/secp256k1 CGO signature implementation.

type Signer = BTCECSigner

func GenSecBytes() (skb, pkb B, err error) {
	// just use the btcec key gen because the performance difference will be nearly zero... just make sure we don't make
	// odd pub keys (BIP-340 doesn't allow them)
	for {
		var sk *btcec.SecretKey
		if sk, err = btcec.NewSecretKey(); chk.E(err) {
			return
		}
		pkb = sk.PubKey().SerializeCompressed()
		if pkb[0] == 2 {
			pkb = pkb[1:]
			skb = sk.Serialize()
			break
		}
	}
	return
}
