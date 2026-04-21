// Magma block cipher per GOST R 34.12-2015.
// 64-bit block, 256-bit key, 32 rounds.
package gost

import (
	"crypto/cipher"
	"encoding/binary"
)

const (
	MagmaBlockSize = 8  // 64 bits
	MagmaKeySize   = 32 // 256 bits
)

// magmaCipher implements the Magma block cipher.
type magmaCipher struct {
	k [8]uint32 // 8 subkeys (32-bit each)
}

// NewMagma creates a new Magma cipher with the given key.
// Key must be exactly 32 bytes (256 bits).
func NewMagma(key []byte) (cipher.Block, error) {
	if len(key) != MagmaKeySize {
		return nil, KeySizeError(len(key))
	}
	c := new(magmaCipher)
	// Split key into 8 x 32-bit subkeys (big-endian per GOST R 34.12-2015)
	for i := 0; i < 8; i++ {
		c.k[i] = binary.BigEndian.Uint32(key[i*4 : i*4+4])
	}
	return c, nil
}

func (c *magmaCipher) BlockSize() int {
	return MagmaBlockSize
}

func (c *magmaCipher) Encrypt(dst, src []byte) {
	if len(src) < MagmaBlockSize || len(dst) < MagmaBlockSize {
		panic("gost: input/output too short")
	}
	// Split block into two 32-bit halves (big-endian per standard)
	a := binary.BigEndian.Uint32(src[0:4])
	b := binary.BigEndian.Uint32(src[4:8])

	// 32 rounds with key schedule: K0..K7 repeated 3 times, then K7..K0
	for i := 0; i < 24; i++ {
		a, b = b, a^magmaG(b+c.k[i%8])
	}
	for i := 7; i >= 0; i-- {
		a, b = b, a^magmaG(b+c.k[i])
	}

	// Output (note: halves are swapped at the end)
	binary.BigEndian.PutUint32(dst[0:4], b)
	binary.BigEndian.PutUint32(dst[4:8], a)
}

func (c *magmaCipher) Decrypt(dst, src []byte) {
	if len(src) < MagmaBlockSize || len(dst) < MagmaBlockSize {
		panic("gost: input/output too short")
	}
	// Split block (big-endian)
	a := binary.BigEndian.Uint32(src[0:4])
	b := binary.BigEndian.Uint32(src[4:8])

	// Reverse key schedule: K0..K7, then K7..K0 repeated 3 times
	for i := 0; i < 8; i++ {
		a, b = b, a^magmaG(b+c.k[i])
	}
	for i := 23; i >= 0; i-- {
		a, b = b, a^magmaG(b+c.k[i%8])
	}

	binary.BigEndian.PutUint32(dst[0:4], b)
	binary.BigEndian.PutUint32(dst[4:8], a)
}

// magmaG is the round function: S-box substitution + left rotation by 11
func magmaG(x uint32) uint32 {
	// Apply 4-bit S-boxes to each nibble
	var y uint32
	for i := 0; i < 8; i++ {
		nibble := (x >> (4 * i)) & 0xF
		y |= uint32(magmaSBox[i][nibble]) << (4 * i)
	}
	// Rotate left by 11 bits
	return (y << 11) | (y >> 21)
}

// magmaSBox contains the S-boxes from GOST R 34.12-2015 (id-tc26-gost-28147-param-Z)
var magmaSBox = [8][16]byte{
	{0xC, 0x4, 0x6, 0x2, 0xA, 0x5, 0xB, 0x9, 0xE, 0x8, 0xD, 0x7, 0x0, 0x3, 0xF, 0x1},
	{0x6, 0x8, 0x2, 0x3, 0x9, 0xA, 0x5, 0xC, 0x1, 0xE, 0x4, 0x7, 0xB, 0xD, 0x0, 0xF},
	{0xB, 0x3, 0x5, 0x8, 0x2, 0xF, 0xA, 0xD, 0xE, 0x1, 0x7, 0x4, 0xC, 0x9, 0x6, 0x0},
	{0xC, 0x8, 0x2, 0x1, 0xD, 0x4, 0xF, 0x6, 0x7, 0x0, 0xA, 0x5, 0x3, 0xE, 0x9, 0xB},
	{0x7, 0xF, 0x5, 0xA, 0x8, 0x1, 0x6, 0xD, 0x0, 0x9, 0x3, 0xE, 0xB, 0x4, 0x2, 0xC},
	{0x5, 0xD, 0xF, 0x6, 0x9, 0x2, 0xC, 0xA, 0xB, 0x7, 0x8, 0x1, 0x4, 0x3, 0xE, 0x0},
	{0x8, 0xE, 0x2, 0x5, 0x6, 0x9, 0x1, 0xC, 0xF, 0x4, 0xB, 0x0, 0xD, 0xA, 0x3, 0x7},
	{0x1, 0x7, 0xE, 0xD, 0x0, 0x5, 0x8, 0x3, 0x4, 0xF, 0xA, 0x6, 0x9, 0xC, 0xB, 0x2},
}
