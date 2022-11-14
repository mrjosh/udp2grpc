package protocol

import "encoding/base64"

const (
	NoisePublicKeySize    = 32
	NoisePrivateKeySize   = 32
	NoisePresharedKeySize = 32
)

type (
	NoisePublicKey  [NoisePublicKeySize]byte
	NoisePrivateKey [NoisePrivateKeySize]byte
	NoiseNonce      uint64 // padded to 12-bytes
)

func (key *NoisePrivateKey) String() string {
	return base64.StdEncoding.EncodeToString(key[:])
}
