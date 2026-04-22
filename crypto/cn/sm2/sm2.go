// Package sm2 implements the SM2 elliptic curve cryptography per GB/T 32918.
package sm2

import (
	"crypto"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"io"
	"math/big"

	"github.com/nativemind/https-vpn/crypto/cn/sm3"
)

var (
	ErrInvalidPublicKey  = errors.New("sm2: invalid public key")
	ErrInvalidPrivateKey = errors.New("sm2: invalid private key")
	ErrInvalidSignature  = errors.New("sm2: invalid signature")
	ErrDecryption        = errors.New("sm2: decryption error")
)

// PublicKey represents an SM2 public key.
type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

// PrivateKey represents an SM2 private key.
type PrivateKey struct {
	PublicKey
	D *big.Int
}

// Public returns the public key corresponding to priv.
func (priv *PrivateKey) Public() crypto.PublicKey {
	return &priv.PublicKey
}

// Sign signs digest with priv, reading randomness from rand.
// The signature is returned as r || s (each 32 bytes, big-endian).
func (priv *PrivateKey) Sign(random io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	r, s, err := Sign(random, priv, digest)
	if err != nil {
		return nil, err
	}

	// Encode as r || s (each 32 bytes)
	sig := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	return sig, nil
}

// GenerateKey generates a new SM2 private key.
func GenerateKey(random io.Reader) (*PrivateKey, error) {
	curve := P256()
	if random == nil {
		random = rand.Reader
	}

	params := curve.Params()
	b := make([]byte, params.BitSize/8+8)
	if _, err := io.ReadFull(random, b); err != nil {
		return nil, err
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, big.NewInt(2))
	k.Mod(k, n)
	k.Add(k, big.NewInt(1))

	priv := new(PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())

	return priv, nil
}

// Sign signs hash using the private key.
// Returns (r, s) signature components.
func Sign(random io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	if random == nil {
		random = rand.Reader
	}

	curve := priv.Curve
	n := curve.Params().N
	e := new(big.Int).SetBytes(hash)

	for {
		// Generate random k
		k, err := randFieldElement(curve, random)
		if err != nil {
			return nil, nil, err
		}

		// (x1, y1) = [k]G
		x1, _ := curve.ScalarBaseMult(k.Bytes())

		// r = (e + x1) mod n
		r = new(big.Int).Add(e, x1)
		r.Mod(r, n)

		// Check r != 0 and r + k != n
		if r.Sign() == 0 {
			continue
		}
		rk := new(big.Int).Add(r, k)
		if rk.Cmp(n) == 0 {
			continue
		}

		// s = ((1 + d)^-1 * (k - r*d)) mod n
		d1 := new(big.Int).Add(priv.D, big.NewInt(1))
		d1Inv := new(big.Int).ModInverse(d1, n)
		if d1Inv == nil {
			continue
		}

		rd := new(big.Int).Mul(r, priv.D)
		krd := new(big.Int).Sub(k, rd)
		krd.Mod(krd, n)

		s = new(big.Int).Mul(d1Inv, krd)
		s.Mod(s, n)

		if s.Sign() != 0 {
			return r, s, nil
		}
	}
}

// Verify verifies the signature (r, s) of hash using the public key.
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	curve := pub.Curve
	n := curve.Params().N

	// Check r, s in [1, n-1]
	if r.Sign() <= 0 || r.Cmp(n) >= 0 {
		return false
	}
	if s.Sign() <= 0 || s.Cmp(n) >= 0 {
		return false
	}

	e := new(big.Int).SetBytes(hash)

	// t = (r + s) mod n
	t := new(big.Int).Add(r, s)
	t.Mod(t, n)
	if t.Sign() == 0 {
		return false
	}

	// (x1, y1) = [s]G + [t]P
	x1g, y1g := curve.ScalarBaseMult(s.Bytes())
	x1p, y1p := curve.ScalarMult(pub.X, pub.Y, t.Bytes())
	x1, _ := curve.Add(x1g, y1g, x1p, y1p)

	// R = (e + x1) mod n
	R := new(big.Int).Add(e, x1)
	R.Mod(R, n)

	return R.Cmp(r) == 0
}

