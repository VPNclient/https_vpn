// CTR mode for GOST block ciphers per GOST R 34.13-2015.
package gost

import (
	"crypto/cipher"
)

// ctr implements CTR mode encryption/decryption.
type ctr struct {
	b       cipher.Block
	ctr     []byte
	out     []byte
	outUsed int
}

// NewCTR returns a Stream that encrypts/decrypts using the given Block in CTR mode.
// The IV must be the same length as the Block's block size.
func NewCTR(block cipher.Block, iv []byte) cipher.Stream {
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		panic("gost/ctr: IV length must equal block size")
	}
	return &ctr{
		b:       block,
		ctr:     append([]byte(nil), iv...),
		out:     make([]byte, blockSize),
		outUsed: blockSize, // force generation on first use
	}
}

func (c *ctr) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("gost/ctr: output smaller than input")
	}
	for i := 0; i < len(src); i++ {
		if c.outUsed >= len(c.out) {
			c.b.Encrypt(c.out, c.ctr)
			c.outUsed = 0
			// Increment counter (big-endian)
			for j := len(c.ctr) - 1; j >= 0; j-- {
				c.ctr[j]++
				if c.ctr[j] != 0 {
					break
				}
			}
		}
		dst[i] = src[i] ^ c.out[c.outUsed]
		c.outUsed++
	}
}
