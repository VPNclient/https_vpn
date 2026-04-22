package sm4

import (
	"crypto/cipher"
)

// NewGCM returns the SM4 cipher wrapped in Galois Counter Mode.
// This uses the standard 12-byte nonce and 16-byte tag.
func NewGCM(key []byte) (cipher.AEAD, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// NewGCMWithNonceSize returns SM4-GCM with custom nonce size.
func NewGCMWithNonceSize(key []byte, nonceSize int) (cipher.AEAD, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCMWithNonceSize(block, nonceSize)
}

// NewGCMWithTagSize returns SM4-GCM with custom tag size.
func NewGCMWithTagSize(key []byte, tagSize int) (cipher.AEAD, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCMWithTagSize(block, tagSize)
}
