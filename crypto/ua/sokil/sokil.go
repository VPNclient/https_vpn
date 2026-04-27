package sokil

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/sha3"
)

// Помилки
var (
	ErrInvalidPublicKey  = errors.New("sokil: недійсний відкритий ключ")
	ErrInvalidPrivateKey = errors.New("sokil: недійсний закритий ключ")
	ErrInvalidSignature  = errors.New("sokil: недійсний підпис")
	ErrSignatureFailed   = errors.New("sokil: не вдалося створити підпис")
)

// GenerateKey генерує нову пару ключів
func GenerateKey(random io.Reader) (*KeyPair, error) {
	if random == nil {
		random = rand.Reader
	}

	// Генеруємо випадковий seed
	var seed [SeedBytes]byte
	if _, err := io.ReadFull(random, seed[:]); err != nil {
		return nil, err
	}

	// Розширюємо seed: (ρ, ρ', K) = H(seed)
	h := sha3.NewShake256()
	h.Write(seed[:])

	var rho [SeedBytes]byte
	var rhoPrime [CRHBytes]byte
	var key [SeedBytes]byte

	h.Read(rho[:])
	h.Read(rhoPrime[:])
	h.Read(key[:])

	// Генеруємо матрицю A з ρ
	var A [K]PolyVecL
	expandA(&A, &rho)

	// Генеруємо секретні вектори s1, s2 з ρ'
	var s1 PolyVecL
	var s2 PolyVecK

	for i := 0; i < L; i++ {
		sampleEta(&s1[i], &rhoPrime, uint16(i))
	}
	for i := 0; i < K; i++ {
		sampleEta(&s2[i], &rhoPrime, uint16(L+i))
	}

	// NTT(s1)
	var s1Hat PolyVecL
	for i := 0; i < L; i++ {
		s1Hat[i] = s1[i]
		ntt(&s1Hat[i])
	}

	// t = A*s1 + s2
	var t PolyVecK
	for i := 0; i < K; i++ {
		// A[i] * s1Hat
		var tmp Poly
		for j := 0; j < L; j++ {
			var prod Poly
			polyMul(&prod, &A[i][j], &s1Hat[j])
			polyAdd(&tmp, &tmp, &prod)
		}
		invNtt(&tmp)
		polyAdd(&t[i], &tmp, &s2[i])
		polyReduce(&t[i])
	}

	// Розкладаємо t = t1*2^D + t0
	var t0, t1 PolyVecK
	for i := 0; i < K; i++ {
		for j := 0; j < N; j++ {
			t0[i][j], t1[i][j] = power2Round(t[i][j])
		}
	}

	// Обчислюємо tr = H(ρ || t1)
	pk := &PublicKey{
		Rho: rho,
		T1:  t1,
	}
	pkBytes := pk.Bytes()

	var tr [CRHBytes]byte
	h2 := sha3.NewShake256()
	h2.Write(pkBytes)
	h2.Read(tr[:])

	sk := &PrivateKey{
		Rho: rho,
		Key: key,
		Tr:  tr,
		S1:  s1,
		S2:  s2,
		T0:  t0,
	}

	return &KeyPair{
		Public:  pk,
		Private: sk,
	}, nil
}

