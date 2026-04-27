package sokil

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestKeyGeneration(t *testing.T) {
	kp, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Помилка генерації ключів: %v", err)
	}

	if kp.Public == nil {
		t.Fatal("Відкритий ключ nil")
	}

	if kp.Private == nil {
		t.Fatal("Закритий ключ nil")
	}

	// Перевіряємо розмір відкритого ключа
	pkBytes := kp.Public.Bytes()
	if len(pkBytes) == 0 {
		t.Error("Відкритий ключ порожній")
	}
}

func TestSignVerify(t *testing.T) {
	kp, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Помилка генерації ключів: %v", err)
	}

	message := []byte("Тестове повідомлення для підпису")

	sig, err := Sign(kp.Private, message)
	if err != nil {
		t.Fatalf("Помилка підпису: %v", err)
	}

	if len(sig) == 0 {
		t.Fatal("Підпис порожній")
	}

	// Верифікація
	valid := Verify(kp.Public, message, sig)
	if !valid {
		t.Error("Верифікація підпису не пройшла")
	}
}

func TestSignVerifyDifferentMessages(t *testing.T) {
	kp, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Помилка генерації ключів: %v", err)
	}

	message1 := []byte("Повідомлення 1")
	message2 := []byte("Повідомлення 2")

	sig, err := Sign(kp.Private, message1)
	if err != nil {
		t.Fatalf("Помилка підпису: %v", err)
	}

	// Верифікація з іншим повідомленням повинна провалитись
	valid := Verify(kp.Public, message2, sig)
	if valid {
		t.Error("Верифікація не повинна пройти для іншого повідомлення")
	}
}

func TestSignVerifyWrongKey(t *testing.T) {
	kp1, _ := GenerateKey(rand.Reader)
	kp2, _ := GenerateKey(rand.Reader)

	message := []byte("Тестове повідомлення")

	sig, err := Sign(kp1.Private, message)
	if err != nil {
		t.Fatalf("Помилка підпису: %v", err)
	}

	// Верифікація з іншим ключем повинна провалитись
	valid := Verify(kp2.Public, message, sig)
	if valid {
		t.Error("Верифікація не повинна пройти для іншого ключа")
	}
}

func TestSignVerifyCorruptedSignature(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)
	message := []byte("Тестове повідомлення")

	sig, err := Sign(kp.Private, message)
	if err != nil {
		t.Fatalf("Помилка підпису: %v", err)
	}

	// Пошкоджуємо підпис
	if len(sig) > 10 {
		sig[10] ^= 0xFF
	}

	valid := Verify(kp.Public, message, sig)
	if valid {
		t.Error("Верифікація не повинна пройти для пошкодженого підпису")
	}
}

func TestMultipleSignatures(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)

	messages := [][]byte{
		[]byte("Перше повідомлення"),
		[]byte("Друге повідомлення"),
		[]byte("Третє повідомлення"),
	}

	for i, msg := range messages {
		sig, err := Sign(kp.Private, msg)
		if err != nil {
			t.Fatalf("Повідомлення %d: помилка підпису: %v", i, err)
		}

		valid := Verify(kp.Public, msg, sig)
		if !valid {
			t.Errorf("Повідомлення %d: верифікація не пройшла", i)
		}
	}
}

func TestEmptyMessage(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)
	message := []byte{}

	sig, err := Sign(kp.Private, message)
	if err != nil {
		t.Fatalf("Помилка підпису порожнього повідомлення: %v", err)
	}

	valid := Verify(kp.Public, message, sig)
	if !valid {
		t.Error("Верифікація порожнього повідомлення не пройшла")
	}
}

func TestLargeMessage(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)

	// 1 MB повідомлення
	message := make([]byte, 1024*1024)
	rand.Read(message)

	sig, err := Sign(kp.Private, message)
	if err != nil {
		t.Fatalf("Помилка підпису великого повідомлення: %v", err)
	}

	valid := Verify(kp.Public, message, sig)
	if !valid {
		t.Error("Верифікація великого повідомлення не пройшла")
	}
}

func TestPower2Round(t *testing.T) {
	tests := []int32{0, 1, Q / 2, Q - 1, 12345}

	for _, a := range tests {
		a0, a1 := power2Round(a)
		reconstructed := a1*(1<<13) + a0

		// Має бути близьким до оригінального значення
		diff := a - reconstructed
		if diff < 0 {
			diff = -diff
		}
		if diff > (1 << 12) {
			t.Errorf("power2Round(%d): різниця %d занадто велика", a, diff)
		}
	}
}

func TestDecompose(t *testing.T) {
	tests := []int32{0, 1, Q / 4, Q / 2, Q - 1}

	for _, a := range tests {
		a0, a1 := decompose(a)
		_ = a0
		_ = a1
		// Базова перевірка що функція не панікує
	}
}

func TestPolyCheckNorm(t *testing.T) {
	var p Poly

	// Поліном з малими коефіцієнтами
	for i := 0; i < N; i++ {
		p[i] = int32(i % 10)
	}

	if !p.checkNorm(100) {
		t.Error("checkNorm має повертати true для малих коефіцієнтів")
	}

	// Поліном з великими коефіцієнтами
	p[0] = 1000
	if p.checkNorm(100) {
		t.Error("checkNorm має повертати false для великих коефіцієнтів")
	}
}

func TestSignatureDeterminism(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)
	message := []byte("Тестове повідомлення")

	sig1, _ := Sign(kp.Private, message)
	sig2, _ := Sign(kp.Private, message)

	// Підписи можуть бути різними через randomization
	// Але обидва повинні верифікуватись
	if !Verify(kp.Public, message, sig1) {
		t.Error("Перший підпис не верифікується")
	}
	if !Verify(kp.Public, message, sig2) {
		t.Error("Другий підпис не верифікується")
	}
}

func TestPublicKeyBytes(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)
	pkBytes := kp.Public.Bytes()

	if len(pkBytes) == 0 {
		t.Error("Серіалізований відкритий ключ порожній")
	}

	// Перевіряємо, що rho на початку
	if !bytes.Equal(pkBytes[:SeedBytes], kp.Public.Rho[:]) {
		t.Error("Rho не на початку серіалізованого ключа")
	}
}

// Бенчмарки

func BenchmarkKeyGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKey(rand.Reader)
	}
}

func BenchmarkSign(b *testing.B) {
	kp, _ := GenerateKey(rand.Reader)
	message := []byte("Бенчмарк повідомлення")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sign(kp.Private, message)
	}
}

func BenchmarkVerify(b *testing.B) {
	kp, _ := GenerateKey(rand.Reader)
	message := []byte("Бенчмарк повідомлення")
	sig, _ := Sign(kp.Private, message)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Verify(kp.Public, message, sig)
	}
}
