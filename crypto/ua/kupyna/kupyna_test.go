package kupyna

import (
	"bytes"
	"encoding/hex"
	"hash"
	"testing"
)

// Тестові вектори з ДСТУ 7564:2014
var testVectors = []struct {
	name     string
	size     int // 256 або 512
	message  string
	expected string
}{
	{
		name:     "Купина-256: порожнє повідомлення",
		size:     256,
		message:  "",
		expected: "cd5101d1ccdf0d1d1f4ada56e888cd724ca1a0838a3521e7131d4fb78d0f5eb6",
	},
	{
		name:     "Купина-512: порожнє повідомлення",
		size:     512,
		message:  "",
		expected: "656b2f4cd71462388b64a37043ea55dbe445d452aecd46c3298343314ef04019bcfa3f04265a9857f91be91fce197096187ceda78c9c1c021c294a0689198538",
	},
	{
		name:    "Купина-256: коротке повідомлення",
		size:    256,
		message: "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
		expected: "08f4ee6f1be6903b324c4e27990cb24ef69dd58dbe84813ee0a52f6631239875",
	},
	{
		name:    "Купина-512: коротке повідомлення",
		size:    512,
		message: "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
		expected: "3813e2109118cdfb5a6d5e72f7208dccc80a2dfb3afdfb02f46992b5edbe536b3560dd1d7e29c6f53978af58b444e37ba685c0dd910533ba5d78efffc13de62a",
	},
}

func TestKupynaVectors(t *testing.T) {
	for _, tc := range testVectors {
		t.Run(tc.name, func(t *testing.T) {
			var h hash.Hash
			if tc.size == 256 {
				h = New256()
			} else {
				h = New512()
			}

			// Декодуємо повідомлення з hex
			msg, err := hex.DecodeString(tc.message)
			if err != nil {
				t.Fatalf("Помилка декодування повідомлення: %v", err)
			}

			h.Write(msg)
			result := h.Sum(nil)
			expected, _ := hex.DecodeString(tc.expected)

			if !bytes.Equal(result, expected) {
				t.Errorf("Хеш не співпадає\nОтримано:  %x\nОчікувано: %x", result, expected)
			}
		})
	}
}

func TestKupyna512Interface(t *testing.T) {
	h := New512()

	// Перевіряємо Size()
	if h.Size() != Size512 {
		t.Errorf("Size() = %d, очікувано %d", h.Size(), Size512)
	}

	// Перевіряємо BlockSize()
	if h.BlockSize() != BlockSize*2 {
		t.Errorf("BlockSize() = %d, очікувано %d", h.BlockSize(), BlockSize*2)
	}
}

func TestKupyna256Interface(t *testing.T) {
	h := New256()

	// Перевіряємо Size()
	if h.Size() != Size256 {
		t.Errorf("Size() = %d, очікувано %d", h.Size(), Size256)
	}

	// Перевіряємо BlockSize()
	if h.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, очікувано %d", h.BlockSize(), BlockSize)
	}
}

func TestKupynaReset(t *testing.T) {
	h := New512()
	data := []byte("тестові дані")

	h.Write(data)
	first := h.Sum(nil)

	h.Reset()
	h.Write(data)
	second := h.Sum(nil)

	if !bytes.Equal(first, second) {
		t.Error("Reset() не скидає стан правильно")
	}
}

func TestKupynaIncremental(t *testing.T) {
	data := []byte("це тестове повідомлення для перевірки інкрементального хешування")

	// Хеш всього повідомлення
	h1 := New512()
	h1.Write(data)
	full := h1.Sum(nil)

	// Хеш частинами
	h2 := New512()
	h2.Write(data[:10])
	h2.Write(data[10:30])
	h2.Write(data[30:])
	incremental := h2.Sum(nil)

	if !bytes.Equal(full, incremental) {
		t.Error("Інкрементальне хешування дає інший результат")
	}
}

func TestSum512(t *testing.T) {
	data := []byte("тест")
	result := Sum512(data)

	h := New512()
	h.Write(data)
	expected := h.Sum(nil)

	if !bytes.Equal(result[:], expected) {
		t.Error("Sum512() не співпадає з New512().Sum()")
	}
}

func TestSum256(t *testing.T) {
	data := []byte("тест")
	result := Sum256(data)

	h := New256()
	h.Write(data)
	expected := h.Sum(nil)

	if !bytes.Equal(result[:], expected) {
		t.Error("Sum256() не співпадає з New256().Sum()")
	}
}

func TestGfMul(t *testing.T) {
	// Базові властивості множення в GF(2^8)
	tests := []struct {
		a, b, expected byte
	}{
		{0x00, 0xFF, 0x00}, // 0 * x = 0
		{0x01, 0xFF, 0xFF}, // 1 * x = x
		{0x02, 0x01, 0x02}, // 2 * 1 = 2
		{0x02, 0x80, 0x1D}, // 2 * 0x80 = редукція
	}

	for _, tc := range tests {
		result := gfMul(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("gfMul(%02x, %02x) = %02x, очікувано %02x", tc.a, tc.b, result, tc.expected)
		}
	}
}

// Бенчмарки

func BenchmarkKupyna512_64B(b *testing.B) {
	data := make([]byte, 64)
	b.SetBytes(64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sum512(data)
	}
}

func BenchmarkKupyna512_1KB(b *testing.B) {
	data := make([]byte, 1024)
	b.SetBytes(1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sum512(data)
	}
}

func BenchmarkKupyna512_64KB(b *testing.B) {
	data := make([]byte, 64*1024)
	b.SetBytes(64 * 1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sum512(data)
	}
}

func BenchmarkKupyna256_1KB(b *testing.B) {
	data := make([]byte, 1024)
	b.SetBytes(1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sum256(data)
	}
}
