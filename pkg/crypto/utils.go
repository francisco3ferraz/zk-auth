package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomBigInt(bits int) (*big.Int, error) {
	n, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func Hash(data ...[]byte) []byte {
	h := sha256.New()

	for _, d := range data {
		h.Write(d)
	}

	return h.Sum(nil)
}

func PadTo(b []byte, length int) []byte {
	if len(b) >= length {
		return b
	}

	padded := make([]byte, length)
	copy(padded[length-len(b):], b)

	return padded
}

func ConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
