package bech32m

import "errors"

var (
	ErrMaxLengthExceeded = errors.New("bech32m: overall max length exceeded")
	ErrMixedCase         = errors.New("bech32m: mixed case found in address")
	ErrInvalidChecksum   = errors.New("bech32m: invalid checksum")
	ErrCorruptInput      = errors.New("bech32m: corrupt base32 data")
)
