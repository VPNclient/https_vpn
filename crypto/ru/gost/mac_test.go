package gost

import (
	"bytes"
	"hash"
	"testing"
)

func TestHMAC256(t *testing.T) {
	key := []byte("test key for HMAC")
	data := []byte("Hello, HMAC-Streebog!")

	mac1 := HMAC256(key, data)
	mac2 := HMAC256(key, data)

	if !bytes.Equal(mac1, mac2) {
		t.Error("HMAC not deterministic")
	}

	if len(mac1) != 32 {
		t.Errorf("HMAC256 length = %d, want 32", len(mac1))
	}
}

func TestHMAC512(t *testing.T) {
	key := []byte("test key for HMAC")
	data := []byte("Hello, HMAC-Streebog!")

	mac1 := HMAC512(key, data)
	mac2 := HMAC512(key, data)

	if !bytes.Equal(mac1, mac2) {
		t.Error("HMAC not deterministic")
	}

	if len(mac1) != 64 {
		t.Errorf("HMAC512 length = %d, want 64", len(mac1))
	}
}

func TestHMACDifferentKeys(t *testing.T) {
	data := []byte("Same message")

	mac1 := HMAC256([]byte("key1"), data)
	mac2 := HMAC256([]byte("key2"), data)

	if bytes.Equal(mac1, mac2) {
		t.Error("Different keys should produce different MACs")
	}
}

func TestHMACDifferentData(t *testing.T) {
	key := []byte("same key")

	mac1 := HMAC256(key, []byte("message1"))
	mac2 := HMAC256(key, []byte("message2"))

	if bytes.Equal(mac1, mac2) {
		t.Error("Different messages should produce different MACs")
	}
}

func TestHMACIncremental(t *testing.T) {
	key := []byte("test key")
	data := []byte("Hello, HMAC-Streebog incremental test!")

	// One-shot
	mac1 := HMAC256(key, data)

	// Incremental
	h := NewHMAC256(key)
	h.Write(data[:10])
	h.Write(data[10:25])
	h.Write(data[25:])
	mac2 := h.Sum(nil)

	if !bytes.Equal(mac1, mac2) {
		t.Error("Incremental HMAC doesn't match one-shot")
	}
}

func TestOMACKuznyechik(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	block, err := NewKuznyechik(key)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("Hello, OMAC with Kuznyechik!")

	mac1 := OMAC(block, data)
	mac2 := OMAC(block, data)

	if !bytes.Equal(mac1, mac2) {
		t.Error("OMAC not deterministic")
	}

	if len(mac1) != 16 {
		t.Errorf("OMAC length = %d, want 16", len(mac1))
	}
}

func TestOMACMagma(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	block, err := NewMagma(key)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("OMAC with Magma!")

	mac := OMAC(block, data)
	if len(mac) != 8 {
		t.Errorf("OMAC length = %d, want 8", len(mac))
	}
}

func TestOMACDifferentKeys(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1

	block1, _ := NewKuznyechik(key1)
	block2, _ := NewKuznyechik(key2)

	data := []byte("Same message")

	mac1 := OMAC(block1, data)
	mac2 := OMAC(block2, data)

	if bytes.Equal(mac1, mac2) {
		t.Error("Different keys should produce different MACs")
	}
}

func TestOMACIncremental(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)

	data := []byte("Hello, OMAC incremental test with longer data!")

	// One-shot
	mac1 := OMAC(block, data)

	// Incremental
	h := NewOMAC(block)
	h.Write(data[:10])
	h.Write(data[10:25])
	h.Write(data[25:])
	mac2 := h.Sum(nil)

	if !bytes.Equal(mac1, mac2) {
		t.Errorf("Incremental OMAC doesn't match one-shot:\none-shot: %x\nincr:     %x", mac1, mac2)
	}
}

func TestOMACReset(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)

	data := []byte("test data")

	h := NewOMAC(block)
	h.Write(data)
	mac1 := h.Sum(nil)

	h.Reset()
	h.Write(data)
	mac2 := h.Sum(nil)

	if !bytes.Equal(mac1, mac2) {
		t.Error("Reset doesn't work correctly")
	}
}

func TestOMACEmpty(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)

	mac := OMAC(block, nil)
	if len(mac) != 16 {
		t.Errorf("OMAC of empty message length = %d, want 16", len(mac))
	}
}

func TestOMACBlockAligned(t *testing.T) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)

	// Test with exactly 16 bytes (one block)
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(i)
	}

	mac := OMAC(block, data)
	if len(mac) != 16 {
		t.Errorf("OMAC length = %d, want 16", len(mac))
	}

	// Test with 32 bytes (two blocks)
	data = make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}

	mac = OMAC(block, data)
	if len(mac) != 16 {
		t.Errorf("OMAC length = %d, want 16", len(mac))
	}
}

// Verify hash.Hash interface
var _ hash.Hash = (*omac)(nil)

func BenchmarkHMAC256(b *testing.B) {
	key := make([]byte, 32)
	data := make([]byte, 1024)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HMAC256(key, data)
	}
}

func BenchmarkOMAC(b *testing.B) {
	key := make([]byte, 32)
	block, _ := NewKuznyechik(key)
	data := make([]byte, 1024)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OMAC(block, data)
	}
}
