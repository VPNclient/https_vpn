// Streebog hash function per GOST R 34.11-2012.
// Supports both 256-bit and 512-bit output.
package gost

import (
	"encoding/binary"
	"hash"
)

const (
	streebogBlockSize = 64 // 512 bits
	streebog256Size   = 32 // 256-bit output
	streebog512Size   = 64 // 512-bit output
)

// streebog implements hash.Hash for Streebog.
type streebog struct {
	h      [64]byte // hash state
	n      [64]byte // message length counter
	sigma  [64]byte // checksum
	buf    [64]byte // partial block buffer
	bufLen int
	size   int // output size (32 or 64)
}

// NewStreebog256 returns a new Streebog-256 hash.
func NewStreebog256() hash.Hash {
	d := &streebog{size: streebog256Size}
	d.Reset()
	return d
}

// NewStreebog512 returns a new Streebog-512 hash.
func NewStreebog512() hash.Hash {
	d := &streebog{size: streebog512Size}
	d.Reset()
	return d
}

func (d *streebog) Reset() {
	d.n = [64]byte{}
	d.sigma = [64]byte{}
	d.buf = [64]byte{}
	d.bufLen = 0
	if d.size == streebog256Size {
		// IV for 256-bit: all 0x01
		for i := range d.h {
			d.h[i] = 0x01
		}
	} else {
		// IV for 512-bit: all zeros
		d.h = [64]byte{}
	}
}

func (d *streebog) Size() int {
	return d.size
}

func (d *streebog) BlockSize() int {
	return streebogBlockSize
}

func (d *streebog) Write(p []byte) (n int, err error) {
	n = len(p)
	// Fill buffer
	if d.bufLen > 0 {
		toFill := streebogBlockSize - d.bufLen
		if toFill > len(p) {
			toFill = len(p)
		}
		copy(d.buf[d.bufLen:], p[:toFill])
		d.bufLen += toFill
		p = p[toFill:]
		if d.bufLen == streebogBlockSize {
			d.processBlock(d.buf[:])
			d.bufLen = 0
		}
	}
	// Process complete blocks
	for len(p) >= streebogBlockSize {
		d.processBlock(p[:streebogBlockSize])
		p = p[streebogBlockSize:]
	}
	// Save remaining
	if len(p) > 0 {
		copy(d.buf[:], p)
		d.bufLen = len(p)
	}
	return
}

func (d *streebog) Sum(in []byte) []byte {
	d0 := *d // Copy state
	hash := d0.finalize()
	return append(in, hash[:d.size]...)
}

func (d *streebog) finalize() [64]byte {
	// Pad message
	var padded [64]byte
	copy(padded[:], d.buf[:d.bufLen])
	padded[d.bufLen] = 0x01 // Padding: 1 || 0*

	// Process padded block with message length
	d.gN(padded[:], d.n[:])

	// Update N with remaining bits
	addMod512(&d.n, uint64(d.bufLen)*8)

	// Add padded block to sigma
	addBlocks(&d.sigma, padded[:])

	// Final compression: h = g_0(h, N)
	var zero [64]byte
	d.gN(d.n[:], zero[:])

	// h = g_0(h, sigma)
	d.gN(d.sigma[:], zero[:])

	return d.h
}

func (d *streebog) processBlock(block []byte) {
	d.gN(block, d.n[:])
	addMod512(&d.n, 512)
	addBlocks(&d.sigma, block)
}

// gN is the compression function: h = g_N(h, m)
func (d *streebog) gN(m []byte, n []byte) {
	var k, tmp [64]byte
	// K = h XOR N
	for i := 0; i < 64; i++ {
		k[i] = d.h[i] ^ n[i]
	}
	// E(K, m)
	copy(tmp[:], m)
	for i := 0; i < 12; i++ {
		streebogS(&tmp)
		streebogP(&tmp)
		streebogL(&tmp)
		xorBlock(&tmp, k[:])
		// Update K for next round
		xorBlock(&k, streebogC[i][:])
		streebogS(&k)
		streebogP(&k)
		streebogL(&k)
	}
	// h = h XOR tmp XOR m
	for i := 0; i < 64; i++ {
		d.h[i] ^= tmp[i] ^ m[i]
	}
}

