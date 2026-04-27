package malva

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/sha3"
)

// Помилки
var (
	ErrInvalidPublicKey  = errors.New("malva: недійсний відкритий ключ")
	ErrInvalidPrivateKey = errors.New("malva: недійсний закритий ключ")
	ErrInvalidCiphertext = errors.New("malva: недійсний шифротекст")
	ErrDecapsulation     = errors.New("malva: помилка декапсуляції")
)

// GenerateKey генерує нову пару ключів
func GenerateKey(random io.Reader) (*KeyPair, error) {
	if random == nil {
		random = rand.Reader
	}

	// Генеруємо випадкові значення
	var d [SymBytes]byte
	var z [SymBytes]byte

	if _, err := io.ReadFull(random, d[:]); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(random, z[:]); err != nil {
		return nil, err
	}

	// G(d) = (ρ, σ)
	g := sha3.New512()
	g.Write(d[:])
	rhoSigma := g.Sum(nil)

	var rho [SymBytes]byte
	var sigma [SymBytes]byte
	copy(rho[:], rhoSigma[:32])
	copy(sigma[:], rhoSigma[32:])

	// Генеруємо матрицю A з ρ
	var A [K]PolyVec
	expandA(&A, &rho)

	// Генеруємо секретний вектор s з σ
	var s PolyVec
	var e PolyVec

	for i := 0; i < K; i++ {
		sampleNoise(&s[i], &sigma, byte(i), Eta1)
		sampleNoise(&e[i], &sigma, byte(K+i), Eta1)
	}

	// NTT(s) та NTT(e)
	s.NTT()
	e.NTT()

	// t = A*s + e (в NTT домені)
	t := MatVecMul(&A, &s, false)
	t.Add(t, &e)
	t.Reduce()

	// Створюємо ключі
	pk := &PublicKey{
		T:   *t,
		Rho: rho,
	}

	sk := &PrivateKey{
		S:  s,
		Pk: pk,
		Z:  z,
	}

	// H(pk)
	pkBytes := pk.Bytes()
	copy(sk.pkData[:], pkBytes)
	h := sha3.Sum256(pkBytes)
	copy(sk.HPk[:], h[:])

	return &KeyPair{
		Public:  pk,
		Private: sk,
	}, nil
}

// Encapsulate створює шифротекст та спільний секрет
func Encapsulate(pk *PublicKey, random io.Reader) (ciphertext []byte, sharedSecret []byte, err error) {
	if random == nil {
		random = rand.Reader
	}

	// Генеруємо випадкове повідомлення m
	var m [SymBytes]byte
	if _, err := io.ReadFull(random, m[:]); err != nil {
		return nil, nil, err
	}

	// Серіалізуємо pk
	pkBytes := pk.Bytes()

	// H(pk)
	hpk := sha3.Sum256(pkBytes)

	// (K̄, r) = G(m || H(pk))
	g := sha3.New512()
	g.Write(m[:])
	g.Write(hpk[:])
	kr := g.Sum(nil)

	var kBar [SymBytes]byte
	var r [SymBytes]byte
	copy(kBar[:], kr[:32])
	copy(r[:], kr[32:])

	// Шифруємо
	ct, err := encrypt(pk, &m, &r)
	if err != nil {
		return nil, nil, err
	}

	ctBytes := ct.Bytes()

	// K = KDF(K̄ || H(c))
	hc := sha3.Sum256(ctBytes)
	kdf := sha3.NewShake256()
	kdf.Write(kBar[:])
	kdf.Write(hc[:])

	sharedSecret = make([]byte, SharedSecretSize)
	kdf.Read(sharedSecret)

	return ctBytes, sharedSecret, nil
}

// Decapsulate відновлює спільний секрет з шифротексту
func Decapsulate(sk *PrivateKey, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) != CiphertextSize {
		return nil, ErrInvalidCiphertext
	}

	// Розпаковуємо шифротекст
	ct, err := parseCiphertext(ciphertext)
	if err != nil {
		return nil, err
	}

	// Розшифровуємо
	mPrime := decrypt(sk, ct)

	// (K̄', r') = G(m' || H(pk))
	g := sha3.New512()
	g.Write(mPrime[:])
	g.Write(sk.HPk[:])
	krPrime := g.Sum(nil)

	var kBarPrime [SymBytes]byte
	var rPrime [SymBytes]byte
	copy(kBarPrime[:], krPrime[:32])
	copy(rPrime[:], krPrime[32:])

	// Перешифровуємо для перевірки
	ctPrime, err := encrypt(sk.Pk, mPrime, &rPrime)
	if err != nil {
		return nil, err
	}

	// Перевіряємо шифротекст (implicit rejection)
	ctPrimeBytes := ctPrime.Bytes()
	valid := constantTimeCompare(ciphertext, ctPrimeBytes)

	// H(c)
	hc := sha3.Sum256(ciphertext)

	// Вибираємо K̄ або z залежно від результату
	kdf := sha3.NewShake256()
	if valid == 1 {
		kdf.Write(kBarPrime[:])
	} else {
		kdf.Write(sk.Z[:])
	}
	kdf.Write(hc[:])

	sharedSecret := make([]byte, SharedSecretSize)
	kdf.Read(sharedSecret)

	return sharedSecret, nil
}

