[![GoDoc](https://godoc.org/github.com/ModChain/bech32m?status.svg)](https://godoc.org/github.com/ModChain/bech32m)

# bech32m

Implementation of bech32m format for segwit addrs.

BIP: https://github.com/bitcoin/bips/blob/master/bip-0350.mediawiki

A lot of this code comes from [Takatoshi Nakagawa's implementation](https://pkg.go.dev/github.com/tnakagawa/goref/bech32m) but with optimizations on various obvious areas, including avoiding allocating memory for polymod calculation and such.


