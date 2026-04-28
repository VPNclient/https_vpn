package aria

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestARIA128(t *testing.T) {
	// Test vector from RFC 5794
	key, _ := hex.DecodeString("00112233445566778899aabbccddeeff")
	plaintext, _ := hex.DecodeString("11111111aaaaaaaa11111111bbbbbbbb")
	expected, _ := hex.DecodeString("c6ecd08e22c30abdb215cf74e2075e6e")

	cipher, err := NewCipher128(key)
	if err != nil {
		t.Fatalf("NewCipher128 failed: %v", err)
	}

	ciphertext := make([]byte, BlockSize)
	cipher.Encrypt(ciphertext, plaintext)

	if !bytes.Equal(ciphertext, expected) {
		t.Errorf("ARIA-128 encryption failed\nExpected: %x\nGot:      %x", expected, ciphertext)
	}

	// Test decryption
	decrypted := make([]byte, BlockSize)
	cipher.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("ARIA-128 decryption failed\nExpected: %x\nGot:      %x", plaintext, decrypted)
	}
}

func TestARIA192(t *testing.T) {
	// Test vector from RFC 5794
	key, _ := hex.DecodeString("00112233445566778899aabbccddeeff0011223344556677")
	plaintext, _ := hex.DecodeString("11111111aaaaaaaa11111111bbbbbbbb")
	expected, _ := hex.DecodeString("8d1470625f59ebacb0e55b534b3e462b")

	cipher, err := NewCipher192(key)
	if err != nil {
		t.Fatalf("NewCipher192 failed: %v", err)
	}

	ciphertext := make([]byte, BlockSize)
	cipher.Encrypt(ciphertext, plaintext)

	if !bytes.Equal(ciphertext, expected) {
		t.Errorf("ARIA-192 encryption failed\nExpected: %x\nGot:      %x", expected, ciphertext)
	}

	// Test decryption
	decrypted := make([]byte, BlockSize)
	cipher.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("ARIA-192 decryption failed\nExpected: %x\nGot:      %x", plaintext, decrypted)
	}
}

func TestARIA256(t *testing.T) {
	// Test vector from RFC 5794
	key, _ := hex.DecodeString("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	plaintext, _ := hex.DecodeString("11111111aaaaaaaa11111111bbbbbbbb")
	expected, _ := hex.DecodeString("58a875e6044ad7fffa4f58420f7f442d")

	cipher, err := NewCipher256(key)
	if err != nil {
		t.Fatalf("NewCipher256 failed: %v", err)
	}

	ciphertext := make([]byte, BlockSize)
	cipher.Encrypt(ciphertext, plaintext)

	if !bytes.Equal(ciphertext, expected) {
		t.Errorf("ARIA-256 encryption failed\nExpected: %x\nGot:      %x", expected, ciphertext)
	}

	// Test decryption
	decrypted := make([]byte, BlockSize)
	cipher.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("ARIA-256 decryption failed\nExpected: %x\nGot:      %x", plaintext, decrypted)
	}
}

func TestNewCipherAutoDetect(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{"128-bit key", 16, false},
		{"192-bit key", 24, false},
		{"256-bit key", 32, false},
		{"Invalid key", 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := NewCipher(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCipher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockSize(t *testing.T) {
	key := make([]byte, KeySize128)
	cipher, _ := NewCipher(key)

	if cipher.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, want %d", cipher.BlockSize(), BlockSize)
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	keySizes := []int{KeySize128, KeySize192, KeySize256}

	for _, keySize := range keySizes {
		t.Run(string(rune(keySize*8))+"bit", func(t *testing.T) {
			key := make([]byte, keySize)
			for i := range key {
				key[i] = byte(i)
			}

			cipher, err := NewCipher(key)
			if err != nil {
				t.Fatalf("NewCipher failed: %v", err)
			}

			plaintext := []byte("0123456789ABCDEF") // 16 bytes
			ciphertext := make([]byte, BlockSize)
			decrypted := make([]byte, BlockSize)

			cipher.Encrypt(ciphertext, plaintext)
			cipher.Decrypt(decrypted, ciphertext)

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Roundtrip failed\nPlaintext:  %x\nDecrypted: %x", plaintext, decrypted)
			}
		})
	}
}

func TestDeterministic(t *testing.T) {
	key, _ := hex.DecodeString("00112233445566778899aabbccddeeff")
	plaintext, _ := hex.DecodeString("11111111aaaaaaaa11111111bbbbbbbb")

	cipher1, _ := NewCipher128(key)
	cipher2, _ := NewCipher128(key)

	ct1 := make([]byte, BlockSize)
	ct2 := make([]byte, BlockSize)

	cipher1.Encrypt(ct1, plaintext)
	cipher2.Encrypt(ct2, plaintext)

	if !bytes.Equal(ct1, ct2) {
		t.Error("ARIA encryption is not deterministic")
	}
}

// Benchmarks

func BenchmarkARIA128Encrypt(b *testing.B) {
	key := make([]byte, KeySize128)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	cipher, _ := NewCipher128(key)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cipher.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkARIA256Encrypt(b *testing.B) {
	key := make([]byte, KeySize256)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	cipher, _ := NewCipher256(key)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cipher.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkARIA128Decrypt(b *testing.B) {
	key := make([]byte, KeySize128)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	cipher, _ := NewCipher128(key)
	cipher.Encrypt(ciphertext, plaintext)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cipher.Decrypt(plaintext, ciphertext)
	}
}
