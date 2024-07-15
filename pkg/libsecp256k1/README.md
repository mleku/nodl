# libsecp256k1

This is a library that uses the `bitcoin-core` optimized secp256k1 elliptic
curve signatures library.

The directory `pkg/libsecp256k1/secp256k1` needs to be initialized and built
and installed, like so:

```bash
cd pkg/libsecp256k1/secp256k1
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

In order to enable it, you need to set the Go build tag `libsecp256k1`. This
will replace the `btcec` interface version with the much faster `cgo` linked C
version.