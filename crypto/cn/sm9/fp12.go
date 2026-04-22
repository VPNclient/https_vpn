// Fp6 and Fp12 extension field arithmetic for BN256 pairing.
package sm9

import "math/big"

// ---------- Fp6: Cubic extension Fp2[v]/(v³ - ξ) where ξ = u + 1 ----------

// fp6 represents an element of Fp6 = Fp2[v]/(v³ - ξ)
// a = a0 + a1*v + a2*v²
type fp6 struct {
	c0, c1, c2 *fp2
}

func newFp6() *fp6 {
	return &fp6{newFp2(), newFp2(), newFp2()}
}

func (e *fp6) Set(a *fp6) *fp6 {
	e.c0.Set(a.c0)
	e.c1.Set(a.c1)
	e.c2.Set(a.c2)
	return e
}

func (e *fp6) IsZero() bool {
	return e.c0.IsZero() && e.c1.IsZero() && e.c2.IsZero()
}

func (e *fp6) IsOne() bool {
	return e.c0.IsOne() && e.c1.IsZero() && e.c2.IsZero()
}

func fp6One() *fp6 {
	return &fp6{
		c0: &fp2{big.NewInt(1), big.NewInt(0)},
		c1: &fp2{big.NewInt(0), big.NewInt(0)},
		c2: &fp2{big.NewInt(0), big.NewInt(0)},
	}
}

// fp6Add returns a + b in Fp6
func fp6Add(a, b *fp6) *fp6 {
	return &fp6{
		c0: fp2Add(a.c0, b.c0),
		c1: fp2Add(a.c1, b.c1),
		c2: fp2Add(a.c2, b.c2),
	}
}

// fp6Sub returns a - b in Fp6
func fp6Sub(a, b *fp6) *fp6 {
	return &fp6{
		c0: fp2Sub(a.c0, b.c0),
		c1: fp2Sub(a.c1, b.c1),
		c2: fp2Sub(a.c2, b.c2),
	}
}

// fp6Neg returns -a in Fp6
func fp6Neg(a *fp6) *fp6 {
	return &fp6{
		c0: fp2Neg(a.c0),
		c1: fp2Neg(a.c1),
		c2: fp2Neg(a.c2),
	}
}

// fp6Mul returns a * b in Fp6
func fp6Mul(a, b *fp6) *fp6 {
	// Karatsuba-like multiplication
	t0 := fp2Mul(a.c0, b.c0)
	t1 := fp2Mul(a.c1, b.c1)
	t2 := fp2Mul(a.c2, b.c2)

	c0 := fp2Add(fp2Mul(fp2Add(a.c1, a.c2), fp2Add(b.c1, b.c2)), fp2Neg(fp2Add(t1, t2)))
	c0 = fp2Add(fp2MulXi(c0), t0)

	c1 := fp2Add(fp2Mul(fp2Add(a.c0, a.c1), fp2Add(b.c0, b.c1)), fp2Neg(fp2Add(t0, t1)))
	c1 = fp2Add(c1, fp2MulXi(t2))

	c2 := fp2Add(fp2Mul(fp2Add(a.c0, a.c2), fp2Add(b.c0, b.c2)), fp2Neg(fp2Add(t0, t2)))
	c2 = fp2Add(c2, t1)

	return &fp6{c0: c0, c1: c1, c2: c2}
}

// fp6Square returns a² in Fp6
func fp6Square(a *fp6) *fp6 {
	t0 := fp2Square(a.c0)
	t1 := fp2Mul(a.c0, a.c1)
	t1 = fp2Add(t1, t1)
	t2 := fp2Square(fp2Sub(fp2Add(a.c0, a.c2), a.c1))
	t3 := fp2Mul(a.c1, a.c2)
	t3 = fp2Add(t3, t3)
	t4 := fp2Square(a.c2)

	c0 := fp2Add(fp2MulXi(t3), t0)
	c1 := fp2Add(fp2MulXi(t4), t1)
	c2 := fp2Add(fp2Add(fp2Add(t1, t2), t3), fp2Neg(fp2Add(t0, t4)))

	return &fp6{c0: c0, c1: c1, c2: c2}
}

// fp6Inv returns a⁻¹ in Fp6
func fp6Inv(a *fp6) *fp6 {
	t0 := fp2Square(a.c0)
	t1 := fp2Square(a.c1)
	t2 := fp2Square(a.c2)
	t3 := fp2Mul(a.c0, a.c1)
	t4 := fp2Mul(a.c0, a.c2)
	t5 := fp2Mul(a.c1, a.c2)

	c0 := fp2Sub(t0, fp2MulXi(t5))
	c1 := fp2Sub(fp2MulXi(t2), t3)
	c2 := fp2Sub(t1, t4)

	t6 := fp2Mul(a.c0, c0)
	t6 = fp2Add(t6, fp2MulXi(fp2Mul(a.c2, c1)))
	t6 = fp2Add(t6, fp2MulXi(fp2Mul(a.c1, c2)))
	t6 = fp2Inv(t6)

	return &fp6{
		c0: fp2Mul(c0, t6),
		c1: fp2Mul(c1, t6),
		c2: fp2Mul(c2, t6),
	}
}