// S-box substitution (applied to each byte)
func streebogS(block *[64]byte) {
	for i := 0; i < 64; i++ {
		block[i] = streebogPi[block[i]]
	}
}

// P transformation (byte permutation)
func streebogP(block *[64]byte) {
	var tmp [64]byte
	for i := 0; i < 64; i++ {
		tmp[streebogTau[i]] = block[i]
	}
	*block = tmp
}

// L transformation (linear, operates on 8-byte chunks)
func streebogL(block *[64]byte) {
	for i := 0; i < 8; i++ {
		var v uint64
		for j := 0; j < 8; j++ {
			b := block[i*8+j]
			for k := 0; k < 8; k++ {
				if (b >> k & 1) == 1 {
					v ^= streebogA[j*8+k]
				}
			}
		}
		binary.LittleEndian.PutUint64(block[i*8:], v)
	}
}

func xorBlock(a *[64]byte, b []byte) {
	for i := 0; i < 64; i++ {
		a[i] ^= b[i]
	}
}

func addBlocks(a *[64]byte, b []byte) {
	var carry uint16
	for i := 63; i >= 0; i-- {
		carry += uint16(a[i]) + uint16(b[i])
		a[i] = byte(carry)
		carry >>= 8
	}
}

func addMod512(a *[64]byte, val uint64) {
	var carry uint64 = val
	for i := 63; i >= 0 && carry > 0; i-- {
		carry += uint64(a[i])
		a[i] = byte(carry)
		carry >>= 8
	}
}

// S-box Pi from GOST R 34.11-2012
var streebogPi = [256]byte{
	0xFC, 0xEE, 0xDD, 0x11, 0xCF, 0x6E, 0x31, 0x16, 0xFB, 0xC4, 0xFA, 0xDA, 0x23, 0xC5, 0x04, 0x4D,
	0xE9, 0x77, 0xF0, 0xDB, 0x93, 0x2E, 0x99, 0xBA, 0x17, 0x36, 0xF1, 0xBB, 0x14, 0xCD, 0x5F, 0xC1,
	0xF9, 0x18, 0x65, 0x5A, 0xE2, 0x5C, 0xEF, 0x21, 0x81, 0x1C, 0x3C, 0x42, 0x8B, 0x01, 0x8E, 0x4F,
	0x05, 0x84, 0x02, 0xAE, 0xE3, 0x6A, 0x8F, 0xA0, 0x06, 0x0B, 0xED, 0x98, 0x7F, 0xD4, 0xD3, 0x1F,
	0xEB, 0x34, 0x2C, 0x51, 0xEA, 0xC8, 0x48, 0xAB, 0xF2, 0x2A, 0x68, 0xA2, 0xFD, 0x3A, 0xCE, 0xCC,
	0xB5, 0x70, 0x0E, 0x56, 0x08, 0x0C, 0x76, 0x12, 0xBF, 0x72, 0x13, 0x47, 0x9C, 0xB7, 0x5D, 0x87,
	0x15, 0xA1, 0x96, 0x29, 0x10, 0x7B, 0x9A, 0xC7, 0xF3, 0x91, 0x78, 0x6F, 0x9D, 0x9E, 0xB2, 0xB1,
	0x32, 0x75, 0x19, 0x3D, 0xFF, 0x35, 0x8A, 0x7E, 0x6D, 0x54, 0xC6, 0x80, 0xC3, 0xBD, 0x0D, 0x57,
	0xDF, 0xF5, 0x24, 0xA9, 0x3E, 0xA8, 0x43, 0xC9, 0xD7, 0x79, 0xD6, 0xF6, 0x7C, 0x22, 0xB9, 0x03,
	0xE0, 0x0F, 0xEC, 0xDE, 0x7A, 0x94, 0xB0, 0xBC, 0xDC, 0xE8, 0x28, 0x50, 0x4E, 0x33, 0x0A, 0x4A,
	0xA7, 0x97, 0x60, 0x73, 0x1E, 0x00, 0x62, 0x44, 0x1A, 0xB8, 0x38, 0x82, 0x64, 0x9F, 0x26, 0x41,
	0xAD, 0x45, 0x46, 0x92, 0x27, 0x5E, 0x55, 0x2F, 0x8C, 0xA3, 0xA5, 0x7D, 0x69, 0xD5, 0x95, 0x3B,
	0x07, 0x58, 0xB3, 0x40, 0x86, 0xAC, 0x1D, 0xF7, 0x30, 0x37, 0x6B, 0xE4, 0x88, 0xD9, 0xE7, 0x89,
	0xE1, 0x1B, 0x83, 0x49, 0x4C, 0x3F, 0xF8, 0xFE, 0x8D, 0x53, 0xAA, 0x90, 0xCA, 0xD8, 0x85, 0x61,
	0x20, 0x71, 0x67, 0xA4, 0x2D, 0x2B, 0x09, 0x5B, 0xCB, 0x9B, 0x25, 0xD0, 0xBE, 0xE5, 0x6C, 0x52,
	0x59, 0xA6, 0x74, 0xD2, 0xE6, 0xF4, 0xB4, 0xC0, 0xD1, 0x66, 0xAF, 0xC2, 0x39, 0x4B, 0x63, 0xB6,
}

