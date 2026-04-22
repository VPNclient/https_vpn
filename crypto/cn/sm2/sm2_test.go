package sm2

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	priv, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	if priv.D == nil {
		t.Error("private key D is nil")
	}
	if priv.X == nil || priv.Y == nil {
		t.Error("public key coordinates are nil")
	}

	// Verify public key is on curve
	if !priv.Curve.IsOnCurve(priv.X, priv.Y) {
		t.Error("public key not on curve")
	}
}

func TestSignVerify(t *testing.T) {
	priv, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	hash := []byte("test message hash 32 bytes long!")

	r, s, err := Sign(rand.Reader, priv, hash)
	if err != nil {
		t.Fatal(err)
	}

	if !Verify(&priv.PublicKey, hash, r, s) {
		t.Error("signature verification failed")
	}

	// Test with wrong hash
	wrongHash := []byte("wrong message hash 32 bytes!!!!")
	if Verify(&priv.PublicKey, wrongHash, r, s) {
		t.Error("verification should fail with wrong hash")
	}
}

func TestSignVerifyWithID(t *testing.T) {
	priv, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	uid := []byte("alice@example.com")
	msg := []byte("Hello, SM2!")

	r, s, err := SignWithID(rand.Reader, priv, uid, msg)
	if err != nil {
		t.Fatal(err)
	}

	if !VerifyWithID(&priv.PublicKey, uid, msg, r, s) {
		t.Error("signature verification with ID failed")
	}

	// Test with wrong ID
	wrongUID := []byte("bob@example.com")
	if VerifyWithID(&priv.PublicKey, wrongUID, msg, r, s) {
		t.Error("verification should fail with wrong ID")
	}

	// Test with wrong message
	wrongMsg := []byte("Goodbye, SM2!")
	if VerifyWithID(&priv.PublicKey, uid, wrongMsg, r, s) {
		t.Error("verification should fail with wrong message")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	priv, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	testCases := [][]byte{
		[]byte("Hello, SM2!"),
		[]byte("Short"),
		[]byte("A longer message that spans multiple blocks of data for testing purposes."),
		make([]byte, 100), // zeros
	}

	for _, plaintext := range testCases {
		ciphertext, err := Encrypt(rand.Reader, &priv.PublicKey, plaintext)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		decrypted, err := Decrypt(priv, ciphertext)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Decrypt(Encrypt(m)) != m")
		}
	}
}

func TestEncryptDecryptWrongKey(t *testing.T) {
	priv1, _ := GenerateKey(rand.Reader)
	priv2, _ := GenerateKey(rand.Reader)

	plaintext := []byte("Secret message")
	ciphertext, err := Encrypt(rand.Reader, &priv1.PublicKey, plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// Try to decrypt with wrong key
	_, err = Decrypt(priv2, ciphertext)
	if err == nil {
		t.Error("Decrypt should fail with wrong key")
	}
}

func TestSignerInterface(t *testing.T) {
	priv, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	hash := []byte("test message hash 32 bytes long!")

	sig, err := priv.Sign(rand.Reader, hash, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64", len(sig))
	}

	if !VerifySignature(&priv.PublicKey, hash, sig) {
		t.Error("VerifySignature failed")
	}
}

func TestP256Curve(t *testing.T) {
	curve := P256()

	if curve.Params().Name != "SM2-P256" {
		t.Errorf("curve name = %s, want SM2-P256", curve.Params().Name)
	}

	if curve.Params().BitSize != 256 {
		t.Errorf("bit size = %d, want 256", curve.Params().BitSize)
	}

	// Verify base point is on curve
	params := curve.Params()
	if !curve.IsOnCurve(params.Gx, params.Gy) {
		t.Error("base point not on curve")
	}
}

func BenchmarkSign(b *testing.B) {
	priv, _ := GenerateKey(rand.Reader)
	hash := []byte("test message hash 32 bytes long!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sign(rand.Reader, priv, hash)
	}
}

func BenchmarkVerify(b *testing.B) {
	priv, _ := GenerateKey(rand.Reader)
	hash := []byte("test message hash 32 bytes long!")
	r, s, _ := Sign(rand.Reader, priv, hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(&priv.PublicKey, hash, r, s)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	priv, _ := GenerateKey(rand.Reader)
	plaintext := make([]byte, 32)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encrypt(rand.Reader, &priv.PublicKey, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	priv, _ := GenerateKey(rand.Reader)
	plaintext := make([]byte, 32)
	ciphertext, _ := Encrypt(rand.Reader, &priv.PublicKey, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(priv, ciphertext)
	}
}