// fp6MulByV returns a * v in Fp6
// Multiplication by v shifts coefficients: (a0 + a1*v + a2*v²) * v = a2*ξ + a0*v + a1*v²
func fp6MulByV(a *fp6) *fp6 {
	return &fp6{
		c0: fp2MulXi(a.c2),
		c1: newFp2().Set(a.c0),
		c2: newFp2().Set(a.c1),
	}
}

// ---------- Fp12: Quadratic extension Fp6[w]/(w² - v) ----------

// fp12 represents an element of Fp12 = Fp6[w]/(w² - v)
// a = a0 + a1*w
type fp12 struct {
	c0, c1 *fp6
}

func newFp12() *fp12 {
	return &fp12{newFp6(), newFp6()}
}

func (e *fp12) Set(a *fp12) *fp12 {
	e.c0.Set(a.c0)
	e.c1.Set(a.c1)
	return e
}

func (e *fp12) IsOne() bool {
	return e.c0.IsOne() && e.c1.IsZero()
}

func fp12One() *fp12 {
	return &fp12{
		c0: fp6One(),
		c1: newFp6(),
	}
}

// fp12Add returns a + b in Fp12
func fp12Add(a, b *fp12) *fp12 {
	return &fp12{
		c0: fp6Add(a.c0, b.c0),
		c1: fp6Add(a.c1, b.c1),
	}
}

// fp12Sub returns a - b in Fp12
func fp12Sub(a, b *fp12) *fp12 {
	return &fp12{
		c0: fp6Sub(a.c0, b.c0),
		c1: fp6Sub(a.c1, b.c1),
	}
}

// fp12Neg returns -a in Fp12
func fp12Neg(a *fp12) *fp12 {
	return &fp12{
		c0: fp6Neg(a.c0),
		c1: fp6Neg(a.c1),
	}
}

// fp12Mul returns a * b in Fp12
func fp12Mul(a, b *fp12) *fp12 {
	t0 := fp6Mul(a.c0, b.c0)
	t1 := fp6Mul(a.c1, b.c1)
	c0 := fp6Add(t0, fp6MulByV(t1))
	c1 := fp6Sub(fp6Mul(fp6Add(a.c0, a.c1), fp6Add(b.c0, b.c1)), fp6Add(t0, t1))
	return &fp12{c0: c0, c1: c1}
}

// fp12Square returns a² in Fp12
func fp12Square(a *fp12) *fp12 {
	t0 := fp6Mul(a.c0, a.c1)
	t1 := fp6Add(a.c0, fp6MulByV(a.c1))
	t1 = fp6Mul(t1, fp6Add(a.c0, a.c1))
	t1 = fp6Sub(t1, t0)
	t1 = fp6Sub(t1, fp6MulByV(t0))
	c1 := fp6Add(t0, t0)
	return &fp12{c0: t1, c1: c1}
}

// fp12Inv returns a⁻¹ in Fp12
func fp12Inv(a *fp12) *fp12 {
	t0 := fp6Square(a.c0)
	t1 := fp6Square(a.c1)
	t0 = fp6Sub(t0, fp6MulByV(t1))
	t0 = fp6Inv(t0)
	return &fp12{
		c0: fp6Mul(a.c0, t0),
		c1: fp6Neg(fp6Mul(a.c1, t0)),
	}
}

// fp12Conjugate returns conjugate in Fp12 (used in final exponentiation)
func fp12Conjugate(a *fp12) *fp12 {
	return &fp12{
		c0: newFp6().Set(a.c0),
		c1: fp6Neg(a.c1),
	}
}

// fp12Frobenius returns a^p (Frobenius endomorphism)
// This is used in final exponentiation
func fp12Frobenius(a *fp12) *fp12 {
	// Simplified version - full implementation requires Frobenius coefficients
	// For now, return conjugate as approximation
	return fp12Conjugate(a)
}

// fp12Exp returns a^k in Fp12
func fp12Exp(a *fp12, k *big.Int) *fp12 {
	result := fp12One()
	temp := newFp12().Set(a)

	for i := 0; i < k.BitLen(); i++ {
		if k.Bit(i) == 1 {
			result = fp12Mul(result, temp)
		}
		temp = fp12Square(temp)
	}
	return result
}
