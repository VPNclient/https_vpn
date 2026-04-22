package gost

import (
	"crypto/rand"
	"testing"
)

func TestGenerateKey256(t *testing.T) {
	curve := TC26256A()
	priv, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Check private key in valid range
	if priv.D.Sign() <= 0 || priv.D.Cmp(curve.N) >= 0 {
		t.Error("Private key out of range")
	}

	// Check public key is on curve
	if !curve.IsOnCurve(priv.X, priv.Y) {
		t.Error("Public key not on curve")
	}
}

func TestGenerateKey512(t *testing.T) {
	curve := TC26512A()
	priv, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if priv.D.Sign() <= 0 || priv.D.Cmp(curve.N) >= 0 {
		t.Error("Private key out of range")
	}

	if !curve.IsOnCurve(priv.X, priv.Y) {
		t.Error("Public key not on curve")
	}
}

func TestSignVerify256(t *testing.T) {
	curve := TC26256A()
	priv, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	hash := make([]byte, 32) // 256-bit hash
	rand.Read(hash)

	r, s, err := Sign(rand.Reader, priv, hash)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if !Verify(&priv.PublicKey, hash, r, s) {
		t.Error("Verify failed for valid signature")
	}
}

func TestSignVerify512(t *testing.T) {
	curve := TC26512A()
	priv, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	hash := make([]byte, 64) // 512-bit hash
	rand.Read(hash)

	r, s, err := Sign(rand.Reader, priv, hash)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if !Verify(&priv.PublicKey, hash, r, s) {
		t.Error("Verify failed for valid signature")
	}
}

func TestSignVerifySignature(t *testing.T) {
	curve := TC26256A()
	priv, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	hash := make([]byte, 32)
	rand.Read(hash)

	sig, err := priv.Sign(rand.Reader, hash, nil)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if len(sig) != 64 { // 2 * 32 bytes for 256-bit curve
		t.Errorf("Signature length = %d, want 64", len(sig))
	}

	if !VerifySignature(&priv.PublicKey, hash, sig) {
		t.Error("VerifySignature failed for valid signature")
	}
}

func TestVerifyTamperedHash(t *testing.T) {
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)

	hash := make([]byte, 32)
	rand.Read(hash)

	r, s, _ := Sign(rand.Reader, priv, hash)

	// Tamper with hash
	hash[0] ^= 0xFF

	if Verify(&priv.PublicKey, hash, r, s) {
		t.Error("Verify should fail for tampered hash")
	}
}

func TestVerifyTamperedSignature(t *testing.T) {
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)

	hash := make([]byte, 32)
	rand.Read(hash)

	sig, _ := priv.Sign(rand.Reader, hash, nil)

	// Tamper with signature
	sig[0] ^= 0xFF

	if VerifySignature(&priv.PublicKey, hash, sig) {
		t.Error("Verify should fail for tampered signature")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	curve := TC26256A()
	priv1, _ := GenerateKey(curve, rand.Reader)
	priv2, _ := GenerateKey(curve, rand.Reader)

	hash := make([]byte, 32)
	rand.Read(hash)

	r, s, _ := Sign(rand.Reader, priv1, hash)

	// Verify with wrong public key
	if Verify(&priv2.PublicKey, hash, r, s) {
		t.Error("Verify should fail for wrong public key")
	}
}

func TestSignDeterministic(t *testing.T) {
	// Signatures should be different each time (using random k)
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)

	hash := make([]byte, 32)
	rand.Read(hash)

	sig1, _ := priv.Sign(rand.Reader, hash, nil)
	sig2, _ := priv.Sign(rand.Reader, hash, nil)

	// Both should verify
	if !VerifySignature(&priv.PublicKey, hash, sig1) {
		t.Error("First signature doesn't verify")
	}
	if !VerifySignature(&priv.PublicKey, hash, sig2) {
		t.Error("Second signature doesn't verify")
	}

	// But they should be different (different random k)
	same := true
	for i := range sig1 {
		if sig1[i] != sig2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("Two signatures of same message should differ (random k)")
	}
}

func TestPublicKeyMethod(t *testing.T) {
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)

	pub := priv.Public()
	pubKey, ok := pub.(*PublicKey)
	if !ok {
		t.Fatal("Public() didn't return *PublicKey")
	}

	if pubKey.X.Cmp(priv.X) != 0 || pubKey.Y.Cmp(priv.Y) != 0 {
		t.Error("Public key mismatch")
	}
}

func BenchmarkGenerateKey256(b *testing.B) {
	curve := TC26256A()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateKey(curve, rand.Reader)
	}
}

func BenchmarkSign256(b *testing.B) {
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)
	hash := make([]byte, 32)
	rand.Read(hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sign(rand.Reader, priv, hash)
	}
}

func BenchmarkVerify256(b *testing.B) {
	curve := TC26256A()
	priv, _ := GenerateKey(curve, rand.Reader)
	hash := make([]byte, 32)
	rand.Read(hash)
	r, s, _ := Sign(rand.Reader, priv, hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(&priv.PublicKey, hash, r, s)
	}
}
