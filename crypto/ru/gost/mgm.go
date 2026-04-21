// MGM (Multilinear Galois Mode) AEAD per GOST R 34.13-2015.
// This is the AEAD mode used in GOST TLS cipher suites.
package gost

import (
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"
	"errors"
)

const (
	mgmTagSize   = 16 // 128-bit authentication tag
	mgmNonceSize = 16 // 128-bit nonce for Kuznyechik, 8 for Magma
)

// mgm implements the AEAD interface for MGM mode.
type mgm struct {
	cipher    cipher.Block
	blockSize int
	tagSize   int
}

// NewMGM creates an AEAD using MGM mode with the given block cipher.
// For Kuznyechik (128-bit block), nonce is 16 bytes.
// For Magma (64-bit block), nonce is 8 bytes.
func NewMGM(block cipher.Block) (cipher.AEAD, error) {
	blockSize := block.BlockSize()
	if blockSize != 16 && blockSize != 8 {
		return nil, errors.New("gost/mgm: block size must be 8 or 16")
	}
	return &mgm{
		cipher:    block,
		blockSize: blockSize,
		tagSize:   blockSize, // Tag size equals block size
	}, nil
}

func (m *mgm) NonceSize() int {
	return m.blockSize
}

func (m *mgm) Overhead() int {
	return m.tagSize
}

func (m *mgm) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	if len(nonce) != m.NonceSize() {
		panic("gost/mgm: incorrect nonce length")
	}

	ret, ciphertext := sliceForAppend(dst, len(plaintext)+m.tagSize)
	tag := ciphertext[len(plaintext):]

	m.seal(ciphertext[:len(plaintext)], tag, nonce, plaintext, additionalData)
	return ret
}

func (m *mgm) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if len(nonce) != m.NonceSize() {
		panic("gost/mgm: incorrect nonce length")
	}
	if len(ciphertext) < m.tagSize {
		return nil, errors.New("gost/mgm: ciphertext too short")
	}

	tag := ciphertext[len(ciphertext)-m.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-m.tagSize]

	ret, plaintext := sliceForAppend(dst, len(ciphertext))

	expectedTag := make([]byte, m.tagSize)
	m.open(plaintext, expectedTag, nonce, ciphertext, additionalData)

	if subtle.ConstantTimeCompare(tag, expectedTag) != 1 {
		// Clear plaintext on authentication failure
		for i := range plaintext {
			plaintext[i] = 0
		}
		return nil, errors.New("gost/mgm: authentication failed")
	}

	return ret, nil
}

func (m *mgm) seal(ciphertext, tag, nonce, plaintext, ad []byte) {
	// Generate ICN (Initial Counter for encryption) and H (for auth)
	icn := make([]byte, m.blockSize)
	h := make([]byte, m.blockSize)

	// ICN = E_K(1 || nonce[1:])
	copy(icn, nonce)
	icn[0] |= 0x80 // Set MSB to 1
	m.cipher.Encrypt(icn, icn)

	// H = E_K(0 || nonce[1:])
	copy(h, nonce)
	h[0] &= 0x7F // Clear MSB
	m.cipher.Encrypt(h, h)

	// Encrypt plaintext using CTR mode with ICN
	m.ctrEncrypt(ciphertext, plaintext, icn)

	// Compute authentication tag
	m.computeTag(tag, h, ad, ciphertext)
}

func (m *mgm) open(plaintext, tag, nonce, ciphertext, ad []byte) {
	// Generate ICN and H
	icn := make([]byte, m.blockSize)
	h := make([]byte, m.blockSize)

	copy(icn, nonce)
	icn[0] |= 0x80
	m.cipher.Encrypt(icn, icn)

	copy(h, nonce)
	h[0] &= 0x7F
	m.cipher.Encrypt(h, h)

	// Compute expected tag over ciphertext
	m.computeTag(tag, h, ad, ciphertext)

	// Decrypt ciphertext
	m.ctrEncrypt(plaintext, ciphertext, icn)
}

func (m *mgm) ctrEncrypt(dst, src []byte, icn []byte) {
	ctr := make([]byte, m.blockSize)
	copy(ctr, icn)
	out := make([]byte, m.blockSize)

	for i := 0; i < len(src); i += m.blockSize {
		// Increment counter
		m.incr(ctr)
		m.cipher.Encrypt(out, ctr)

		end := i + m.blockSize
		if end > len(src) {
			end = len(src)
		}
		for j := i; j < end; j++ {
			dst[j] = src[j] ^ out[j-i]
		}
	}
}

