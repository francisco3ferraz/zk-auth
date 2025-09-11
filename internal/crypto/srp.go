package crypto

import (
	"fmt"
	"math/big"
)

type SRP struct {
	N *big.Int
	G *big.Int
	K *big.Int
}

func NewSRP() *SRP {
	return &SRP{
		N: N,
		G: G,
		K: K,
	}
}

func (s *SRP) GenerateSalt() ([]byte, error) {
	return GenerateRandomBytes(SaltLength)
}

// Compute the password verifier v = g^x mod N
// x = H(salt | H(username | ":" | password))
func (s *SRP) ComputeVerifier(username, password string, salt []byte) (*big.Int, error) {
	x := s.computeX(username, password, salt)
	v := new(big.Int).Exp(s.G, x, s.N)
	return v, nil
}

// Compute x = H(salt | H(username | ":" | password))
func (s *SRP) computeX(username, password string, salt []byte) *big.Int {
	// H(username | ":" | password)
	credentials := fmt.Sprintf("%s:%s", username, password)
	hCred := Hash([]byte(credentials))

	// x = H(salt | H(username | ":" | password))
	xBytes := Hash(salt, hCred)
	return new(big.Int).SetBytes(xBytes)
}
