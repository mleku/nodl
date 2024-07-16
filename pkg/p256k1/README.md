# p256k1

This is a library that uses the `bitcoin-core` optimized secp256k1 elliptic
curve signatures library for `nostr` schnorr signatures.

If you don't want to use cgo or can't use cgo you need to set the `btcec` 
build tag, which will set an override to use the code from the 
[btcsuite](https://github.com/btcsuite/btcd)/[decred](https://github.com/decred/dcrd/tree/master/dcrec), 
the decred is actually where the schnorr signatures are (ikr?) - this repo 
uses my fork of this mess of shitcoinery and bad, slow Go code is cleaned up 
and unified in [github.com/mleku/btcec](https://github.com/mleku/btcec) and 
includes the bech32 and base58check libraries. 

The directory `pkg/libsecp256k1/secp256k1` needs to be initialized and built
and installed, like so:

```bash
cd pkg/p256k1/secp256k1
git submodule init
git submodule update
```

Then to build, you can refer to the [instructions](./secp256k1/README.md) or
just use the default autotools:

```bash
./autogen.sh
./configure --enable-module-schnorrsig --prefix=/usr
make
sudo make install
```