// VerifySignature verifies a signature in r||s format.
func VerifySignature(pub *PublicKey, hash, sig []byte) bool {
	if len(sig) != 64 {
		return false
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return Verify(pub, hash, r, s)
}

// ComputeZA computes the Z value for SM2 signature (user ID hash).
// Z = SM3(ENTL || ID || a || b || Gx || Gy || Px || Py)
func ComputeZA(pub *PublicKey, uid []byte) []byte {
	if uid == nil {
		uid = []byte("1234567812345678") // Default ID
	}

	curve := pub.Curve
	params := curve.Params()
	a := sm2A()

	// ENTL is the bit length of ID (2 bytes, big-endian)
	entl := len(uid) * 8
	if entl > 65535 {
		entl = 65535
	}

	h := sm3.New()
	h.Write([]byte{byte(entl >> 8), byte(entl)})
	h.Write(uid)
	h.Write(toBytes32(a))
	h.Write(toBytes32(params.B))
	h.Write(toBytes32(params.Gx))
	h.Write(toBytes32(params.Gy))
	h.Write(toBytes32(pub.X))
	h.Write(toBytes32(pub.Y))

	return h.Sum(nil)
}

// SignWithID signs message with user ID.
// First computes e = SM3(ZA || M), then signs e.
func SignWithID(random io.Reader, priv *PrivateKey, uid, msg []byte) (r, s *big.Int, err error) {
	za := ComputeZA(&priv.PublicKey, uid)

	h := sm3.New()
	h.Write(za)
	h.Write(msg)
	e := h.Sum(nil)

	return Sign(random, priv, e)
}

// VerifyWithID verifies signature with user ID.
func VerifyWithID(pub *PublicKey, uid, msg []byte, r, s *big.Int) bool {
	za := ComputeZA(pub, uid)

	h := sm3.New()
	h.Write(za)
	h.Write(msg)
	e := h.Sum(nil)

	return Verify(pub, e, r, s)
}

// Encrypt encrypts plaintext using public key (SM2 encryption).
// Output format: C1 || C3 || C2 (new standard, GB/T 32918.4-2016)
func Encrypt(random io.Reader, pub *PublicKey, plaintext []byte) ([]byte, error) {
	if random == nil {
		random = rand.Reader
	}

	curve := pub.Curve

	for {
		// Generate random k
		k, err := randFieldElement(curve, random)
		if err != nil {
			return nil, err
		}

		// C1 = [k]G (point, uncompressed: 04 || x || y)
		x1, y1 := curve.ScalarBaseMult(k.Bytes())
		c1 := make([]byte, 65)
		c1[0] = 0x04
		copy(c1[1:33], toBytes32(x1))
		copy(c1[33:65], toBytes32(y1))

		// (x2, y2) = [k]P
		x2, y2 := curve.ScalarMult(pub.X, pub.Y, k.Bytes())

		// t = KDF(x2 || y2, klen)
		kdfInput := append(toBytes32(x2), toBytes32(y2)...)
		t := kdf(kdfInput, len(plaintext))

		// Check t is not all zeros
		allZero := true
		for _, b := range t {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			continue
		}

		// C2 = M XOR t
		c2 := make([]byte, len(plaintext))
		for i := range plaintext {
			c2[i] = plaintext[i] ^ t[i]
		}

		// C3 = SM3(x2 || M || y2)
		h := sm3.New()
		h.Write(toBytes32(x2))
		h.Write(plaintext)
		h.Write(toBytes32(y2))
		c3 := h.Sum(nil)

		// Output: C1 || C3 || C2
		result := make([]byte, 65+32+len(c2))
		copy(result[0:65], c1)
		copy(result[65:97], c3)
		copy(result[97:], c2)

		return result, nil
	}
}

// Decrypt decrypts ciphertext using private key.
// Input format: C1 || C3 || C2 (new standard)
func Decrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 65+32+1 {
		return nil, ErrDecryption
	}

	curve := priv.Curve

	// Parse C1 (uncompressed point)
	if ciphertext[0] != 0x04 {
		return nil, ErrDecryption
	}
	x1 := new(big.Int).SetBytes(ciphertext[1:33])
	y1 := new(big.Int).SetBytes(ciphertext[33:65])

	// Verify C1 is on curve
	if !curve.IsOnCurve(x1, y1) {
		return nil, ErrDecryption
	}

	// Parse C3 and C2
	c3 := ciphertext[65:97]
	c2 := ciphertext[97:]

	// (x2, y2) = [d]C1
	x2, y2 := curve.ScalarMult(x1, y1, priv.D.Bytes())

	// t = KDF(x2 || y2, klen)
	kdfInput := append(toBytes32(x2), toBytes32(y2)...)
	t := kdf(kdfInput, len(c2))

	// Check t is not all zeros
	allZero := true
	for _, b := range t {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return nil, ErrDecryption
	}

	// M = C2 XOR t
	plaintext := make([]byte, len(c2))
	for i := range c2 {
		plaintext[i] = c2[i] ^ t[i]
	}

	// Verify C3 = SM3(x2 || M || y2)
	h := sm3.New()
	h.Write(toBytes32(x2))
	h.Write(plaintext)
	h.Write(toBytes32(y2))
	u := h.Sum(nil)

	for i := range c3 {
		if c3[i] != u[i] {
			return nil, ErrDecryption
		}
	}

	return plaintext, nil
}

// kdf is the key derivation function per GB/T 32918.4.
// Uses SM3 as the hash function.
func kdf(z []byte, klen int) []byte {
	ct := 1
	k := make([]byte, 0, klen)

	for len(k) < klen {
		h := sm3.New()
		h.Write(z)
		h.Write([]byte{byte(ct >> 24), byte(ct >> 16), byte(ct >> 8), byte(ct)})
		k = append(k, h.Sum(nil)...)
		ct++
	}

	return k[:klen]
}

// randFieldElement returns a random element of the field.
func randFieldElement(curve elliptic.Curve, random io.Reader) (*big.Int, error) {
	params := curve.Params()
	b := make([]byte, params.BitSize/8+8)
	if _, err := io.ReadFull(random, b); err != nil {
		return nil, err
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, big.NewInt(2))
	k.Mod(k, n)
	k.Add(k, big.NewInt(1))

	return k, nil
}

// toBytes32 converts a big.Int to a 32-byte big-endian slice.
func toBytes32(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) > 32 {
		return b[len(b)-32:]
	}
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(b):], b)
		return padded
	}
	return b
}
