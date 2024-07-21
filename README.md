
![nostr canary](assets/icon.png)

# nodl

Go ~~relay, client and (TODO)~~ libraries for the nostr protocol with a 
focus on 
performance

## building

see [p256k1 docs](./pkg/p256k1/README.md) for building with the 
`bitcoin-core/secp256k1` library interfaced with CGO (it is about 2x faster 
at verification and 4x faster at signing) but if you don't want to use CGO 
or can't, set the build tag `btcec` to disable the `secp256k1` CGO binding 
interface.