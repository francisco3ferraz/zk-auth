package crypto

import (
	"crypto/rand"
	"crypto/sha256"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Hash(data ...[]byte) []byte {
	h := sha256.New()

	for _, d := range data {
		h.Write(d)
	}

	return h.Sum(nil)
}
