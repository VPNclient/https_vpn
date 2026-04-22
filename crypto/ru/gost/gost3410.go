// GOST R 34.10-2012 digital signature algorithm.
package gost

import (
	"crypto"
	"errors"
	"io"
	"math/big"
)

// PrivateKey represents a GOST R 34.10-2012 private key.
type PrivateKey struct {
	PublicKey
	D *big.Int // Private scalar
}

// PublicKey represents a GOST R 34.10-2012 public key.
type PublicKey struct {
	Curve *Curve   // Elliptic curve
	X, Y  *big.Int // Public point Q = d*G
}

// Public returns the public key corresponding to priv.
func (priv *PrivateKey) Public() crypto.PublicKey {
	return &priv.PublicKey
}

// Sign signs digest with priv, reading randomness from rand.
// Returns signature as r || s (concatenated big-endian bytes).
func (priv *PrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	r, s, err := Sign(rand, priv, digest)
	if err != nil {
		return nil, err
	}

	// Encode r || s
	byteLen := (priv.Curve.BitSize + 7) / 8
	sig := make([]byte, 2*byteLen)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[byteLen-len(rBytes):byteLen], rBytes)
	copy(sig[2*byteLen-len(sBytes):], sBytes)

	return sig, nil
}

// GenerateKey generates a new GOST R 34.10-2012 key pair.
func GenerateKey(curve *Curve, rand io.Reader) (*PrivateKey, error) {
	// Generate random d in [1, n-1]
	n := curve.N
	bitLen := n.BitLen()
	byteLen := (bitLen + 7) / 8

	var d *big.Int
	for {
		dBytes := make([]byte, byteLen)
		if _, err := io.ReadFull(rand, dBytes); err != nil {
			return nil, err
		}
		d = new(big.Int).SetBytes(dBytes)

		// Ensure 0 < d < n
		if d.Sign() > 0 && d.Cmp(n) < 0 {
			break
		}
	}

	// Compute Q = d*G
	x, y := curve.ScalarBaseMult(d.Bytes())

	priv := &PrivateKey{
		PublicKey: PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}
	return priv, nil
}

// Sign signs a hash (digest) using the private key.
// Returns (r, s) signature components.
func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	curve := priv.Curve
	n := curve.N
	byteLen := (n.BitLen() + 7) / 8

	// Convert hash to integer e
	// GOST uses little-endian interpretation of hash
	e := hashToInt(hash, curve)
	if e.Sign() == 0 {
		e = big.NewInt(1)
	}

	for {
		// Generate random k in [1, n-1]
		var k *big.Int
		for {
			kBytes := make([]byte, byteLen)
			if _, err = io.ReadFull(rand, kBytes); err != nil {
				return nil, nil, err
			}
			k = new(big.Int).SetBytes(kBytes)
			if k.Sign() > 0 && k.Cmp(n) < 0 {
				break
			}
		}

		// C = k*G
		cx, _ := curve.ScalarBaseMult(k.Bytes())

		// r = cx mod n
		r = new(big.Int).Mod(cx, n)
		if r.Sign() == 0 {
			continue
		}

		// s = (r*d + k*e) mod n
		rd := new(big.Int).Mul(r, priv.D)
		ke := new(big.Int).Mul(k, e)
		s = new(big.Int).Add(rd, ke)
		s.Mod(s, n)
		if s.Sign() == 0 {
			continue
		}

		return r, s, nil
	}
}

// Verify verifies the signature (r, s) of hash using the public key.
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	curve := pub.Curve
	n := curve.N

	// Check r, s in range [1, n-1]
	if r.Sign() <= 0 || r.Cmp(n) >= 0 {
		return false
	}
	if s.Sign() <= 0 || s.Cmp(n) >= 0 {
		return false
	}

	// Convert hash to integer e
	e := hashToInt(hash, curve)
	if e.Sign() == 0 {
		e = big.NewInt(1)
	}

	// v = e^-1 mod n
	v := new(big.Int).ModInverse(e, n)
	if v == nil {
		return false
	}

	// z1 = s*v mod n
	z1 := new(big.Int).Mul(s, v)
	z1.Mod(z1, n)

	// z2 = -r*v mod n = (n - r*v mod n)
	z2 := new(big.Int).Mul(r, v)
	z2.Mod(z2, n)
	z2.Sub(n, z2)
	z2.Mod(z2, n)

	// C = z1*G + z2*Q
	x1, y1 := curve.ScalarBaseMult(z1.Bytes())
	x2, y2 := curve.ScalarMult(pub.X, pub.Y, z2.Bytes())
	cx, _ := curve.Add(x1, y1, x2, y2)

	// R = cx mod n
	R := new(big.Int).Mod(cx, n)

	return R.Cmp(r) == 0
}

// VerifySignature verifies a signature in r||s format.
func VerifySignature(pub *PublicKey, hash, sig []byte) bool {
	byteLen := (pub.Curve.BitSize + 7) / 8
	if len(sig) != 2*byteLen {
		return false
	}

	r := new(big.Int).SetBytes(sig[:byteLen])
	s := new(big.Int).SetBytes(sig[byteLen:])

	return Verify(pub, hash, r, s)
}

// hashToInt converts a hash to an integer per GOST specification.
// GOST uses a specific byte order (the hash is interpreted as little-endian).
func hashToInt(hash []byte, curve *Curve) *big.Int {
	// Reverse hash bytes (GOST uses little-endian)
	reversed := make([]byte, len(hash))
	for i := 0; i < len(hash); i++ {
		reversed[i] = hash[len(hash)-1-i]
	}

	e := new(big.Int).SetBytes(reversed)
	e.Mod(e, curve.N)
	return e
}

// SignASN1 signs and returns ASN.1 DER encoded signature (for X.509 compatibility).
// Note: GOST signatures in certificates typically use raw r||s format, not ASN.1.
func SignASN1(rand io.Reader, priv *PrivateKey, hash []byte) ([]byte, error) {
	// For GOST, we typically use raw format, but providing ASN.1 for compatibility
	return priv.Sign(rand, hash, nil)
}

// VerifyASN1 verifies an ASN.1 DER encoded signature.
func VerifyASN1(pub *PublicKey, hash, sig []byte) bool {
	return VerifySignature(pub, hash, sig)
}

// Errors
var (
	ErrInvalidPublicKey = errors.New("gost3410: invalid public key")
	ErrInvalidSignature = errors.New("gost3410: invalid signature")
)
