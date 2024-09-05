package bech32m

import (
	"testing"
)

var original_generator = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

func original_polymod(values []byte) int {
	// Internal function that computes the Bech32 checksum.
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ int(v)
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= original_generator[i]
			} else {
				chk ^= 0
			}
		}
	}
	return chk
}

func original_hrpExpand(hrp string) []byte {
	// Expand the HRP into values for checksum computation.
	ret := []byte{}
	for _, c := range hrp {
		ret = append(ret, byte(c>>5))
	}
	ret = append(ret, 0)
	for _, c := range hrp {
		ret = append(ret, byte(c&31))
	}
	return ret
}

type polymodTestV struct {
	hrp  string
	data [][]byte
}

func pTestV(hrp string, data ...[]byte) *polymodTestV {
	return &polymodTestV{hrp: hrp, data: data}
}

func TestPolymod(t *testing.T) {
	// compare results from original polymod implementation and ours
	testV := []*polymodTestV{
		pTestV("bc", []byte{0, 0, 0, 0, 5, 8, 3, 2}),
	}

	for _, vect := range testV {
		buf := original_hrpExpand(vect.hrp)
		for _, n := range vect.data {
			buf = append(buf, n...)
		}
		a := original_polymod(buf)
		b := polymodHrp(vect.hrp, vect.data...)

		if a != int(b) {
			t.Errorf("polymod result difference, 0x%x != 0x%x", a, b)
		}
	}
}
