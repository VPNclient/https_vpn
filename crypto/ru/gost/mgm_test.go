package gost

import (
	"bytes"
	"crypto/cipher"
	"testing"
)

func TestCTR(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	iv := make([]byte, 16)
	for i := range iv {
		iv[i] = byte(i + 32)
	}

	block, err := NewKuznyechik(key)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("Hello, GOST CTR mode! This is a test message.")
	ciphertext := make([]byte, len(plaintext))
	decrypted := make([]byte, len(plaintext))

	// Encrypt
	stream := NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)

	// Verify ciphertext is different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Ciphertext equals plaintext")
	}

	// Decrypt
	stream = NewCTR(block, iv)
	stream.XORKeyStream(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decryption failed:\ngot:  %x\nwant: %x", decrypted, plaintext)
	}
}

func TestCTRMagma(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	iv := make([]byte, 8)
	for i := range iv {
		iv[i] = byte(i + 32)
	}

	block, err := NewMagma(key)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("Magma CTR test!")
	ciphertext := make([]byte, len(plaintext))
	decrypted := make([]byte, len(plaintext))

	stream := NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)

	stream = NewCTR(block, iv)
	stream.XORKeyStream(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decryption failed:\ngot:  %x\nwant: %x", decrypted, plaintext)
	}
}

func TestMGMKuznyechik(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	block, err := NewKuznyechik(key)
	if err != nil {
		t.Fatal(err)
	}

	aead, err := NewMGM(block)
	if err != nil {
		t.Fatal(err)
	}

	// Verify sizes
	if aead.NonceSize() != 16 {
		t.Errorf("NonceSize = %d, want 16", aead.NonceSize())
	}
	if aead.Overhead() != 16 {
		t.Errorf("Overhead = %d, want 16", aead.Overhead())
	}

	nonce := make([]byte, 16)
	for i := range nonce {
		nonce[i] = byte(i + 100)
	}

	plaintext := []byte("MGM AEAD test with Kuznyechik cipher!")
	additionalData := []byte("associated data")

	// Seal
	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)
	if len(ciphertext) != len(plaintext)+16 {
		t.Errorf("Ciphertext length = %d, want %d", len(ciphertext), len(plaintext)+16)
	}

	// Open
	decrypted, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decryption failed:\ngot:  %x\nwant: %x", decrypted, plaintext)
	}
}

func TestMGMMagma(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	block, err := NewMagma(key)
	if err != nil {
		t.Fatal(err)
	}

	aead, err := NewMGM(block)
	if err != nil {
		t.Fatal(err)
	}

	if aead.NonceSize() != 8 {
		t.Errorf("NonceSize = %d, want 8", aead.NonceSize())
	}

	nonce := make([]byte, 8)
	for i := range nonce {
		nonce[i] = byte(i + 100)
	}

	plaintext := []byte("MGM with Magma!")
	additionalData := []byte("AD")

	ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)
	decrypted, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decryption failed")
	}
}

func TestMGMAuthFailure(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)
	aead, _ := NewMGM(block)

	nonce := make([]byte, 16)
	plaintext := []byte("test")
	ad := []byte("ad")

	ciphertext := aead.Seal(nil, nonce, plaintext, ad)

	// Tamper with ciphertext
	ciphertext[0] ^= 0xFF

	_, err := aead.Open(nil, nonce, ciphertext, ad)
	if err == nil {
		t.Error("Expected authentication failure")
	}
}

func TestMGMWrongAD(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)
	aead, _ := NewMGM(block)

	nonce := make([]byte, 16)
	plaintext := []byte("test")
	ad := []byte("correct AD")

	ciphertext := aead.Seal(nil, nonce, plaintext, ad)

	// Try with wrong AD
	_, err := aead.Open(nil, nonce, ciphertext, []byte("wrong AD"))
	if err == nil {
		t.Error("Expected authentication failure with wrong AD")
	}
}

func BenchmarkMGMSeal(b *testing.B) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)
	aead, _ := NewMGM(block)
	nonce := make([]byte, 16)
	plaintext := make([]byte, 1024)
	ad := make([]byte, 16)

	b.SetBytes(int64(len(plaintext)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aead.Seal(plaintext[:0], nonce, plaintext, ad)
	}
}

func BenchmarkMGMOpen(b *testing.B) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)
	aead, _ := NewMGM(block)
	nonce := make([]byte, 16)
	plaintext := make([]byte, 1024)
	ad := make([]byte, 16)
	ciphertext := aead.Seal(nil, nonce, plaintext, ad)

	b.SetBytes(int64(len(plaintext)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aead.Open(plaintext[:0], nonce, ciphertext, ad)
	}
}

// Test that MGM implements cipher.AEAD
var _ cipher.AEAD = (*mgm)(nil)
