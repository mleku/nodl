package main

import (
	"bytes"
	"fmt"
	"os"

	"ec.mleku.dev/v2/secp256k1"
	"git.replicatr.dev/pkg/codec/bech32encoding"
	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/util/hex"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("keyfix: check your nsec (in bech32 or hex format) makes an even key and if not, " +
			"give you the fixed version\n\n")
		fmt.Printf("Usage: keyfix <nsec>\n")
		os.Exit(0)
	}
	nsec := B(os.Args[1])
	sec := make(B, secp256k1.SecKeyBytesLen)
	var err E
	if bytes.HasPrefix(nsec, bech32encoding.NsecHRP) {
		var prf B
		var val any
		if prf, val, err = bech32encoding.Decode(nsec); err != nil {
			fmt.Printf("error decoding nsec: %v\n", err)
			os.Exit(1)
		}
		if !equals(prf, bech32encoding.NsecHRP) {
			fmt.Printf("key '%s' is not an nsec, human readable prefix found '%s'\n", nsec, prf)
			os.Exit(1)
		}
		// we got a valid nsec
		skh := val.(B)
		if _, err = hex.DecBytes(sec, skh); err != nil {
			fmt.Printf("error decoding hex '%s': %v\n", skh, err)
			os.Exit(1)
		}
	} else {
		if _, err = hex.DecBytes(sec, nsec); err != nil {
			fmt.Printf("error decoding hex '%s': %v\n", nsec, err)
			os.Exit(1)
		}
	}
	// now we have the secret, derive the compressed pubkey
	signer := &p256k.Signer{}
	if err = signer.InitSec(sec); err != nil {
		if err.Error() != "provided secret generates a public key with odd Y coordinate, fixed version returned" {
			fmt.Printf("error initializing sec: %v\n", err)
			fmt.Printf("%0x", signer.Sec())
			os.Exit(1)
		}
	}
	if err != nil {
		fmt.Printf("# fixed key!\n")
	} else {
		fmt.Printf("# key was already correct\n")
	}
	ecpkb := signer.ECPub()
	if ecpkb[0] == 3 {
		signer.Negate()
	}
	fmt.Printf("HEXSEC = %0x\n", signer.Sec())
	fmt.Printf("HEXPUB = %0x\n", signer.Pub())
	var np, ns B
	ns, err = bech32encoding.BinToNsec(signer.Sec())
	chk.E(err)
	np, err = bech32encoding.BinToNpub(signer.Pub())
	chk.E(err)
	fmt.Printf("NSEC = %s\n", ns)
	fmt.Printf("NPUB = %s\n", np)
}
