# keyfix

This is a little tool to check if a nostr secret key generates a BIP-340 
compliant even Y coordinate in the public key.

Nostr secret keys that have odd Y coordinate public keys do cannot receive 
encrypted direct messages.

It accepts both nsec and hexadecimal formatted nostr secret keys and prints out 
the correct information in bash shell format:

```bash
# $ go run ./cmd/keyfix/. d5633530f5bcfebceb5584cfbbf718a30df0751b729dd9a789b9f30c0587d74e
# fixed key!
HEXSEC = 2a9ccacf0a43014314aa7b304408e75bacbe67cb3caac69436186b80caae69f3
HEXPUB = ff17bf710b09d1d36093c7af1a3ea9a8f43df3443bc51b84d5ea8a50db61807d
NSEC = nsec192wv4nc2gvq5x9920vcygz88twktue7t8j4vd9pkrp4cpj4wd8ess7hpc7
NPUB = npub1lutm7ugtp8gaxcync7h3504f4r6rmu6y80z3hpx4a299pkmpsp7spjf96c
# $ go run ./cmd/keyfix/. nsec192wv4nc2gvq5x9920vcygz88twktue7t8j4vd9pkrp4cpj4wd8ess7hpc7
# key was already correct
HEXSEC = 2a9ccacf0a43014314aa7b304408e75bacbe67cb3caac69436186b80caae69f3
HEXPUB = ff17bf710b09d1d36093c7af1a3ea9a8f43df3443bc51b84d5ea8a50db61807d
NSEC = nsec192wv4nc2gvq5x9920vcygz88twktue7t8j4vd9pkrp4cpj4wd8ess7hpc7
NPUB = npub1lutm7ugtp8gaxcync7h3504f4r6rmu6y80z3hpx4a299pkmpsp7spjf96c
```

## building

Use the `btcec` build tag if you don't want to or can't use the 
[secp256k1](https://github.com/bitcoin-core/secp256k1) CGO.

Run it directly like this:

```bash
go run -tags btcec git.replicatr.dev/cmd/keyfix@latest
```

If you have got the development headers and library objects available for 
the secp256k1 library you can omit `-tags btcec` but it is so much faster 
this is the implicit default, if you also want to use `vainstr` also found 
in this repository.