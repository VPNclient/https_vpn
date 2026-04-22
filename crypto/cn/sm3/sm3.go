// Package sm3 implements the SM3 cryptographic hash algorithm per GB/T 32905-2016.
package sm3

import (
	"encoding/binary"
	"hash"
)

const (
	// Size is the size of an SM3 checksum in bytes.
	Size = 32
	// BlockSize is the block size of SM3 in bytes.
	BlockSize = 64
)

// digest represents the partial evaluation of a checksum.
type digest struct {
	h   [8]uint32
	x   [BlockSize]byte
	nx  int
	len uint64
}

// New returns a new hash.Hash computing the SM3 checksum.
func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

// Sum returns the SM3 checksum of the data.
func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}

func (d *digest) Reset() {
	d.h[0] = 0x7380166F
	d.h[1] = 0x4914B2B9
	d.h[2] = 0x172442D7
	d.h[3] = 0xDA8A0600
	d.h[4] = 0xA96F30BC
	d.h[5] = 0x163138AA
	d.h[6] = 0xE38DEE4D
	d.h[7] = 0xB0FB0E4E
	d.nx = 0
	d.len = 0
}

func (d *digest) Size() int      { return Size }
func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)

	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == BlockSize {
			d.block(d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}

	for len(p) >= BlockSize {
		d.block(p[:BlockSize])
		p = p[BlockSize:]
	}

	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}

	return
}

func (d *digest) Sum(in []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	len := d.len
	var tmp [64]byte
	tmp[0] = 0x80

	if len%64 < 56 {
		d.Write(tmp[0 : 56-len%64])
	} else {
		d.Write(tmp[0 : 64+56-len%64])
	}

	len <<= 3
	binary.BigEndian.PutUint64(tmp[:], len)
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte
	binary.BigEndian.PutUint32(digest[0:], d.h[0])
	binary.BigEndian.PutUint32(digest[4:], d.h[1])
	binary.BigEndian.PutUint32(digest[8:], d.h[2])
	binary.BigEndian.PutUint32(digest[12:], d.h[3])
	binary.BigEndian.PutUint32(digest[16:], d.h[4])
	binary.BigEndian.PutUint32(digest[20:], d.h[5])
	binary.BigEndian.PutUint32(digest[24:], d.h[6])
	binary.BigEndian.PutUint32(digest[28:], d.h[7])

	return digest
}

func (dig *digest) block(p []byte) {
	var w [68]uint32
	var w1 [64]uint32

	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(p[i*4:])
	}

	for i := 16; i < 68; i++ {
		w[i] = p1(w[i-16]^w[i-9]^leftRotate(w[i-3], 15)) ^ leftRotate(w[i-13], 7) ^ w[i-6]
	}

	for i := 0; i < 64; i++ {
		w1[i] = w[i] ^ w[i+4]
	}

	a, b, c, d, e, f, g, h := dig.h[0], dig.h[1], dig.h[2], dig.h[3], dig.h[4], dig.h[5], dig.h[6], dig.h[7]

	for i := 0; i < 16; i++ {
		ss1 := leftRotate(leftRotate(a, 12)+e+leftRotate(0x79CC4519, i%32), 7)
		ss2 := ss1 ^ leftRotate(a, 12)
		tt1 := ff0(a, b, c) + d + ss2 + w1[i]
		tt2 := gg0(e, f, g) + h + ss1 + w[i]
		d = c
		c = leftRotate(b, 9)
		b = a
		a = tt1
		h = g
		g = leftRotate(f, 19)
		f = e
		e = p0(tt2)
	}

	for i := 16; i < 64; i++ {
		ss1 := leftRotate(leftRotate(a, 12)+e+leftRotate(0x7A879D8A, i%32), 7)
		ss2 := ss1 ^ leftRotate(a, 12)
		tt1 := ff1(a, b, c) + d + ss2 + w1[i]
		tt2 := gg1(e, f, g) + h + ss1 + w[i]
		d = c
		c = leftRotate(b, 9)
		b = a
		a = tt1
		h = g
		g = leftRotate(f, 19)
		f = e
		e = p0(tt2)
	}

	dig.h[0] ^= a
	dig.h[1] ^= b
	dig.h[2] ^= c
	dig.h[3] ^= d
	dig.h[4] ^= e
	dig.h[5] ^= f
	dig.h[6] ^= g
	dig.h[7] ^= h
}

func leftRotate(x uint32, n int) uint32 {
	return (x << n) | (x >> (32 - n))
}

// Boolean functions
func ff0(x, y, z uint32) uint32 { return x ^ y ^ z }
func ff1(x, y, z uint32) uint32 { return (x & y) | (x & z) | (y & z) }
func gg0(x, y, z uint32) uint32 { return x ^ y ^ z }
func gg1(x, y, z uint32) uint32 { return (x & y) | (^x & z) }

// Permutation functions
func p0(x uint32) uint32 { return x ^ leftRotate(x, 9) ^ leftRotate(x, 17) }
func p1(x uint32) uint32 { return x ^ leftRotate(x, 15) ^ leftRotate(x, 23) }
