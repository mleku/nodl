# vainstr
nostr vanity key miner

## usage

```
Usage: vainstr [--threads THREADS] [STRING [POSITION]]

Positional arguments:
  STRING
  POSITION               [begin|contain|end]

Options:
  --threads THREADS      number of threads to mine with - defaults to using all CPU threads available
  --help, -h             display this help and exit
```

## install

if you haven't installed the bitcoin-core cgo or don't want to use it, this 
will work for anywhere:

```bash
go install -tags btcec git.replicatr.dev/cmd/vainstr@v0.0.8
```

note that this version is very much slower, ~1/4 the speed

if you follow the directions to build and install the dependencies for 
libsecp256k1 found in this repository here: 
[../../pkg/crypto/p256k](../../pkg/crypto/p256k) you can leave out the build 
tag:

```bash
go install git.replicatr.dev/cmd/vainstr@v0.0.8
```

a linux binary release will also be available in the github releases page... 
for windows users, this can be run using WSL2 directly 
[https://github.com/mleku/nodl/releases](https://github.com/mleku/nodl/releases)
