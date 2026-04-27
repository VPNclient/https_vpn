// Package malva реалізує KEM Мальва на базі Module-LWE.
// Аналог ML-KEM (Kyber-1024) для постквантової стійкості Category 5.
package malva

// Параметри Мальви-1024 (аналог Kyber-1024)
const (
	// N - розмірність поліному (степінь x^N + 1)
	N = 256

	// K - розмір модуля (кількість поліномів у векторі)
	K = 4

	// Q - модуль для коефіцієнтів поліномів
	Q = 3329

	// Eta1 - параметр розподілу для секретного ключа
	Eta1 = 2

	// Eta2 - параметр розподілу для шуму
	Eta2 = 2

	// Du - кількість біт для стиснення u
	Du = 11

	// Dv - кількість біт для стиснення v
	Dv = 5

	// PublicKeySize - розмір відкритого ключа в байтах
	// pk = (ρ || t) де t = 12*K*N/8 + 32
	PublicKeySize = 12*K*N/8 + 32 // 1568 байт

	// PrivateKeySize - розмір закритого ключа в байтах
	// sk = (s || pk || H(pk) || z)
	PrivateKeySize = 12*K*N/8 + PublicKeySize + 32 + 32 // 3168 байт

	// CiphertextSize - розмір шифротексту в байтах
	// ct = (c1 || c2) де c1 = Du*K*N/8, c2 = Dv*N/8
	CiphertextSize = Du*K*N/8 + Dv*N/8 // 1568 байт

	// SharedSecretSize - розмір спільного секрету в байтах
	SharedSecretSize = 32

	// SymBytes - розмір симетричних ключів/хешів
	SymBytes = 32
)

// Poly представляє поліном в Z_Q[X]/(X^N + 1)
type Poly [N]int16

// PolyVec представляє вектор поліномів
type PolyVec [K]Poly

// PublicKey представляє відкритий ключ Мальви
type PublicKey struct {
	T   PolyVec        // t = A*s + e (в NTT домені)
	Rho [SymBytes]byte // seed для генерації матриці A
}

// PrivateKey представляє закритий ключ Мальви
type PrivateKey struct {
	S      PolyVec            // секретний вектор s (в NTT домені)
	Pk     *PublicKey         // відкритий ключ
	HPk    [SymBytes]byte     // H(pk)
	Z      [SymBytes]byte     // випадкове значення для implicit rejection
	pkData [PublicKeySize]byte // серіалізований pk
}

// Ciphertext представляє шифротекст KEM
type Ciphertext struct {
	U PolyVec // стиснутий вектор u
	V Poly    // стиснутий поліном v
}

// KeyPair представляє пару ключів
type KeyPair struct {
	Public  *PublicKey
	Private *PrivateKey
}

// Bytes серіалізує відкритий ключ
func (pk *PublicKey) Bytes() []byte {
	result := make([]byte, PublicKeySize)
	offset := 0

	// Серіалізуємо t
	for i := 0; i < K; i++ {
		for j := 0; j < N/2; j++ {
			t0 := uint16(pk.T[i][2*j])
			t1 := uint16(pk.T[i][2*j+1])
			// 12 біт на коефіцієнт
			result[offset] = byte(t0)
			result[offset+1] = byte((t0>>8)|(t1<<4))
			result[offset+2] = byte(t1 >> 4)
			offset += 3
		}
	}

	// Додаємо rho
	copy(result[offset:], pk.Rho[:])

	return result
}

// Bytes серіалізує закритий ключ
func (sk *PrivateKey) Bytes() []byte {
	result := make([]byte, PrivateKeySize)
	offset := 0

	// Серіалізуємо s
	for i := 0; i < K; i++ {
		for j := 0; j < N/2; j++ {
			s0 := uint16(sk.S[i][2*j])
			s1 := uint16(sk.S[i][2*j+1])
			result[offset] = byte(s0)
			result[offset+1] = byte((s0>>8)|(s1<<4))
			result[offset+2] = byte(s1 >> 4)
			offset += 3
		}
	}

	// Додаємо pk
	copy(result[offset:], sk.pkData[:])
	offset += PublicKeySize

	// Додаємо H(pk)
	copy(result[offset:], sk.HPk[:])
	offset += SymBytes

	// Додаємо z
	copy(result[offset:], sk.Z[:])

	return result
}

// Bytes серіалізує шифротекст
func (ct *Ciphertext) Bytes() []byte {
	result := make([]byte, CiphertextSize)
	offset := 0

	// Стиснення u (Du біт на коефіцієнт)
	for i := 0; i < K; i++ {
		for j := 0; j < N; j += 8 {
			// Пакуємо 8 коефіцієнтів у 11 байт (Du=11)
			for k := 0; k < 8 && j+k < N; k++ {
				// Спрощена серіалізація
				u := compress(ct.U[i][j+k], Du)
				bitOffset := k * Du
				byteOffset := bitOffset / 8
				bitPos := bitOffset % 8

				result[offset+byteOffset] |= byte(u << bitPos)
				if bitPos+Du > 8 {
					result[offset+byteOffset+1] |= byte(u >> (8 - bitPos))
				}
				if bitPos+Du > 16 {
					result[offset+byteOffset+2] |= byte(u >> (16 - bitPos))
				}
			}
			offset += (8 * Du) / 8
		}
	}

	// Стиснення v (Dv біт на коефіцієнт)
	for j := 0; j < N; j += 8 {
		for k := 0; k < 8 && j+k < N; k++ {
			v := compress(ct.V[j+k], Dv)
			bitOffset := k * Dv
			byteOffset := bitOffset / 8
			bitPos := bitOffset % 8

			result[offset+byteOffset] |= byte(v << bitPos)
			if bitPos+Dv > 8 {
				result[offset+byteOffset+1] |= byte(v >> (8 - bitPos))
			}
		}
		offset += (8 * Dv) / 8
	}

	return result
}

// compress стискає значення до d біт
func compress(x int16, d int) uint16 {
	// compress_d(x) = round((2^d / Q) * x) mod 2^d
	if x < 0 {
		x += Q
	}
	return uint16((uint32(x)<<d + Q/2) / Q) & ((1 << d) - 1)
}

// decompress розтискає значення з d біт
func decompress(x uint16, d int) int16 {
	// decompress_d(x) = round((Q / 2^d) * x)
	return int16((uint32(x)*Q + (1 << (d - 1))) >> d)
}

// reduce приводить коефіцієнт до діапазону [0, Q)
func reduce(x int32) int16 {
	r := x % Q
	if r < 0 {
		r += Q
	}
	return int16(r)
}

// barrettReduce швидка редукція за модулем Q використовуючи Barrett reduction
func barrettReduce(a int16) int16 {
	const v = ((1 << 26) + Q/2) / Q
	t := int32(v) * int32(a) >> 26
	t = int32(a) - t*Q
	if t >= Q {
		t -= Q
	}
	if t < 0 {
		t += Q
	}
	return int16(t)
}

// montgomeryReduce Montgomery reduction для результату множення
func montgomeryReduce(a int32) int16 {
	const qinv int32 = 62209 // Q^(-1) mod 2^16
	t := int16(int32(int16(a)) * qinv)
	return int16((a - int32(t)*Q) >> 16)
}
