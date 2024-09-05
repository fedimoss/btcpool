package bech32m

const (
	Bech32  = 1
	Bech32m = 2
	Failed  = -1

	charset      = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	bech32mConst = 0x2bc830a3
)

var deccharset = makeDecoder(charset)

func makeDecoder(charset string) (res [128]byte) {
	for i := range res {
		res[i] = 0xff
	}

	for i, c := range charset {
		res[c] = byte(i)
	}
	return
}