// encrypt виконує внутрішнє шифрування
func encrypt(pk *PublicKey, m *[SymBytes]byte, r *[SymBytes]byte) (*Ciphertext, error) {
	// Генеруємо матрицю A
	var A [K]PolyVec
	expandA(&A, &pk.Rho)

	// Генеруємо вектори з r
	var rVec PolyVec
	var e1 PolyVec
	var e2 Poly

	for i := 0; i < K; i++ {
		sampleNoise(&rVec[i], r, byte(i), Eta1)
		sampleNoise(&e1[i], r, byte(K+i), Eta2)
	}
	sampleNoise(&e2, r, byte(2*K), Eta2)

	// NTT(r)
	rVec.NTT()

	// u = A^T * r + e1
	u := MatVecMul(&A, &rVec, true)
	u.InvNTT()
	u.Add(u, &e1)
	u.Reduce()

	// v = t^T * r + e2 + encode(m)
	tNTT := pk.T
	v := InnerProduct(&tNTT, &rVec)
	v.InvNTT()
	v.Add(v, &e2)

	// Encode message
	var msgPoly Poly
	for i := 0; i < N && i < SymBytes*8; i++ {
		bit := (m[i/8] >> (i % 8)) & 1
		msgPoly[i] = int16(bit) * ((Q + 1) / 2)
	}
	v.Add(v, &msgPoly)
	v.reduce()

	return &Ciphertext{
		U: *u,
		V: *v,
	}, nil
}

// decrypt виконує внутрішнє розшифрування
func decrypt(sk *PrivateKey, ct *Ciphertext) *[SymBytes]byte {
	// NTT(u)
	uNTT := ct.U
	uNTT.NTT()

	// v - s^T * u
	sTimesU := InnerProduct(&sk.S, &uNTT)
	sTimesU.InvNTT()

	var mp Poly
	mp.Sub(&ct.V, sTimesU)
	mp.reduce()

	// Decode message
	var m [SymBytes]byte
	for i := 0; i < N && i < SymBytes*8; i++ {
		// Якщо коефіцієнт ближче до Q/2, то біт = 1
		coef := mp[i]
		if coef < 0 {
			coef += Q
		}
		// Порівнюємо відстань до 0 та Q/2
		dist0 := coef
		distHalf := coef - Q/2
		if distHalf < 0 {
			distHalf = -distHalf
		}
		if int(dist0) > int(distHalf) {
			m[i/8] |= 1 << (i % 8)
		}
	}

	return &m
}

// expandA генерує матрицю A з seed ρ
func expandA(A *[K]PolyVec, rho *[SymBytes]byte) {
	for i := 0; i < K; i++ {
		for j := 0; j < K; j++ {
			// A[i][j] = Parse(XOF(ρ || i || j))
			xof := sha3.NewShake128()
			xof.Write(rho[:])
			xof.Write([]byte{byte(j), byte(i)})

			var buf [3]byte
			k := 0
			for k < N {
				xof.Read(buf[:])
				d1 := uint16(buf[0]) | (uint16(buf[1]&0x0F) << 8)
				d2 := uint16(buf[1]>>4) | (uint16(buf[2]) << 4)

				if d1 < Q {
					A[i][j][k] = int16(d1)
					k++
				}
				if k < N && d2 < Q {
					A[i][j][k] = int16(d2)
					k++
				}
			}
		}
	}
}

// sampleNoise генерує поліном з центрованого біноміального розподілу
func sampleNoise(p *Poly, seed *[SymBytes]byte, nonce byte, eta int) {
	// CBD_η (Centered Binomial Distribution)
	prf := sha3.NewShake256()
	prf.Write(seed[:])
	prf.Write([]byte{nonce})

	buf := make([]byte, eta*N/4)
	prf.Read(buf)

	idx := 0
	for i := 0; i < N; i++ {
		var a, b int16
		for j := 0; j < eta; j++ {
			byteIdx := idx / 8
			bitIdx := idx % 8
			a += int16((buf[byteIdx] >> bitIdx) & 1)
			idx++
		}
		for j := 0; j < eta; j++ {
			byteIdx := idx / 8
			bitIdx := idx % 8
			b += int16((buf[byteIdx] >> bitIdx) & 1)
			idx++
		}
		p[i] = a - b
	}
}

// parseCiphertext розбирає байти шифротексту
func parseCiphertext(data []byte) (*Ciphertext, error) {
	if len(data) != CiphertextSize {
		return nil, ErrInvalidCiphertext
	}

	ct := &Ciphertext{}

	// Спрощений парсинг (потрібна повна реалізація)
	offset := 0

	// Розпаковуємо u
	for i := 0; i < K; i++ {
		for j := 0; j < N; j++ {
			// Спрощено: беремо 2 байти на коефіцієнт
			if offset+1 < len(data) {
				ct.U[i][j] = decompress(uint16(data[offset])|uint16(data[offset+1])<<8, Du)
				offset += 2
			}
		}
	}

	// Розпаковуємо v
	for j := 0; j < N && offset < len(data); j++ {
		ct.V[j] = decompress(uint16(data[offset]), Dv)
		offset++
	}

	return ct, nil
}

// constantTimeCompare порівнює два байтові зрізи за константний час
func constantTimeCompare(a, b []byte) int {
	if len(a) != len(b) {
		return 0
	}

	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}

	// Повертаємо 1 якщо v == 0, інакше 0
	return int((uint32(v)-1)>>31) & 1
}