// Sign створює підпис повідомлення
func Sign(sk *PrivateKey, message []byte) ([]byte, error) {
	// Генеруємо матрицю A
	var A [K]PolyVecL
	expandA(&A, &sk.Rho)

	// NTT(s1), NTT(s2), NTT(t0)
	var s1Hat PolyVecL
	var s2Hat, t0Hat PolyVecK

	for i := 0; i < L; i++ {
		s1Hat[i] = sk.S1[i]
		ntt(&s1Hat[i])
	}
	for i := 0; i < K; i++ {
		s2Hat[i] = sk.S2[i]
		ntt(&s2Hat[i])
		t0Hat[i] = sk.T0[i]
		ntt(&t0Hat[i])
	}

	// μ = H(tr || M)
	var mu [CRHBytes]byte
	h := sha3.NewShake256()
	h.Write(sk.Tr[:])
	h.Write(message)
	h.Read(mu[:])

	// Генеруємо випадковий nonce для rejection sampling
	var rhoPrime [CRHBytes]byte
	h2 := sha3.NewShake256()
	h2.Write(sk.Key[:])
	h2.Write(mu[:])
	h2.Read(rhoPrime[:])

	var kappa uint16 = 0
	const maxAttempts = 1000

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Генеруємо y з рівномірного розподілу [-Gamma1+1, Gamma1]
		var y PolyVecL
		for i := 0; i < L; i++ {
			sampleGamma1(&y[i], &rhoPrime, kappa+uint16(i))
		}

		// NTT(y)
		var yHat PolyVecL
		for i := 0; i < L; i++ {
			yHat[i] = y[i]
			ntt(&yHat[i])
		}

		// w = A*y
		var w PolyVecK
		for i := 0; i < K; i++ {
			var tmp Poly
			for j := 0; j < L; j++ {
				var prod Poly
				polyMul(&prod, &A[i][j], &yHat[j])
				polyAdd(&tmp, &tmp, &prod)
			}
			invNtt(&tmp)
			w[i] = tmp
			polyReduce(&w[i])
		}

		// Розкладаємо w
		var w1 PolyVecK
		for i := 0; i < K; i++ {
			for j := 0; j < N; j++ {
				_, w1[i][j] = decompose(w[i][j])
			}
		}

		// c~ = H(μ || w1)
		var cTilde [SeedBytes]byte
		h3 := sha3.NewShake256()
		h3.Write(mu[:])
		for i := 0; i < K; i++ {
			for j := 0; j < N; j++ {
				h3.Write([]byte{byte(w1[i][j])})
			}
		}
		h3.Read(cTilde[:])

		// c = SampleInBall(c~)
		var c Poly
		sampleChallenge(&c, &cTilde)

		// NTT(c)
		var cHat Poly
		cHat = c
		ntt(&cHat)

		// z = y + c*s1
		var z PolyVecL
		for i := 0; i < L; i++ {
			var cs1 Poly
			polyMul(&cs1, &cHat, &s1Hat[i])
			invNtt(&cs1)
			polyAdd(&z[i], &y[i], &cs1)
			polyReduce(&z[i])
		}

		// Перевіряємо норму z
		validZ := true
		for i := 0; i < L; i++ {
			if !z[i].checkNorm(Gamma1 - Beta) {
				validZ = false
				break
			}
		}
		if !validZ {
			kappa += uint16(L)
			continue
		}

		// r0 = w - c*s2, перевіряємо норму
		var r0 PolyVecK
		valid := true
		for i := 0; i < K; i++ {
			var cs2 Poly
			polyMul(&cs2, &cHat, &s2Hat[i])
			invNtt(&cs2)
			polySub(&r0[i], &w[i], &cs2)
			polyReduce(&r0[i])
			if !r0[i].checkNorm(Gamma2 - Beta) {
				valid = false
				break
			}
		}
		if !valid {
			kappa += uint16(L)
			continue
		}

		// Створюємо hint
		var ct0 PolyVecK
		for i := 0; i < K; i++ {
			polyMul(&ct0[i], &cHat, &t0Hat[i])
			invNtt(&ct0[i])
		}

		var hint [K][]int
		hintCount := 0
		for i := 0; i < K; i++ {
			hint[i] = make([]int, 0)
			for j := 0; j < N; j++ {
				h := makeHint(-ct0[i][j], r0[i][j]+ct0[i][j])
				if h != 0 {
					hint[i] = append(hint[i], j)
					hintCount++
				}
			}
		}

		if hintCount > Omega {
			kappa += uint16(L)
			continue
		}

		// Успішно створили підпис
		sig := &Signature{
			C:    cTilde,
			Z:    z,
			Hint: hint,
		}

		return sig.Bytes(), nil
	}

	return nil, ErrSignatureFailed
}

