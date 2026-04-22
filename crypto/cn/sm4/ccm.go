package sm4

import (
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"
	"errors"
)

// CCM implements Counter with CBC-MAC mode for SM4.
type ccm struct {
	cipher    cipher.Block
	nonceSize int
	tagSize   int
}

// NewCCM returns SM4 cipher wrapped in CCM mode.
// Standard nonce size is 12 bytes, tag size is 16 bytes.
func NewCCM(key []byte) (cipher.AEAD, error) {
	return NewCCMWithNonceAndTagSize(key, 12, 16)
}

// NewCCMWithNonceAndTagSize returns SM4-CCM with custom nonce and tag sizes.
// Nonce size must be in [7, 13]. Tag size must be in {4, 6, 8, 10, 12, 14, 16}.
func NewCCMWithNonceAndTagSize(key []byte, nonceSize, tagSize int) (cipher.AEAD, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	if nonceSize < 7 || nonceSize > 13 {
		return nil, errors.New("sm4: invalid CCM nonce size")
	}

	if tagSize < 4 || tagSize > 16 || tagSize%2 != 0 {
		return nil, errors.New("sm4: invalid CCM tag size")
	}

	return &ccm{
		cipher:    block,
		nonceSize: nonceSize,
		tagSize:   tagSize,
	}, nil
}

func (c *ccm) NonceSize() int { return c.nonceSize }
func (c *ccm) Overhead() int  { return c.tagSize }

func (c *ccm) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	if len(nonce) != c.nonceSize {
		panic("sm4: incorrect nonce length")
	}

	ret, out := sliceForAppend(dst, len(plaintext)+c.tagSize)
	tag := c.auth(nonce, plaintext, additionalData)
	copy(out[len(plaintext):], tag[:c.tagSize])

	c.ctr(out[:len(plaintext)], plaintext, nonce)

	// Encrypt tag with first counter block
	var tagBlock [BlockSize]byte
	c.ctr(tagBlock[:c.tagSize], tag[:c.tagSize], nonce)
	copy(out[len(plaintext):], tagBlock[:c.tagSize])

	return ret
}

func (c *ccm) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if len(nonce) != c.nonceSize {
		return nil, errors.New("sm4: incorrect nonce length")
	}

	if len(ciphertext) < c.tagSize {
		return nil, errors.New("sm4: ciphertext too short")
	}

	plaintextLen := len(ciphertext) - c.tagSize
	ret, out := sliceForAppend(dst, plaintextLen)

	// Decrypt tag
	var tagBlock [BlockSize]byte
	encryptedTag := ciphertext[plaintextLen:]
	c.ctr(tagBlock[:c.tagSize], encryptedTag, nonce)

	// Decrypt plaintext
	c.ctr(out, ciphertext[:plaintextLen], nonce)

	// Compute expected tag
	expectedTag := c.auth(nonce, out, additionalData)

	if subtle.ConstantTimeCompare(tagBlock[:c.tagSize], expectedTag[:c.tagSize]) != 1 {
		// Clear output on auth failure
		for i := range out {
			out[i] = 0
		}
		return nil, errors.New("sm4: authentication failed")
	}

	return ret, nil
}

// auth computes CBC-MAC tag
func (c *ccm) auth(nonce, plaintext, additionalData []byte) [BlockSize]byte {
	var tag [BlockSize]byte
	var b [BlockSize]byte

	// B0: flags || nonce || message length
	q := 15 - c.nonceSize // Length field size
	b[0] = byte((c.tagSize-2)/2)<<3 | byte(q-1)
	if len(additionalData) > 0 {
		b[0] |= 0x40
	}
	copy(b[1:1+c.nonceSize], nonce)

	// Encode message length in last q bytes
	msgLen := uint64(len(plaintext))
	for i := 0; i < q; i++ {
		b[15-i] = byte(msgLen >> (8 * i))
	}

	c.cipher.Encrypt(tag[:], b[:])

	// Process additional data
	if len(additionalData) > 0 {
		c.authAdditionalData(&tag, additionalData)
	}

	// Process plaintext
	c.authBlocks(&tag, plaintext)

	return tag
}

func (c *ccm) authAdditionalData(tag *[BlockSize]byte, data []byte) {
	var b [BlockSize]byte
	var lenBytes int

	// Encode length
	if len(data) < 0xFF00 {
		b[0] = byte(len(data) >> 8)
		b[1] = byte(len(data))
		lenBytes = 2
	} else {
		b[0] = 0xFF
		b[1] = 0xFE
		binary.BigEndian.PutUint32(b[2:], uint32(len(data)))
		lenBytes = 6
	}

	// XOR length and first part of data
	n := copy(b[lenBytes:], data)
	for i := 0; i < BlockSize; i++ {
		tag[i] ^= b[i]
	}
	c.cipher.Encrypt(tag[:], tag[:])
	data = data[n:]

	// Process remaining data
	c.authBlocks(tag, data)
}

func (c *ccm) authBlocks(tag *[BlockSize]byte, data []byte) {
	for len(data) >= BlockSize {
		for i := 0; i < BlockSize; i++ {
			tag[i] ^= data[i]
		}
		c.cipher.Encrypt(tag[:], tag[:])
		data = data[BlockSize:]
	}

	if len(data) > 0 {
		for i := 0; i < len(data); i++ {
			tag[i] ^= data[i]
		}
		c.cipher.Encrypt(tag[:], tag[:])
	}
}

// ctr performs CTR encryption/decryption
func (c *ccm) ctr(dst, src []byte, nonce []byte) {
	var counter [BlockSize]byte
	var block [BlockSize]byte

	q := 15 - c.nonceSize
	counter[0] = byte(q - 1)
	copy(counter[1:1+c.nonceSize], nonce)

	// Start from counter 1 for data (counter 0 is for tag)
	counter[15] = 1

	for len(src) > 0 {
		c.cipher.Encrypt(block[:], counter[:])

		n := len(src)
		if n > BlockSize {
			n = BlockSize
		}

		for i := 0; i < n; i++ {
			dst[i] = src[i] ^ block[i]
		}

		src = src[n:]
		dst = dst[n:]

		// Increment counter
		for i := 15; i >= 1+c.nonceSize; i-- {
			counter[i]++
			if counter[i] != 0 {
				break
			}
		}
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
