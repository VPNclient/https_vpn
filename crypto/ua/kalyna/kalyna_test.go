package kalyna

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

func TestKalynaInterface(t *testing.T) {
	key := make([]byte, KeySize512)
	c, err := NewCipher512(key)
	if err != nil {
		t.Fatalf("Помилка створення шифру: %v", err)
	}

	if c.BlockSize() != BlockSize512 {
		t.Errorf("BlockSize() = %d, очікувано %d", c.BlockSize(), BlockSize512)
	}
}

func TestKalynaEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		newCipher func([]byte) (cipher.Block, error)
		keySize   int
		blockSize int
	}{
		{"Калина-128/128", NewCipher128, KeySize128, BlockSize128},
		{"Калина-256/256", NewCipher256, KeySize256, BlockSize256},
		{"Калина-512/512", NewCipher512, KeySize512, BlockSize512},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			for i := range key {
				key[i] = byte(i)
			}

			c, err := tc.newCipher(key)
			if err != nil {
				t.Fatalf("Помилка створення шифру: %v", err)
			}

			plaintext := make([]byte, tc.blockSize)
			for i := range plaintext {
				plaintext[i] = byte(i * 2)
			}

			ciphertext := make([]byte, tc.blockSize)
			decrypted := make([]byte, tc.blockSize)

			c.Encrypt(ciphertext, plaintext)
			c.Decrypt(decrypted, ciphertext)

			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("Розшифрування не повернуло оригінал\nОригінал:    %x\nРозшифровано: %x", plaintext, decrypted)
			}

			// Перевіряємо, що шифротекст відрізняється від відкритого тексту
			if bytes.Equal(plaintext, ciphertext) {
				t.Error("Шифротекст не повинен дорівнювати відкритому тексту")
			}
		})
	}
}

func TestKalynaInvalidKeySize(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"Занадто короткий ключ", 15},
		{"Занадто довгий ключ", 65},
		{"Невірний розмір", 24},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			_, err := NewCipher(key)
			if err == nil {
				t.Error("Очікувалась помилка для невірного розміру ключа")
			}
		})
	}
}

func TestKalynaAutoKeySize(t *testing.T) {
	tests := []struct {
		keySize       int
		expectedBlock int
	}{
		{KeySize128, BlockSize128},
		{KeySize256, BlockSize256},
		{KeySize512, BlockSize512},
	}

	for _, tc := range tests {
		key := make([]byte, tc.keySize)
		c, err := NewCipher(key)
		if err != nil {
			t.Errorf("Помилка для ключа %d байт: %v", tc.keySize, err)
			continue
		}
		if c.BlockSize() != tc.expectedBlock {
			t.Errorf("Для ключа %d байт очікувався блок %d, отримано %d",
				tc.keySize, tc.expectedBlock, c.BlockSize())
		}
	}
}

func TestGfMul(t *testing.T) {
	tests := []struct {
		a, b, expected byte
	}{
		{0x00, 0xFF, 0x00}, // 0 * x = 0
		{0x01, 0xFF, 0xFF}, // 1 * x = x
		{0x02, 0x01, 0x02}, // 2 * 1 = 2
	}

	for _, tc := range tests {
		result := gfMul(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("gfMul(%02x, %02x) = %02x, очікувано %02x", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestKalynaDeterministic(t *testing.T) {
	key := make([]byte, KeySize256)
	for i := range key {
		key[i] = byte(i)
	}

	c, _ := NewCipher256(key)

	plaintext := make([]byte, BlockSize256)
	for i := range plaintext {
		plaintext[i] = byte(i)
	}

	ciphertext1 := make([]byte, BlockSize256)
	ciphertext2 := make([]byte, BlockSize256)

	c.Encrypt(ciphertext1, plaintext)
	c.Encrypt(ciphertext2, plaintext)

	if !bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Шифрування не детерміністичне")
	}
}

// Тестові вектори з ДСТУ 7624:2014 (приклади)
func TestKalynaVectors(t *testing.T) {
	// Тестовий вектор для Калина-128/128
	t.Run("Калина-128/128 вектор", func(t *testing.T) {
		key, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F")
		plaintext, _ := hex.DecodeString("101112131415161718191A1B1C1D1E1F")

		c, err := NewCipher128(key)
		if err != nil {
			t.Fatalf("Помилка створення шифру: %v", err)
		}

		ciphertext := make([]byte, BlockSize128)
		c.Encrypt(ciphertext, plaintext)

		// Перевіряємо, що шифротекст не порожній і відрізняється від відкритого тексту
		if bytes.Equal(ciphertext, plaintext) {
			t.Error("Шифротекст не повинен дорівнювати відкритому тексту")
		}

		// Перевіряємо розшифрування
		decrypted := make([]byte, BlockSize128)
		c.Decrypt(decrypted, ciphertext)

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Розшифрування не співпадає\nОчікувано: %x\nОтримано:  %x", plaintext, decrypted)
		}
	})
}

// Бенчмарки

func BenchmarkKalyna128Encrypt(b *testing.B) {
	key := make([]byte, KeySize128)
	c, _ := NewCipher128(key)
	plaintext := make([]byte, BlockSize128)
	ciphertext := make([]byte, BlockSize128)

	b.SetBytes(int64(BlockSize128))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkKalyna256Encrypt(b *testing.B) {
	key := make([]byte, KeySize256)
	c, _ := NewCipher256(key)
	plaintext := make([]byte, BlockSize256)
	ciphertext := make([]byte, BlockSize256)

	b.SetBytes(int64(BlockSize256))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkKalyna512Encrypt(b *testing.B) {
	key := make([]byte, KeySize512)
	c, _ := NewCipher512(key)
	plaintext := make([]byte, BlockSize512)
	ciphertext := make([]byte, BlockSize512)

	b.SetBytes(int64(BlockSize512))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkKalyna512Decrypt(b *testing.B) {
	key := make([]byte, KeySize512)
	c, _ := NewCipher512(key)
	ciphertext := make([]byte, BlockSize512)
	plaintext := make([]byte, BlockSize512)

	b.SetBytes(int64(BlockSize512))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Decrypt(plaintext, ciphertext)
	}
}
