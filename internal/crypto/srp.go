package crypto

import (
	"crypto/sha256"
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

// Compute the server's session key
// S = (A * v^u)^b mod N
// K = H(S)
func (s *SRP) ComputeServerSessionKey(A, b, v *big.Int, u *big.Int) ([]byte, error) {
	// Validate A
	if A.Mod(A, s.N).Sign() == 0 {
		return nil, fmt.Errorf("invalid A value")
	}

	// S = (A * v^u)^b mod N
	vu := new(big.Int).Exp(v, u, s.N)
	Avu := new(big.Int).Mul(A, vu)
	Avu.Mod(Avu, s.N)
	S := new(big.Int).Exp(Avu, b, s.N)

	// K = H(S)
	return Hash(S.Bytes()), nil
}

// Compute the server proof M2 = H(A | M1 | K)
func (s *SRP) ComputeServerProof(A *big.Int, M1, K []byte) []byte {
	return Hash(A.Bytes(), M1, K)
}

// Compute the client proof M1 = H(H(N) XOR H(g) | H(username) | salt | A | B | K)
func (s *SRP) ComputeClientProof(username string, salt []byte, A, B *big.Int, K []byte) []byte {
	// H(N) XOR H(g)
	hN := Hash(s.N.Bytes())
	hG := Hash(s.G.Bytes())
	hNxorG := make([]byte, len(hN))
	for i := range hN {
		hNxorG[i] = hN[i] ^ hG[i]
	}

	// H(username)
	hU := Hash([]byte(username))

	// M1 = H(H(N) XOR H(g) | H(username) | salt | A | B | K)
	return Hash(hNxorG, hU, salt, A.Bytes(), B.Bytes(), K)
}

// Generates server ephemeral keys (b, B)
// B = k*v + g^b mod N
func (s *SRP) GenerateServerKeys(verifier *big.Int) (b, B *big.Int, err error) {
	b, err = GenerateRandomBigInt(256)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate b: %w", err)
	}

	// B = k*v + g^b mod N
	gb := new(big.Int).Exp(s.G, b, s.N)
	kv := new(big.Int).Mul(s.K, verifier)
	B = new(big.Int).Add(kv, gb)
	B.Mod(B, s.N)

	if B.Sign() == 0 {
		return s.GenerateServerKeys(verifier) // Retry
	}

	return b, B, nil
}

// Verify the client's proof
func (s *SRP) VerifyClientProof(username string, salt []byte, A, B *big.Int, K, clientM1 []byte) bool {
	expectedM1 := s.ComputeClientProof(username, salt, A, B, K)
	return ConstantTimeCompare(expectedM1, clientM1)
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

// Compute u = H(A | B)
func (s *SRP) ComputeU(A, B *big.Int) *big.Int {
	h := sha256.New()
	h.Write(PadTo(A.Bytes(), 256))
	h.Write(PadTo(B.Bytes(), 256))
	return new(big.Int).SetBytes(h.Sum(nil))
}
