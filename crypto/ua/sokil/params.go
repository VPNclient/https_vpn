// Package sokil реалізує цифровий підпис Сокіл на базі Module-LWE/SIS.
// Аналог ML-DSA (Dilithium-5) для постквантової стійкості Category 5.
package sokil

// Параметри Сокола-512 (аналог Dilithium-5)
const (
	// N - розмірність поліному
	N = 256

	// Q - модуль
	Q = 8380417

	// K - кількість рядків у матриці A
	K = 8

	// L - кількість стовпців у матриці A
	L = 7

	// Eta - параметр для секретних векторів
	Eta = 2

	// Tau - кількість ненульових коефіцієнтів у challenge
	Tau = 60

	// Beta - межа для перевірки
	Beta = Tau * Eta

	// Gamma1 - параметр для маскування
	Gamma1 = 1 << 19 // 2^19

	// Gamma2 - параметр для розкладу
	Gamma2 = (Q - 1) / 32

	// Omega - максимальна кількість одиниць у hint
	Omega = 75

	// PublicKeySize - розмір відкритого ключа
	PublicKeySize = 32 + K*N*10/8 // 2592 байт

	// PrivateKeySize - розмір закритого ключа
	PrivateKeySize = 32 + 32 + 64 + K*N*Eta*2/8 + L*N*Eta*2/8 + K*N*13/8 // ~4880 байт

	// SignatureSize - розмір підпису
	SignatureSize = 32 + L*N*20/8 + Omega + K // ~4627 байт

	// SeedBytes - розмір seed
	SeedBytes = 32

	// CRHBytes - розмір collision-resistant hash
	CRHBytes = 64
)

// Poly представляє поліном в Z_Q[X]/(X^N + 1)
type Poly [N]int32

// PolyVecK представляє вектор з K поліномів
type PolyVecK [K]Poly

// PolyVecL представляє вектор з L поліномів
type PolyVecL [L]Poly

// PublicKey представляє відкритий ключ Сокола
type PublicKey struct {
	Rho [SeedBytes]byte // seed для матриці A
	T1  PolyVecK        // t1 = верхні біти t
}

// PrivateKey представляє закритий ключ Сокола
type PrivateKey struct {
	Rho [SeedBytes]byte // seed для матриці A
	Key [SeedBytes]byte // seed для підпису
	Tr  [CRHBytes]byte  // H(pk)
	S1  PolyVecL        // секретний вектор s1
	S2  PolyVecK        // секретний вектор s2
	T0  PolyVecK        // t0 = нижні біти t
}

// Signature представляє цифровий підпис
type Signature struct {
	C    [SeedBytes]byte // challenge hash
	Z    PolyVecL        // відповідь z
	Hint [K][]int        // hint для відновлення
}

// KeyPair представляє пару ключів
type KeyPair struct {
	Public  *PublicKey
	Private *PrivateKey
}

// reduce приводить коефіцієнт до діапазону [0, Q)
func reduce(x int32) int32 {
	r := x % Q
	if r < 0 {
		r += Q
	}
	return r
}

// caddq додає Q якщо від'ємне
func caddq(x int32) int32 {
	if x < 0 {
		return x + Q
	}
	return x
}

// freeze приводить до діапазону (-Q/2, Q/2]
func freeze(x int32) int32 {
	r := reduce(x)
	if r > Q/2 {
		r -= Q
	}
	return r
}

// power2Round розкладає a = a1*2^D + a0
func power2Round(a int32) (a0, a1 int32) {
	const D = 13
	a1 = (a + (1 << (D - 1)) - 1) >> D
	a0 = a - (a1 << D)
	return
}

// decompose розкладає a = a1*Alpha + a0
func decompose(a int32) (a0, a1 int32) {
	a1 = (a + 127) >> 7
	if a1 > 43 {
		a1 = (a1*11275 + (1 << 23)) >> 24
	} else {
		a1 = (a1*1025 + (1 << 21)) >> 22
	}
	a1 &= 15
	a0 = a - a1*2*Gamma2
	a0 -= (((Q-1)/2 - a0) >> 31) & Q
	return
}

// makeHint створює hint для h
func makeHint(a0, a1 int32) int {
	if a0 > Gamma2 || a0 < -Gamma2 || (a0 == -Gamma2 && a1 != 0) {
		return 1
	}
	return 0
}

// useHint використовує hint для відновлення
func useHint(a int32, hint int) int32 {
	a0, a1 := decompose(a)
	if hint == 0 {
		return a1
	}
	if a0 > 0 {
		return (a1 + 1) & 15
	}
	return (a1 - 1) & 15
}

// infNorm обчислює нескінченну норму поліному
func (p *Poly) infNorm() int32 {
	var max int32
	for i := 0; i < N; i++ {
		t := p[i]
		if t < 0 {
			t = -t
		}
		if t > max {
			max = t
		}
	}
	return max
}

// checkNorm перевіряє, чи норма <= bound
func (p *Poly) checkNorm(bound int32) bool {
	return p.infNorm() <= bound
}