func (m *mgm) incr(ctr []byte) {
	// Increment as big-endian integer
	for i := len(ctr) - 1; i >= 0; i-- {
		ctr[i]++
		if ctr[i] != 0 {
			break
		}
	}
}

func (m *mgm) computeTag(tag, h, ad, ct []byte) {
	// Pad and authenticate AD
	sum := make([]byte, m.blockSize)
	z := make([]byte, m.blockSize)
	copy(z, h)
	enc := make([]byte, m.blockSize)

	// Process additional data
	for i := 0; i < len(ad); i += m.blockSize {
		m.incr(z)
		m.cipher.Encrypt(enc, z)

		end := i + m.blockSize
		if end > len(ad) {
			// Pad last block
			block := make([]byte, m.blockSize)
			copy(block, ad[i:])
			m.gfMul(enc, block)
		} else {
			m.gfMul(enc, ad[i:end])
		}
		xorBytes(sum, sum, enc)
	}

	// Process ciphertext
	for i := 0; i < len(ct); i += m.blockSize {
		m.incr(z)
		m.cipher.Encrypt(enc, z)

		end := i + m.blockSize
		if end > len(ct) {
			block := make([]byte, m.blockSize)
			copy(block, ct[i:])
			m.gfMul(enc, block)
		} else {
			m.gfMul(enc, ct[i:end])
		}
		xorBytes(sum, sum, enc)
	}

	// Add length block: len(A) || len(C) in bits
	lenBlock := make([]byte, m.blockSize)
	if m.blockSize == 16 {
		binary.BigEndian.PutUint64(lenBlock[0:8], uint64(len(ad)*8))
		binary.BigEndian.PutUint64(lenBlock[8:16], uint64(len(ct)*8))
	} else {
		binary.BigEndian.PutUint32(lenBlock[0:4], uint32(len(ad)*8))
		binary.BigEndian.PutUint32(lenBlock[4:8], uint32(len(ct)*8))
	}
	m.incr(z)
	m.cipher.Encrypt(enc, z)
	m.gfMul(enc, lenBlock)
	xorBytes(sum, sum, enc)

	// Final encryption
	m.cipher.Encrypt(tag, sum)
}

// gfMul multiplies a and b in GF(2^n) and stores result in a.
func (m *mgm) gfMul(a, b []byte) {
	if m.blockSize == 16 {
		gfMul128(a, b)
	} else {
		gfMul64(a, b)
	}
}

// gfMul128 multiplies in GF(2^128) with polynomial x^128 + x^7 + x^2 + x + 1
func gfMul128(a, b []byte) {
	var z [16]byte
	var v [16]byte
	copy(v[:], a)

	for i := 0; i < 16; i++ {
		for j := 7; j >= 0; j-- {
			if (b[i] >> j & 1) == 1 {
				xorBytes(z[:], z[:], v[:])
			}
			// v = v * x
			carry := v[15] & 1
			for k := 15; k > 0; k-- {
				v[k] = (v[k] >> 1) | (v[k-1] << 7)
			}
			v[0] >>= 1
			if carry == 1 {
				v[0] ^= 0xE1 // Reduction polynomial
			}
		}
	}
	copy(a, z[:])
}

// gfMul64 multiplies in GF(2^64) with polynomial x^64 + x^4 + x^3 + x + 1
func gfMul64(a, b []byte) {
	var z [8]byte
	var v [8]byte
	copy(v[:], a)

	for i := 0; i < 8; i++ {
		for j := 7; j >= 0; j-- {
			if (b[i] >> j & 1) == 1 {
				xorBytes(z[:], z[:], v[:])
			}
			carry := v[7] & 1
			for k := 7; k > 0; k-- {
				v[k] = (v[k] >> 1) | (v[k-1] << 7)
			}
			v[0] >>= 1
			if carry == 1 {
				v[0] ^= 0x1B // Reduction polynomial
			}
		}
	}
	copy(a, z[:])
}

func xorBytes(dst, a, b []byte) {
	for i := range dst {
		dst[i] = a[i] ^ b[i]
	}
}

func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}