// Tau permutation for P transformation
var streebogTau = [64]byte{
	0, 8, 16, 24, 32, 40, 48, 56,
	1, 9, 17, 25, 33, 41, 49, 57,
	2, 10, 18, 26, 34, 42, 50, 58,
	3, 11, 19, 27, 35, 43, 51, 59,
	4, 12, 20, 28, 36, 44, 52, 60,
	5, 13, 21, 29, 37, 45, 53, 61,
	6, 14, 22, 30, 38, 46, 54, 62,
	7, 15, 23, 31, 39, 47, 55, 63,
}

// A matrix coefficients for L transformation (64 values)
var streebogA = [64]uint64{
	0x8e20faa72ba0b470, 0x47107ddd9b505a38, 0xad08b0e0c3282d1c, 0xd8045870ef14980e,
	0x6c022c38f90a4c07, 0x3601161cf205268d, 0x1b8e0b0e798c13c8, 0x83478b07b2468764,
	0xa011d380818e8f40, 0x5086e740ce47c920, 0x2843fd2067adea10, 0x14aff010bdd87508,
	0x0ad97808d06cb404, 0x05e23c0468365a02, 0x8c711e02341b2d01, 0x46b60f011a83988e,
	0x90dab52a387ae76f, 0x486dd4151c3dfdb9, 0x24b86a840e90f0d2, 0x125c354207487869,
	0x092e94218d243cba, 0x8a174a9ec8121e5d, 0x4585254f64090fa0, 0xaccc9ca9328a8950,
	0x9d4df05d5f661451, 0xc0a878a0a1330aa6, 0x60543c50de970553, 0x302a1e286fc58ca7,
	0x18150f14b9ec46dd, 0x0c84890ad27623e0, 0x0642ca05693b9f70, 0x0321658cba93c138,
	0x86275df09ce8aaa8, 0x439da0784e745554, 0xafc0503c273aa42a, 0xd960281e9d1d5215,
	0xe230140fc0802984, 0x71180a8960409a42, 0xb60c05ca30204d21, 0x5b068c651810a89e,
	0x456c34887a3805b9, 0xac361a443d1c8cd2, 0x561b0d22900e4669, 0x2b838811480723ba,
	0x9bcf4486248d9f5d, 0xc3e9224312c8c1a0, 0xeffa11af0964ee50, 0xf97d86d98a327728,
	0xe4fa2054a80b329c, 0x727d102a548b194e, 0x39b008152acb8227, 0x9258048415eb419d,
	0x492c024284fbaec0, 0xaa16012142f35760, 0x550b8e9e21f7a530, 0xa48b474f9ef5dc18,
	0x70a6a56e2440598e, 0x3853dc371220a247, 0x1ca76e95091051ad, 0x0edd37c48a08a6d8,
	0x07e095624504536c, 0x8d70c431ac02a736, 0xc83862965601dd1b, 0x641c314b2b8ee083,
}

// Round constants C for key schedule
var streebogC [12][64]byte

func init() {
	// Generate round constants
	for i := 0; i < 12; i++ {
		var c [64]byte
		for j := 0; j < 64; j++ {
			c[j] = byte(i*64 + j)
		}
		streebogS(&c)
		streebogP(&c)
		streebogL(&c)
		streebogC[i] = c
	}
}
