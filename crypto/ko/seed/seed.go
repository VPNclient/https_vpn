package seed

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
)

// seedCipher implements the cipher.Block interface for SEED
type seedCipher struct {
	k [32]uint32 // round keys
}

// NewCipher creates a new SEED cipher with the given key
func NewCipher(key []byte) (cipher.Block, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("seed: invalid key size %d (must be 16)", len(key))
	}

	c := &seedCipher{}
	c.expandKey(key)
	return c, nil
}

// BlockSize returns the SEED block size
func (c *seedCipher) BlockSize() int {
	return BlockSize
}

// Encrypt encrypts a single block
func (c *seedCipher) Encrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("seed: input not full block")
	}
	if len(dst) < BlockSize {
		panic("seed: output not full block")
	}
	c.encrypt(dst, src)
}

// Decrypt decrypts a single block
func (c *seedCipher) Decrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("seed: input not full block")
	}
	if len(dst) < BlockSize {
		panic("seed: output not full block")
	}
	c.decrypt(dst, src)
}

// expandKey performs SEED key expansion
func (c *seedCipher) expandKey(key []byte) {
	// Split key into K0, K1, K2, K3 (big-endian)
	k0 := binary.BigEndian.Uint32(key[0:4])
	k1 := binary.BigEndian.Uint32(key[4:8])
	k2 := binary.BigEndian.Uint32(key[8:12])
	k3 := binary.BigEndian.Uint32(key[12:16])

	for i := 0; i < Rounds; i++ {
		t0 := k0 + k2 - kc[i]
		t1 := k1 - k3 + kc[i]

		c.k[2*i] = g(t0)
		c.k[2*i+1] = g(t1)

		if i%2 == 0 {
			// Rotate K0||K1 left by 8 bits
			t := k0 >> 24
			k0 = (k0 << 8) | (k1 >> 24)
			k1 = (k1 << 8) | t
		} else {
			// Rotate K2||K3 right by 8 bits
			t := k3 & 0xff
			k3 = (k3 >> 8) | (k2 << 24)
			k2 = (k2 >> 8) | (t << 24)
		}
	}
}

// g is the G function
func g(x uint32) uint32 {
	return ss0[byte(x)] ^
		ss1[byte(x>>8)] ^
		ss2[byte(x>>16)] ^
		ss3[byte(x>>24)]
}

// f is the F function (Feistel function)
func f(c0, c1, k0, k1 uint32) (uint32, uint32) {
	t0 := c0 ^ k0
	t1 := c1 ^ k1
	t1 = t1 ^ t0

	t1 = g(t1)
	t0 = t0 + t1
	t0 = g(t0)
	t1 = t1 + t0
	t1 = g(t1)
	t0 = t0 + t1

	return t0, t1
}

// encrypt performs SEED encryption
func (c *seedCipher) encrypt(dst, src []byte) {
	// Load plaintext (big-endian)
	l0 := binary.BigEndian.Uint32(src[0:4])
	l1 := binary.BigEndian.Uint32(src[4:8])
	r0 := binary.BigEndian.Uint32(src[8:12])
	r1 := binary.BigEndian.Uint32(src[12:16])

	// 16 rounds
	for i := 0; i < Rounds; i++ {
		t0, t1 := f(r0, r1, c.k[2*i], c.k[2*i+1])
		t0 ^= l0
		t1 ^= l1
		l0, l1 = r0, r1
		r0, r1 = t0, t1
	}

	// Store ciphertext (big-endian)
	binary.BigEndian.PutUint32(dst[0:4], r0)
	binary.BigEndian.PutUint32(dst[4:8], r1)
	binary.BigEndian.PutUint32(dst[8:12], l0)
	binary.BigEndian.PutUint32(dst[12:16], l1)
}

// decrypt performs SEED decryption
func (c *seedCipher) decrypt(dst, src []byte) {
	// Load ciphertext (big-endian)
	r0 := binary.BigEndian.Uint32(src[0:4])
	r1 := binary.BigEndian.Uint32(src[4:8])
	l0 := binary.BigEndian.Uint32(src[8:12])
	l1 := binary.BigEndian.Uint32(src[12:16])

	// 16 rounds (reverse order)
	for i := Rounds - 1; i >= 0; i-- {
		t0, t1 := f(l0, l1, c.k[2*i], c.k[2*i+1])
		t0 ^= r0
		t1 ^= r1
		r0, r1 = l0, l1
		l0, l1 = t0, t1
	}

	// Store plaintext (big-endian)
	binary.BigEndian.PutUint32(dst[0:4], l0)
	binary.BigEndian.PutUint32(dst[4:8], l1)
	binary.BigEndian.PutUint32(dst[8:12], r0)
	binary.BigEndian.PutUint32(dst[12:16], r1)
}
