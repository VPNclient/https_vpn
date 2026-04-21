package gost

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Test vectors from GOST R 34.12-2015 Appendix A.
func TestKuznyechik(t *testing.T) {
	// Key from standard
	key, _ := hex.DecodeString("8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef")
	// Plaintext from standard
	plaintext, _ := hex.DecodeString("1122334455667700ffeeddccbbaa9988")
	// Expected ciphertext from standard
	expected, _ := hex.DecodeString("7f679d90bebc24305a468d42b9d4edcd")

	cipher, err := NewKuznyechik(key)
	if err != nil {
		t.Fatalf("NewKuznyechik failed: %v", err)
	}

	// Test encryption
	ciphertext := make([]byte, 16)
	cipher.Encrypt(ciphertext, plaintext)
	if !bytes.Equal(ciphertext, expected) {
		t.Errorf("Encrypt failed:\ngot:  %x\nwant: %x", ciphertext, expected)
	}

	// Test decryption
	decrypted := make([]byte, 16)
	cipher.Decrypt(decrypted, ciphertext)
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt failed:\ngot:  %x\nwant: %x", decrypted, plaintext)
	}
}

func TestKuznyechikBlockSize(t *testing.T) {
	key := make([]byte, 32)
	cipher, err := NewKuznyechik(key)
	if err != nil {
		t.Fatalf("NewKuznyechik failed: %v", err)
	}
	if cipher.BlockSize() != 16 {
		t.Errorf("BlockSize = %d, want 16", cipher.BlockSize())
	}
}

func TestKuznyechikInvalidKeySize(t *testing.T) {
	// Test various invalid key sizes
	for _, size := range []int{0, 15, 16, 31, 33, 64} {
		key := make([]byte, size)
		_, err := NewKuznyechik(key)
		if err == nil {
			t.Errorf("NewKuznyechik accepted invalid key size %d", size)
		}
	}
}

func TestKuznyechikSBox(t *testing.T) {
	// Verify S-box and inverse are consistent
	for i := 0; i < 256; i++ {
		s := kuzPi[i]
		inv := kuzPiInv[s]
		if byte(i) != inv {
			t.Errorf("S-box inconsistent at %d: Pi[%d]=%d, PiInv[%d]=%d", i, i, s, s, inv)
		}
	}
}

func TestKuznyechikLTransform(t *testing.T) {
	// Test that L and L^-1 are inverses
	original := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	block := original
	kuzL(&block)
	kuzLInv(&block)
	if block != original {
		t.Errorf("L transform not invertible:\ngot:  %x\nwant: %x", block, original)
	}
}

func TestGFMul(t *testing.T) {
	// Test GF multiplication properties
	// a * 1 = a
	for a := 0; a < 256; a++ {
		if kuzGFMul(byte(a), 1) != byte(a) {
			t.Errorf("gfMul(%d, 1) = %d, want %d", a, kuzGFMul(byte(a), 1), a)
		}
	}
	// a * 0 = 0
	for a := 0; a < 256; a++ {
		if kuzGFMul(byte(a), 0) != 0 {
			t.Errorf("gfMul(%d, 0) = %d, want 0", a, kuzGFMul(byte(a), 0))
		}
	}
}

func BenchmarkKuznyechikEncrypt(b *testing.B) {
	key := make([]byte, 32)
	cipher, _ := NewKuznyechik(key)
	plaintext := make([]byte, 16)
	ciphertext := make([]byte, 16)

	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkKuznyechikDecrypt(b *testing.B) {
	key := make([]byte, 32)
	cipher, _ := NewKuznyechik(key)
	plaintext := make([]byte, 16)
	ciphertext := make([]byte, 16)

	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Decrypt(plaintext, ciphertext)
	}
}
