// OMAC (One-key MAC, also known as CMAC) for GOST block ciphers.
// Per GOST R 34.13-2015.
package gost

import (
	"crypto/cipher"
	"hash"
)

// omac implements hash.Hash for OMAC/CMAC.
type omac struct {
	cipher cipher.Block
	k1, k2 []byte // Subkeys
	x      []byte // Running state
	buf    []byte // Partial block buffer
	bufLen int
	size   int // Block size
}

// NewOMAC creates a new OMAC with the given block cipher.
// Works with Kuznyechik (128-bit) or Magma (64-bit).
func NewOMAC(block cipher.Block) hash.Hash {
	blockSize := block.BlockSize()
	o := &omac{
		cipher: block,
		k1:     make([]byte, blockSize),
		k2:     make([]byte, blockSize),
		x:      make([]byte, blockSize),
		buf:    make([]byte, blockSize),
		size:   blockSize,
	}
	o.deriveSubkeys()
	return o
}

// deriveSubkeys computes K1 and K2 from K.
func (o *omac) deriveSubkeys() {
	// L = E_K(0)
	L := make([]byte, o.size)
	o.cipher.Encrypt(L, L)

	// K1 = L << 1, with conditional XOR
	o.k1 = leftShift(L, o.size)
	if L[0]&0x80 != 0 {
		// XOR with Rb (reduction polynomial)
		if o.size == 16 {
			o.k1[15] ^= 0x87 // R128
		} else {
			o.k1[7] ^= 0x1B // R64
		}
	}

	// K2 = K1 << 1, with conditional XOR
	o.k2 = leftShift(o.k1, o.size)
	if o.k1[0]&0x80 != 0 {
		if o.size == 16 {
			o.k2[15] ^= 0x87
		} else {
			o.k2[7] ^= 0x1B
		}
	}
}

func leftShift(in []byte, size int) []byte {
	out := make([]byte, size)
	for i := 0; i < size-1; i++ {
		out[i] = (in[i] << 1) | (in[i+1] >> 7)
	}
	out[size-1] = in[size-1] << 1
	return out
}

func (o *omac) Reset() {
	for i := range o.x {
		o.x[i] = 0
	}
	for i := range o.buf {
		o.buf[i] = 0
	}
	o.bufLen = 0
}

func (o *omac) Size() int {
	return o.size
}

func (o *omac) BlockSize() int {
	return o.size
}

func (o *omac) Write(p []byte) (n int, err error) {
	n = len(p)

	// Fill buffer
	if o.bufLen > 0 {
		toFill := o.size - o.bufLen
		if toFill > len(p) {
			toFill = len(p)
		}
		copy(o.buf[o.bufLen:], p[:toFill])
		o.bufLen += toFill
		p = p[toFill:]

		if o.bufLen == o.size && len(p) > 0 {
			// Process complete block (not the last one)
			for i := 0; i < o.size; i++ {
				o.x[i] ^= o.buf[i]
			}
			o.cipher.Encrypt(o.x, o.x)
			o.bufLen = 0
		}
	}

	// Process complete blocks (except potentially the last)
	for len(p) > o.size {
		for i := 0; i < o.size; i++ {
			o.x[i] ^= p[i]
		}
		o.cipher.Encrypt(o.x, o.x)
		p = p[o.size:]
	}

	// Save remaining
	if len(p) > 0 {
		copy(o.buf[:], p)
		o.bufLen = len(p)
	}

	return
}

func (o *omac) Sum(in []byte) []byte {
	// Copy state for finalization
	x := make([]byte, o.size)
	copy(x, o.x)
	buf := make([]byte, o.size)
	copy(buf, o.buf[:o.bufLen])
	bufLen := o.bufLen

	// Finalize
	if bufLen == o.size {
		// Complete block: XOR with K1
		for i := 0; i < o.size; i++ {
			x[i] ^= buf[i] ^ o.k1[i]
		}
	} else {
		// Incomplete block: pad and XOR with K2
		buf[bufLen] = 0x80
		for i := bufLen + 1; i < o.size; i++ {
			buf[i] = 0
		}
		for i := 0; i < o.size; i++ {
			x[i] ^= buf[i] ^ o.k2[i]
		}
	}

	o.cipher.Encrypt(x, x)
	return append(in, x...)
}

// OMAC computes OMAC/CMAC of data with the given block cipher.
func OMAC(block cipher.Block, data []byte) []byte {
	h := NewOMAC(block)
	h.Write(data)
	return h.Sum(nil)
}
