package sm4

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGCMRoundtrip(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewGCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	plaintext := []byte("Hello, SM4-GCM!")
	additionalData := []byte("additional data")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	decrypted, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted text doesn't match plaintext")
	}
}

func TestGCMAuthFailure(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewGCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	plaintext := []byte("Hello, SM4-GCM!")
	additionalData := []byte("additional data")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	// Tamper with ciphertext
	ciphertext[0] ^= 0xFF

	_, err = aead.Open(nil, nonce, ciphertext, additionalData)
	if err == nil {
		t.Error("Open should fail with tampered ciphertext")
	}
}

func TestCCMRoundtrip(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewCCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	plaintext := []byte("Hello, SM4-CCM!")
	additionalData := []byte("additional data")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	decrypted, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted text doesn't match plaintext")
	}
}

func TestCCMAuthFailure(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewCCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	plaintext := []byte("Hello, SM4-CCM!")
	additionalData := []byte("additional data")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	// Tamper with ciphertext
	ciphertext[0] ^= 0xFF

	_, err = aead.Open(nil, nonce, ciphertext, additionalData)
	if err == nil {
		t.Error("Open should fail with tampered ciphertext")
	}
}

func TestCCMWrongAdditionalData(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewCCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	plaintext := []byte("Hello, SM4-CCM!")
	additionalData := []byte("additional data")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

	// Try to decrypt with wrong additional data
	_, err = aead.Open(nil, nonce, ciphertext, []byte("wrong data"))
	if err == nil {
		t.Error("Open should fail with wrong additional data")
	}
}

func TestCCMVariousLengths(t *testing.T) {
	key := make([]byte, KeySize)
	rand.Read(key)

	aead, err := NewCCM(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)

	lengths := []int{0, 1, 15, 16, 17, 31, 32, 100, 1000}

	for _, length := range lengths {
		plaintext := make([]byte, length)
		rand.Read(plaintext)

		ciphertext := aead.Seal(nil, nonce, plaintext, nil)
		decrypted, err := aead.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			t.Fatalf("length %d: Open failed: %v", length, err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("length %d: decrypted doesn't match", length)
		}
	}
}

func BenchmarkGCMSeal(b *testing.B) {
	key := make([]byte, KeySize)
	aead, _ := NewGCM(key)
	nonce := make([]byte, aead.NonceSize())
	plaintext := make([]byte, 1024)
	dst := make([]byte, 0, len(plaintext)+aead.Overhead())

	b.SetBytes(int64(len(plaintext)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aead.Seal(dst[:0], nonce, plaintext, nil)
	}
}

func BenchmarkCCMSeal(b *testing.B) {
	key := make([]byte, KeySize)
	aead, _ := NewCCM(key)
	nonce := make([]byte, aead.NonceSize())
	plaintext := make([]byte, 1024)
	dst := make([]byte, 0, len(plaintext)+aead.Overhead())

	b.SetBytes(int64(len(plaintext)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aead.Seal(dst[:0], nonce, plaintext, nil)
	}
}
