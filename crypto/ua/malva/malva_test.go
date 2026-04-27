package malva

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

	// Перевіряємо розміри
	pkBytes := kp.Public.Bytes()
	if len(pkBytes) != PublicKeySize {
		t.Errorf("Розмір відкритого ключа = %d, очікувано %d", len(pkBytes), PublicKeySize)
	}

	skBytes := kp.Private.Bytes()
	if len(skBytes) != PrivateKeySize {
		t.Errorf("Розмір закритого ключа = %d, очікувано %d", len(skBytes), PrivateKeySize)
	}
}

func TestEncapsulateDecapsulate(t *testing.T) {
	// Генеруємо ключі
	kp, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Помилка генерації ключів: %v", err)
	}

	// Encapsulate
	ct, ss1, err := Encapsulate(kp.Public, rand.Reader)
	if err != nil {
		t.Fatalf("Помилка encapsulate: %v", err)
	}

	if len(ct) != CiphertextSize {
		t.Errorf("Розмір шифротексту = %d, очікувано %d", len(ct), CiphertextSize)
	}

	if len(ss1) != SharedSecretSize {
		t.Errorf("Розмір спільного секрету = %d, очікувано %d", len(ss1), SharedSecretSize)
	}

	// Decapsulate
	ss2, err := Decapsulate(kp.Private, ct)
	if err != nil {
		t.Fatalf("Помилка decapsulate: %v", err)
	}

	// Перевіряємо, що секрети співпадають
	if !bytes.Equal(ss1, ss2) {
		t.Error("Спільні секрети не співпадають")
		t.Logf("ss1: %x", ss1)
		t.Logf("ss2: %x", ss2)
	}
}

func TestEncapsulateDecapsulateMultiple(t *testing.T) {
	kp, err := GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Помилка генерації ключів: %v", err)
	}

	for i := 0; i < 10; i++ {
		ct, ss1, err := Encapsulate(kp.Public, rand.Reader)
		if err != nil {
			t.Fatalf("Ітерація %d: помилка encapsulate: %v", i, err)
		}

		ss2, err := Decapsulate(kp.Private, ct)
		if err != nil {
			t.Fatalf("Ітерація %d: помилка decapsulate: %v", i, err)
		}

		if !bytes.Equal(ss1, ss2) {
			t.Errorf("Ітерація %d: секрети не співпадають", i)
		}
	}
}

func TestDifferentKeysProduceDifferentSecrets(t *testing.T) {
	kp1, _ := GenerateKey(rand.Reader)
	kp2, _ := GenerateKey(rand.Reader)

	ct, ss1, _ := Encapsulate(kp1.Public, rand.Reader)

	// Спроба розшифрувати іншим ключем
	ss2, _ := Decapsulate(kp2.Private, ct)

	// Секрети повинні бути різними (implicit rejection)
	if bytes.Equal(ss1, ss2) {
		t.Error("Секрети не повинні співпадати для різних ключів")
	}
}

func TestInvalidCiphertext(t *testing.T) {
	kp, _ := GenerateKey(rand.Reader)

	// Пошкоджений шифротекст
	ct := make([]byte, CiphertextSize)
	rand.Read(ct)

	// Decapsulate повинен працювати (implicit rejection)
	ss, err := Decapsulate(kp.Private, ct)
	if err != nil {
		t.Fatalf("Decapsulate не повинен повертати помилку: %v", err)
	}

	if len(ss) != SharedSecretSize {
		t.Error("Розмір секрету неправильний")
	}
}

func TestNTT(t *testing.T) {
	// Тестуємо NTT -> InvNTT = identity
	var p Poly
	for i := 0; i < N; i++ {
		p[i] = int16(i % Q)
	}

	original := p

	p.NTT()
	p.InvNTT()

	for i := 0; i < N; i++ {
		expected := barrettReduce(original[i])
		got := barrettReduce(p[i])
		if expected != got {
			t.Errorf("NTT->InvNTT[%d]: очікувано %d, отримано %d", i, expected, got)
		}
	}
}

func TestPolyAdd(t *testing.T) {
	var a, b, c Poly

	for i := 0; i < N; i++ {
		a[i] = int16(i)
		b[i] = int16(N - i)
	}

	c.Add(&a, &b)

	for i := 0; i < N; i++ {
		expected := a[i] + b[i]
		if c[i] != expected {
			t.Errorf("Add[%d]: очікувано %d, отримано %d", i, expected, c[i])
		}
	}
}

func TestPolySub(t *testing.T) {
	var a, b, c Poly

	for i := 0; i < N; i++ {
		a[i] = int16(N)
		b[i] = int16(i)
	}

	c.Sub(&a, &b)

	for i := 0; i < N; i++ {
		expected := a[i] - b[i]
		if c[i] != expected {
			t.Errorf("Sub[%d]: очікувано %d, отримано %d", i, expected, c[i])
		}
	}
}

func TestCompressDecompress(t *testing.T) {
	tests := []struct {
		value int16
		d     int
	}{
		{0, 11},
		{Q / 2, 11},
		{Q - 1, 11},
		{100, 5},
		{1000, 5},
	}

	for _, tc := range tests {
		compressed := compress(tc.value, tc.d)
		decompressed := decompress(compressed, tc.d)

		// Допустима похибка через втрату точності
		diff := tc.value - decompressed
		if diff < 0 {
			diff = -diff
		}
		maxError := int16(Q / (1 << tc.d))
		if diff > maxError {
			t.Errorf("compress/decompress(%d, %d): різниця %d > %d",
				tc.value, tc.d, diff, maxError)
		}
	}
}

func TestBarrettReduce(t *testing.T) {
	tests := []int16{0, 1, Q - 1, Q, Q + 1, 2 * Q, -1, -Q}

	for _, x := range tests {
		result := barrettReduce(x)
		if result < 0 || result >= Q {
			t.Errorf("barrettReduce(%d) = %d, поза межами [0, Q)", x, result)
		}

		expected := x % Q
		if expected < 0 {
			expected += Q
		}
		if result != expected {
			t.Errorf("barrettReduce(%d) = %d, очікувано %d", x, result, expected)
		}
	}
}

func TestConstantTimeCompare(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3, 4, 5}
	c := []byte{1, 2, 3, 4, 6}
	d := []byte{1, 2, 3}

	if constantTimeCompare(a, b) != 1 {
		t.Error("Однакові зрізи повинні давати 1")
	}

	if constantTimeCompare(a, c) != 0 {
		t.Error("Різні зрізи повинні давати 0")
	}

	if constantTimeCompare(a, d) != 0 {
		t.Error("Зрізи різної довжини повинні давати 0")
	}
}

// Бенчмарки

func BenchmarkKeyGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKey(rand.Reader)
	}
}

func BenchmarkEncapsulate(b *testing.B) {
	kp, _ := GenerateKey(rand.Reader)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Encapsulate(kp.Public, rand.Reader)
	}
}

func BenchmarkDecapsulate(b *testing.B) {
	kp, _ := GenerateKey(rand.Reader)
	ct, _, _ := Encapsulate(kp.Public, rand.Reader)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Decapsulate(kp.Private, ct)
	}
}

func BenchmarkNTT(b *testing.B) {
	var p Poly
	for i := 0; i < N; i++ {
		p[i] = int16(i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.NTT()
	}
}

func BenchmarkInvNTT(b *testing.B) {
	var p Poly
	for i := 0; i < N; i++ {
		p[i] = int16(i)
	}
	p.NTT()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.InvNTT()
	}
}