// Verify перевіряє підпис
func Verify(pk *PublicKey, message, signature []byte) bool {
	sig, err := parseSignature(signature)
	if err != nil {
		return false
	}

	// Генеруємо матрицю A
	var A [K]PolyVecL
	expandA(&A, &pk.Rho)

	// μ = H(H(pk) || M)
	pkBytes := pk.Bytes()
	var tr [CRHBytes]byte
	h := sha3.NewShake256()
	h.Write(pkBytes)
	h.Read(tr[:])

	var mu [CRHBytes]byte
	h2 := sha3.NewShake256()
	h2.Write(tr[:])
	h2.Write(message)
	h2.Read(mu[:])

	// c = SampleInBall(c~)
	var c Poly
	sampleChallenge(&c, &sig.C)

	// NTT(c), NTT(z)
	var cHat Poly
	cHat = c
	ntt(&cHat)

	var zHat PolyVecL
	for i := 0; i < L; i++ {
		zHat[i] = sig.Z[i]
		ntt(&zHat[i])
	}

	// Перевіряємо норму z
	for i := 0; i < L; i++ {
		if !sig.Z[i].checkNorm(Gamma1 - Beta) {
			return false
		}
	}

	// w'1 = UseHint(A*z - c*t1*2^D, h)
	var t1Hat PolyVecK
	for i := 0; i < K; i++ {
		// t1 * 2^D
		for j := 0; j < N; j++ {
			t1Hat[i][j] = pk.T1[i][j] << 13
		}
		ntt(&t1Hat[i])
	}

	var w1Prime PolyVecK
	for i := 0; i < K; i++ {
		// A[i] * z
		var tmp Poly
		for j := 0; j < L; j++ {
			var prod Poly
			polyMul(&prod, &A[i][j], &zHat[j])
			polyAdd(&tmp, &tmp, &prod)
		}

		// - c * t1 * 2^D
		var ct1 Poly
		polyMul(&ct1, &cHat, &t1Hat[i])
		polySub(&tmp, &tmp, &ct1)
		invNtt(&tmp)
		polyReduce(&tmp)

		// UseHint
		for j := 0; j < N; j++ {
			hint := 0
			for _, idx := range sig.Hint[i] {
				if idx == j {
					hint = 1
					break
				}
			}
			w1Prime[i][j] = useHint(tmp[j], hint)
		}
	}

	// c'~ = H(μ || w'1)
	var cTildePrime [SeedBytes]byte
	h3 := sha3.NewShake256()
	h3.Write(mu[:])
	for i := 0; i < K; i++ {
		for j := 0; j < N; j++ {
			h3.Write([]byte{byte(w1Prime[i][j])})
		}
	}
	h3.Read(cTildePrime[:])

	// Перевіряємо c~ == c'~
	for i := 0; i < SeedBytes; i++ {
		if sig.C[i] != cTildePrime[i] {
			return false
		}
	}

	return true
}

// Bytes серіалізує відкритий ключ
func (pk *PublicKey) Bytes() []byte {
	result := make([]byte, 0, PublicKeySize)
	result = append(result, pk.Rho[:]...)

	// Серіалізуємо t1 (10 біт на коефіцієнт)
	for i := 0; i < K; i++ {
		for j := 0; j < N; j += 4 {
			// 4 коефіцієнти по 10 біт = 5 байт
			t0 := uint32(pk.T1[i][j]) & 0x3FF
			t1 := uint32(pk.T1[i][j+1]) & 0x3FF
			t2 := uint32(pk.T1[i][j+2]) & 0x3FF
			t3 := uint32(pk.T1[i][j+3]) & 0x3FF

			result = append(result, byte(t0))
			result = append(result, byte((t0>>8)|(t1<<2)))
			result = append(result, byte((t1>>6)|(t2<<4)))
			result = append(result, byte((t2>>4)|(t3<<6)))
			result = append(result, byte(t3>>2))
		}
	}

	return result
}

// Bytes серіалізує підпис
func (sig *Signature) Bytes() []byte {
	result := make([]byte, 0, SignatureSize)
	result = append(result, sig.C[:]...)

	// Серіалізуємо z (20 біт на коефіцієнт)
	for i := 0; i < L; i++ {
		for j := 0; j < N; j += 4 {
			z0 := uint32(sig.Z[i][j]+Gamma1) & 0xFFFFF
			z1 := uint32(sig.Z[i][j+1]+Gamma1) & 0xFFFFF
			z2 := uint32(sig.Z[i][j+2]+Gamma1) & 0xFFFFF
			z3 := uint32(sig.Z[i][j+3]+Gamma1) & 0xFFFFF

			result = append(result, byte(z0))
			result = append(result, byte(z0>>8))
			result = append(result, byte((z0>>16)|(z1<<4)))
			result = append(result, byte(z1>>4))
			result = append(result, byte(z1>>12))
			result = append(result, byte(z2))
			result = append(result, byte(z2>>8))
			result = append(result, byte((z2>>16)|(z3<<4)))
			result = append(result, byte(z3>>4))
			result = append(result, byte(z3>>12))
		}
	}

	// Серіалізуємо hint
	for i := 0; i < K; i++ {
		for _, idx := range sig.Hint[i] {
			result = append(result, byte(idx))
		}
	}
	// Padding
	for len(result) < SignatureSize {
		result = append(result, 0)
	}

	return result
}

