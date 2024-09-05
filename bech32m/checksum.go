package bech32m

// verifyChecksum verifies a checksum given HRP and converted data characters.
func verifyChecksum(hrp string, data []byte) int {
	c := polymodHrp(hrp, data)
	if c == 1 {
		return Bech32
	}
	if c == bech32mConst {
		return Bech32m
	}
	return Failed
}

// createChecksum computes the checksum values given HRP and data.
func createChecksum(hrp string, data []byte, spec int) []byte {
	// Compute the checksum values given HRP and data.
	c := uint32(1)
	if spec == Bech32m {
		c = bech32mConst
	}
	mod := polymodHrp(hrp, data, []byte{0, 0, 0, 0, 0, 0}) ^ c
	ret := make([]byte, 6)
	for i := 0; i < len(ret); i++ {
		ret[i] = byte(mod>>uint32(5*(5-i))) & 31
	}
	return ret
}
