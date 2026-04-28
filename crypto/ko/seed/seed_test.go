package seed

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSEED(t *testing.T) {
	// Test vector from RFC 4269
	tests := []struct {
		name       string
		key        string
		plaintext  string
		ciphertext string
	}{
		{
			name:       "Vector 1",
			key:        "00000000000000000000000000000000",
			plaintext:  "000102030405060708090A0B0C0D0E0F",
			ciphertext: "5EBAC6E0054E166819AFF1CC6D346CDB",
		},
		{
			name:       "Vector 2",
			key:        "00010203040506070809101112131415",
			plaintext:  "00000000000000000000000000000000",
			ciphertext: "C11F22F20140505084483597E4370F43",
		},
		{
			name:       "Vector 3",
			key:        "47064808BBEEF28E0A7E5C2E9740F073",
			plaintext:  "83A2F8A288641FB9A4E9A5CC2F131C7D",
			ciphertext: "EE54D13EBCAE706D226BC3142CD40D4A",
		},
		{
			name:       "Vector 4",
			key:        "28DBC3BC49FFD87DCFA509B11D422BE7",
			plaintext:  "B41E6BE2EBA84A148E2EED84593C5EC7",
			ciphertext: "9B9B7BFCD1813CB95D0B3618F40F5122",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := hex.DecodeString(tt.key)
			plaintext, _ := hex.DecodeString(tt.plaintext)
			expected, _ := hex.DecodeString(tt.ciphertext)

			cipher, err := NewCipher(key)
			if err != nil {
				t.Fatalf("NewCipher failed: %v", err)
			}

			// Test encryption
			ciphertext := make([]byte, BlockSize)
			cipher.Encrypt(ciphertext, plaintext)

			if !bytes.Equal(ciphertext, expected) {
				t.Errorf("Encryption failed\nExpected: %X\nGot:      %X", expected, ciphertext)
			}

			// Test decryption
			decrypted := make([]byte, BlockSize)
			cipher.Decrypt(decrypted, ciphertext)

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Decryption failed\nExpected: %X\nGot:      %X", plaintext, decrypted)
			}
		})
	}
}

func TestInvalidKeySize(t *testing.T) {
	tests := []int{0, 8, 15, 17, 24, 32}

	for _, size := range tests {
		key := make([]byte, size)
		_, err := NewCipher(key)
		if err == nil {
			t.Errorf("Expected error for key size %d", size)
		}
	}
}

func TestBlockSize(t *testing.T) {
	key := make([]byte, KeySize)
	cipher, _ := NewCipher(key)

	if cipher.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, want %d", cipher.BlockSize(), BlockSize)
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16 bytes
	plaintext := []byte("Hello, SEED!!!!")

	cipher, err := NewCipher(key)
	if err != nil {
		t.Fatalf("NewCipher failed: %v", err)
	}

	ciphertext := make([]byte, BlockSize)
	decrypted := make([]byte, BlockSize)

	cipher.Encrypt(ciphertext, plaintext)
	cipher.Decrypt(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Roundtrip failed\nPlaintext:  %X\nDecrypted: %X", plaintext, decrypted)
	}
}

func TestDeterministic(t *testing.T) {
	key := make([]byte, KeySize)
	plaintext := make([]byte, BlockSize)

	cipher1, _ := NewCipher(key)
	cipher2, _ := NewCipher(key)

	ct1 := make([]byte, BlockSize)
	ct2 := make([]byte, BlockSize)

	cipher1.Encrypt(ct1, plaintext)
	cipher2.Encrypt(ct2, plaintext)

	if !bytes.Equal(ct1, ct2) {
		t.Error("SEED encryption is not deterministic")
	}
}

// Benchmarks

func BenchmarkSEEDEncrypt(b *testing.B) {
	key := make([]byte, KeySize)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	cipher, _ := NewCipher(key)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cipher.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkSEEDDecrypt(b *testing.B) {
	key := make([]byte, KeySize)
	plaintext := make([]byte, BlockSize)
	ciphertext := make([]byte, BlockSize)

	cipher, _ := NewCipher(key)
	cipher.Encrypt(ciphertext, plaintext)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cipher.Decrypt(plaintext, ciphertext)
	}
}