// parseSignature розбирає підпис
func parseSignature(data []byte) (*Signature, error) {
	if len(data) < SeedBytes {
		return nil, ErrInvalidSignature
	}

	sig := &Signature{}
	copy(sig.C[:], data[:SeedBytes])

	// Спрощений парсинг
	offset := SeedBytes
	for i := 0; i < L; i++ {
		for j := 0; j < N && offset+2 < len(data); j++ {
			sig.Z[i][j] = int32(data[offset]) | int32(data[offset+1])<<8
			sig.Z[i][j] -= Gamma1
			offset += 2
		}
	}

	// Парсинг hint (спрощено)
	for i := 0; i < K; i++ {
		sig.Hint[i] = make([]int, 0)
	}

	return sig, nil
}

// expandA генерує матрицю A
func expandA(A *[K]PolyVecL, rho *[SeedBytes]byte) {
	for i := 0; i < K; i++ {
		for j := 0; j < L; j++ {
			sampleUniform(&A[i][j], rho, byte(i), byte(j))
		}
	}
}

// sampleUniform генерує рівномірно розподілений поліном
func sampleUniform(p *Poly, rho *[SeedBytes]byte, i, j byte) {
	xof := sha3.NewShake128()
	xof.Write(rho[:])
	xof.Write([]byte{j, i})

	var buf [3]byte
	k := 0
	for k < N {
		xof.Read(buf[:])
		d := uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]&0x7F) << 16)
		if d < Q {
			p[k] = int32(d)
			k++
		}
	}
}

// sampleEta генерує поліном з малими коефіцієнтами
func sampleEta(p *Poly, seed *[CRHBytes]byte, nonce uint16) {
	xof := sha3.NewShake256()
	xof.Write(seed[:])
	xof.Write([]byte{byte(nonce), byte(nonce >> 8)})

	buf := make([]byte, N)
	xof.Read(buf)

	for i := 0; i < N; i++ {
		t0 := buf[i] & 0x0F
		t1 := buf[i] >> 4
		p[i] = int32(t0) - int32(t1)
		if p[i] < -Eta {
			p[i] = -Eta
		}
		if p[i] > Eta {
			p[i] = Eta
		}
	}
}

// sampleGamma1 генерує поліном з рівномірним розподілом
func sampleGamma1(p *Poly, seed *[CRHBytes]byte, nonce uint16) {
	xof := sha3.NewShake256()
	xof.Write(seed[:])
	xof.Write([]byte{byte(nonce), byte(nonce >> 8)})

	buf := make([]byte, N*3)
	xof.Read(buf)

	for i := 0; i < N; i++ {
		t := uint32(buf[3*i]) | (uint32(buf[3*i+1]) << 8) | (uint32(buf[3*i+2]&0x0F) << 16)
		p[i] = Gamma1 - int32(t%(2*Gamma1))
	}
}

// sampleChallenge генерує challenge поліном
func sampleChallenge(c *Poly, seed *[SeedBytes]byte) {
	xof := sha3.NewShake256()
	xof.Write(seed[:])

	var signs uint64
	var buf [8]byte
	xof.Read(buf[:])
	for i := 0; i < 8; i++ {
		signs |= uint64(buf[i]) << (8 * i)
	}

	for i := 0; i < N; i++ {
		c[i] = 0
	}

	var pos [1]byte
	for i := N - Tau; i < N; i++ {
		for {
			xof.Read(pos[:])
			if int(pos[0]) <= i {
				break
			}
		}
		c[i] = c[pos[0]]
		if signs&1 != 0 {
			c[pos[0]] = -1
		} else {
			c[pos[0]] = 1
		}
		signs >>= 1
	}
}

// NTT операції (спрощені)
func ntt(p *Poly) {
	// Спрощена реалізація NTT для Dilithium
	// В реальності потрібна повна реалізація з правильними zetas
	for i := 0; i < N; i++ {
		p[i] = reduce(p[i])
	}
}

func invNtt(p *Poly) {
	for i := 0; i < N; i++ {
		p[i] = reduce(p[i])
	}
}

// Операції з поліномами
func polyAdd(c, a, b *Poly) {
	for i := 0; i < N; i++ {
		c[i] = a[i] + b[i]
	}
}

func polySub(c, a, b *Poly) {
	for i := 0; i < N; i++ {
		c[i] = a[i] - b[i]
	}
}

func polyMul(c, a, b *Poly) {
	// Спрощене множення (для повної реалізації потрібен NTT)
	for i := 0; i < N; i++ {
		c[i] = reduce(a[i] * b[i])
	}
}

func polyReduce(p *Poly) {
	for i := 0; i < N; i++ {
		p[i] = reduce(p[i])
	}
}
