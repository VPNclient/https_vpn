package sm4

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Test vectors from GB/T 32907-2016 Appendix A
func TestSM4(t *testing.T) {
	key, _ := hex.DecodeString("0123456789ABCDEFFEDCBA9876543210")
	plaintext, _ := hex.DecodeString("0123456789ABCDEFFEDCBA9876543210")
	expectedCipher, _ := hex.DecodeString("681EDF34D206965E86B3E94F536E4246")

	c, err := NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	// Test encryption
	ciphertext := make([]byte, BlockSize)
	c.Encrypt(ciphertext, plaintext)

	if !bytes.Equal(ciphertext, expectedCipher) {
		t.Errorf("Encrypt: got %x, want %x", ciphertext, expectedCipher)
	}

	// Test decryption
	decrypted := make([]byte, BlockSize)
	c.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt: got %x, want %x", decrypted, plaintext)
	}
}

// Test 1,000,000 iterations (from GB/T 32907-2016 Appendix A.2)
func TestSM4Million(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping million iteration test in short mode")
	}

	key, _ := hex.DecodeString("0123456789ABCDEFFEDCBA9876543210")
	plaintext, _ := hex.DecodeString("0123456789ABCDEFFEDCBA9876543210")
	expected, _ := hex.DecodeString("595298C7C6FD271F0402F804C33D3F66")

	c, err := NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	result := make([]byte, BlockSize)
	copy(result, plaintext)

	for i := 0; i < 1000000; i++ {
		c.Encrypt(result, result)
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("After 1M encryptions: got %x, want %x", result, expected)
	}
}

func TestSM4InvalidKeySize(t *testing.T) {
	_, err := NewCipher(make([]byte, 8))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestSM4BlockSize(t *testing.T) {
	key := make([]byte, KeySize)
	c, err := NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	if c.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, want %d", c.BlockSize(), BlockSize)
	}
}

func TestSM4EncryptDecryptRoundtrip(t *testing.T) {
	key, _ := hex.DecodeString("0123456789ABCDEFFEDCBA9876543210")

	c, err := NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	testCases := [][]byte{
		make([]byte, BlockSize),
		{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
	}

	for _, plaintext := range testCases {
		ciphertext := make([]byte, BlockSize)
		decrypted := make([]byte, BlockSize)

		c.Encrypt(ciphertext, plaintext)
		c.Decrypt(decrypted, ciphertext)

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Roundtrip failed for %x", plaintext)
		}
	}
}

func BenchmarkSM4Encrypt(b *testing.B) {
	key := make([]byte, KeySize)
	c, _ := NewCipher(key)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	b.SetBytes(BlockSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkSM4Decrypt(b *testing.B) {
	key := make([]byte, KeySize)
	c, _ := NewCipher(key)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	b.SetBytes(BlockSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Decrypt(plaintext, ciphertext)
	}
}
