package gost

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Test vectors from GOST R 34.12-2015 Appendix B.
func TestMagma(t *testing.T) {
	// Key from standard (same as Kuznyechik test)
	key, _ := hex.DecodeString("ffeeddccbbaa99887766554433221100f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff")
	// Plaintext from standard
	plaintext, _ := hex.DecodeString("fedcba9876543210")
	// Expected ciphertext from standard
	expected, _ := hex.DecodeString("4ee901e5c2d8ca3d")

	cipher, err := NewMagma(key)
	if err != nil {
		t.Fatalf("NewMagma failed: %v", err)
	}

	// Test encryption
	ciphertext := make([]byte, 8)
	cipher.Encrypt(ciphertext, plaintext)
	if !bytes.Equal(ciphertext, expected) {
		t.Errorf("Encrypt failed:\ngot:  %x\nwant: %x", ciphertext, expected)
	}

	// Test decryption
	decrypted := make([]byte, 8)
	cipher.Decrypt(decrypted, ciphertext)
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt failed:\ngot:  %x\nwant: %x", decrypted, plaintext)
	}
}

func TestMagmaBlockSize(t *testing.T) {
	key := make([]byte, 32)
	cipher, err := NewMagma(key)
	if err != nil {
		t.Fatalf("NewMagma failed: %v", err)
	}
	if cipher.BlockSize() != 8 {
		t.Errorf("BlockSize = %d, want 8", cipher.BlockSize())
	}
}

func TestMagmaInvalidKeySize(t *testing.T) {
	for _, size := range []int{0, 15, 16, 31, 33, 64} {
		key := make([]byte, size)
		_, err := NewMagma(key)
		if err == nil {
			t.Errorf("NewMagma accepted invalid key size %d", size)
		}
	}
}

func TestMagmaRoundTrip(t *testing.T) {
	// Test with random-ish data
	key := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	plaintext := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}

	cipher, _ := NewMagma(key)
	ciphertext := make([]byte, 8)
	decrypted := make([]byte, 8)

	cipher.Encrypt(ciphertext, plaintext)
	cipher.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Round trip failed:\noriginal:  %x\ndecrypted: %x", plaintext, decrypted)
	}
}

func BenchmarkMagmaEncrypt(b *testing.B) {
	key := make([]byte, 32)
	cipher, _ := NewMagma(key)
	plaintext := make([]byte, 8)
	ciphertext := make([]byte, 8)

	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkMagmaDecrypt(b *testing.B) {
	key := make([]byte, 32)
	cipher, _ := NewMagma(key)
	plaintext := make([]byte, 8)
	ciphertext := make([]byte, 8)

	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Decrypt(plaintext, ciphertext)
	}
}
