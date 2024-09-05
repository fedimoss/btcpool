package bech32m

// code from encoding/base32 adapted to this specific use case
// bech32(m) uses base32 encoding with padding

func base32EncLen(n int) int {
	return n/5*8 + (n%5*8+4)/5
}

// base32Encode will encode the 8 bits src into 5 bits values in dst, including padding
func base32Encode(dst, src []byte) {
	if len(src) == 0 {
		return
	}
	di, si := 0, 0
	n := (len(src) / 5) * 5
	for si < n {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		hi := uint32(src[si+0])<<24 | uint32(src[si+1])<<16 | uint32(src[si+2])<<8 | uint32(src[si+3])
		lo := hi<<8 | uint32(src[si+4])

		dst[di+0] = byte((hi >> 27) & 0x1F)
		dst[di+1] = byte((hi >> 22) & 0x1F)
		dst[di+2] = byte((hi >> 17) & 0x1F)
		dst[di+3] = byte((hi >> 12) & 0x1F)
		dst[di+4] = byte((hi >> 7) & 0x1F)
		dst[di+5] = byte((hi >> 2) & 0x1F)
		dst[di+6] = byte((lo >> 5) & 0x1F)
		dst[di+7] = byte((lo) & 0x1F)

		si += 5
		di += 8
	}

	// Add the remaining small block
	remain := len(src) - si
	if remain == 0 {
		return
	}

	// Encode the remaining bytes in reverse order.
	val := uint32(0)
	switch remain {
	case 4:
		val |= uint32(src[si+3])
		dst[di+6] = byte(val << 3 & 0x1F)
		dst[di+5] = byte(val >> 2 & 0x1F)
		fallthrough
	case 3:
		val |= uint32(src[si+2]) << 8
		dst[di+4] = byte(val >> 7 & 0x1F)
		fallthrough
	case 2:
		val |= uint32(src[si+1]) << 16
		dst[di+3] = byte(val >> 12 & 0x1F)
		dst[di+2] = byte(val >> 17 & 0x1F)
		fallthrough
	case 1:
		val |= uint32(src[si+0]) << 24
		dst[di+1] = byte(val >> 22 & 0x1F)
		dst[di+0] = byte(val >> 27 & 0x1F)
	}
}

func base32DecLen(n int) int {
	return n/8*5 + n%8*5/8
}

func base32Decode(dst, src []byte) (n int, end bool, err error) {
	dsti := 0

	for len(src) > 0 && !end {
		// Decode quantum using the base32 alphabet
		var dbuf [8]byte
		dlen := 8

		for j := 0; j < 8; {
			if len(src) == 0 {
				// We have reached the end and are not expecting any padding
				dlen, end = j, true
				break
			}
			in := src[0]
			src = src[1:]
			dbuf[j] = in
			if in > 0x1f {
				return n, false, ErrCorruptInput
			}
			j++
		}

		// Pack 8x 5-bit source blocks into 5 byte destination
		// quantum
		switch dlen {
		case 8:
			dst[dsti+4] = dbuf[6]<<5 | dbuf[7]
			n++
			fallthrough
		case 7:
			dst[dsti+3] = dbuf[4]<<7 | dbuf[5]<<2 | dbuf[6]>>3
			n++
			fallthrough
		case 5:
			dst[dsti+2] = dbuf[3]<<4 | dbuf[4]>>1
			n++
			fallthrough
		case 4:
			dst[dsti+1] = dbuf[1]<<6 | dbuf[2]<<1 | dbuf[3]>>4
			n++
			fallthrough
		case 2:
			dst[dsti+0] = dbuf[0]<<3 | dbuf[1]>>2
			n++
		}
		dsti += 5
	}
	return n, end, nil
}
