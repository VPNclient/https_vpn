// Package sm9 implements SM9 identity-based cryptography per GB/T 38635.
// This file implements the BN256 pairing-friendly elliptic curve.
package sm9

import (
	"crypto/rand"
	"io"
	"math/big"
)

// BN256 curve parameters per GB/T 38635
// t = 0x600000000058F98A
// p = 36t^4 + 36t^3 + 24t^2 + 6t + 1
// n = 36t^4 + 36t^3 + 18t^2 + 6t + 1
// E: y² = x³ + 5 over Fp

var (
	// t is the BN parameter
	t = bigFromHex("600000000058F98A")

	// p is the field characteristic
	p = bigFromHex("B640000002A3A6F1D603AB4FF58EC74521F2934B1A7AEEDBE56F9B27E351457D")

	// n is the curve order (also order of G1 and G2)
	n = bigFromHex("B640000002A3A6F1D603AB4FF58EC74449F2934B18EA8BEEE56EE19CD69ECF25")

	// b is the curve coefficient (E: y² = x³ + b)
	b = big.NewInt(5)

	// Generator of G1: P1 = (x1, y1)
	g1x = bigFromHex("93DE051D62BF718FF5ED0704487D01D6E1E4086909DC3280E8C4E4817C66DDDD")
	g1y = bigFromHex("21FE8DDA4F21E607631065125C395BBC1C1C00CBFA6024350C464CD70A3EA616")

	// Generator of G2: P2 = (x2, y2) where x2, y2 ∈ Fp2
	// x2 = x2_0 + x2_1 * u
	g2x0 = bigFromHex("85AEF3D078640C98597B6027B441A01FF1DD2C190F5E93C454806C11D8806141")
	g2x1 = bigFromHex("3722755292130B08D2AAB97FD34EC120EE265948D19C17ABF9B7213BAF82D65B")
	// y2 = y2_0 + y2_1 * u
	g2y0 = bigFromHex("17509B092E845C1266BA0D262CBEE6ED0736A96FA347C8BD856DC76B84EBEB96")
	g2y1 = bigFromHex("A7CF28D519BE3DA65F3170153D278FF247EFBA98A71A08116215BBA5C999A7C7")

	// Precomputed values
	pMinus1Over2 = new(big.Int).Rsh(p, 1)        // (p-1)/2
	pMinus2      = new(big.Int).Sub(p, big.NewInt(2)) // p-2 for inversion
)

func bigFromHex(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 16)
	return n
}

// ---------- Fp: Base field arithmetic (mod p) ----------

// fpAdd returns (a + b) mod p
func fpAdd(a, b *big.Int) *big.Int {
	r := new(big.Int).Add(a, b)
	if r.Cmp(p) >= 0 {
		r.Sub(r, p)
	}
	return r
}

// fpSub returns (a - b) mod p
func fpSub(a, b *big.Int) *big.Int {
	r := new(big.Int).Sub(a, b)
	if r.Sign() < 0 {
		r.Add(r, p)
	}
	return r
}

// fpMul returns (a * b) mod p
func fpMul(a, b *big.Int) *big.Int {
	r := new(big.Int).Mul(a, b)
	r.Mod(r, p)
	return r
}

// fpSquare returns a² mod p
func fpSquare(a *big.Int) *big.Int {
	return fpMul(a, a)
}

// fpNeg returns -a mod p
func fpNeg(a *big.Int) *big.Int {
	if a.Sign() == 0 {
		return new(big.Int)
	}
	return new(big.Int).Sub(p, a)
}

// fpInv returns a⁻¹ mod p using Fermat's little theorem
func fpInv(a *big.Int) *big.Int {
	return new(big.Int).Exp(a, pMinus2, p)
}

// fpDiv returns a/b mod p
func fpDiv(a, b *big.Int) *big.Int {
	return fpMul(a, fpInv(b))
}

// fpDouble returns 2a mod p
func fpDouble(a *big.Int) *big.Int {
	return fpAdd(a, a)
}

// fpTriple returns 3a mod p
func fpTriple(a *big.Int) *big.Int {
	return fpAdd(fpDouble(a), a)
}

// ---------- Fp2: Quadratic extension Fp[u]/(u² + 1) ----------

// fp2 represents an element of Fp2 = Fp[u]/(u² + 1)
// a = a0 + a1*u
type fp2 struct {
	c0, c1 *big.Int
}

func newFp2() *fp2 {
	return &fp2{new(big.Int), new(big.Int)}
}

func fp2FromInt(a, b *big.Int) *fp2 {
	return &fp2{new(big.Int).Set(a), new(big.Int).Set(b)}
}

func (e *fp2) Set(a *fp2) *fp2 {
	e.c0.Set(a.c0)
	e.c1.Set(a.c1)
	return e
}

func (e *fp2) IsZero() bool {
	return e.c0.Sign() == 0 && e.c1.Sign() == 0
}

func (e *fp2) IsOne() bool {
	return e.c0.Cmp(big.NewInt(1)) == 0 && e.c1.Sign() == 0
}

// fp2Add returns a + b in Fp2
func fp2Add(a, b *fp2) *fp2 {
	return &fp2{
		c0: fpAdd(a.c0, b.c0),
		c1: fpAdd(a.c1, b.c1),
	}
}

// fp2Sub returns a - b in Fp2
func fp2Sub(a, b *fp2) *fp2 {
	return &fp2{
		c0: fpSub(a.c0, b.c0),
		c1: fpSub(a.c1, b.c1),
	}
}

// fp2Neg returns -a in Fp2
func fp2Neg(a *fp2) *fp2 {
	return &fp2{
		c0: fpNeg(a.c0),
		c1: fpNeg(a.c1),
	}
}

// fp2Mul returns a * b in Fp2
// (a0 + a1*u)(b0 + b1*u) = (a0*b0 - a1*b1) + (a0*b1 + a1*b0)*u
func fp2Mul(a, b *fp2) *fp2 {
	t0 := fpMul(a.c0, b.c0)
	t1 := fpMul(a.c1, b.c1)
	t2 := fpAdd(a.c0, a.c1)
	t3 := fpAdd(b.c0, b.c1)
	t2 = fpMul(t2, t3)
	t2 = fpSub(t2, t0)
	t2 = fpSub(t2, t1)
	return &fp2{
		c0: fpSub(t0, t1), // a0*b0 - a1*b1 (since u² = -1)
		c1: t2,
	}
}

// fp2Square returns a² in Fp2
func fp2Square(a *fp2) *fp2 {
	// (a0 + a1*u)² = (a0² - a1²) + 2*a0*a1*u
	t0 := fpAdd(a.c0, a.c1)
	t1 := fpSub(a.c0, a.c1)
	t2 := fpDouble(fpMul(a.c0, a.c1))
	return &fp2{
		c0: fpMul(t0, t1),
		c1: t2,
	}
}

// fp2Inv returns a⁻¹ in Fp2
// 1/(a0 + a1*u) = (a0 - a1*u) / (a0² + a1²)
func fp2Inv(a *fp2) *fp2 {
	t0 := fpSquare(a.c0)
	t1 := fpSquare(a.c1)
	t0 = fpAdd(t0, t1) // a0² + a1²
	t0 = fpInv(t0)
	return &fp2{
		c0: fpMul(a.c0, t0),
		c1: fpNeg(fpMul(a.c1, t0)),
	}
}

// fp2MulXi returns a * ξ where ξ = u + 1 (used in twist)
// a * (1 + u) = (a0 - a1) + (a0 + a1)*u
func fp2MulXi(a *fp2) *fp2 {
	return &fp2{
		c0: fpSub(a.c0, a.c1),
		c1: fpAdd(a.c0, a.c1),
	}
}

// fp2Conjugate returns conjugate of a: a0 - a1*u
func fp2Conjugate(a *fp2) *fp2 {
	return &fp2{
		c0: new(big.Int).Set(a.c0),
		c1: fpNeg(a.c1),
	}
}

// ---------- G1: Points on E(Fp) ----------

// G1 represents a point on the curve E: y² = x³ + 5 over Fp
type G1 struct {
	x, y, z *big.Int // Jacobian coordinates
}

// G1Generator returns the generator of G1
func G1Generator() *G1 {
	return &G1{
		x: new(big.Int).Set(g1x),
		y: new(big.Int).Set(g1y),
		z: big.NewInt(1),
	}
}

// G1Identity returns the identity element (point at infinity)
func G1Identity() *G1 {
	return &G1{
		x: big.NewInt(0),
		y: big.NewInt(1),
		z: big.NewInt(0),
	}
}

// IsIdentity returns true if g is the identity
func (g *G1) IsIdentity() bool {
	return g.z.Sign() == 0
}

// Set copies a to g
func (g *G1) Set(a *G1) *G1 {
	g.x = new(big.Int).Set(a.x)
	g.y = new(big.Int).Set(a.y)
	g.z = new(big.Int).Set(a.z)
	return g
}

// ToAffine converts from Jacobian to affine coordinates
func (g *G1) ToAffine() (*big.Int, *big.Int) {
	if g.IsIdentity() {
		return nil, nil
	}
	zInv := fpInv(g.z)
	zInv2 := fpSquare(zInv)
	zInv3 := fpMul(zInv2, zInv)
	return fpMul(g.x, zInv2), fpMul(g.y, zInv3)
}

// g1Add returns a + b on G1 (Jacobian coordinates)
func g1Add(a, b *G1) *G1 {
	if a.IsIdentity() {
		return new(G1).Set(b)
	}
	if b.IsIdentity() {
		return new(G1).Set(a)
	}

	// http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian.html#addition-add-2007-bl
	z1z1 := fpSquare(a.z)
	z2z2 := fpSquare(b.z)
	u1 := fpMul(a.x, z2z2)
	u2 := fpMul(b.x, z1z1)
	s1 := fpMul(a.y, fpMul(b.z, z2z2))
	s2 := fpMul(b.y, fpMul(a.z, z1z1))
	h := fpSub(u2, u1)
	i := fpSquare(fpDouble(h))
	j := fpMul(h, i)
	r := fpDouble(fpSub(s2, s1))

	if h.Sign() == 0 {
		if r.Sign() == 0 {
			return g1Double(a)
		}
		return G1Identity()
	}

	v := fpMul(u1, i)
	x3 := fpSub(fpSub(fpSquare(r), j), fpDouble(v))
	y3 := fpSub(fpMul(r, fpSub(v, x3)), fpDouble(fpMul(s1, j)))
	z3 := fpMul(fpSub(fpSquare(fpAdd(a.z, b.z)), fpAdd(z1z1, z2z2)), h)

	return &G1{x: x3, y: y3, z: z3}
}

// g1Double returns 2a on G1
func g1Double(a *G1) *G1 {
	if a.IsIdentity() {
		return G1Identity()
	}

	// http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian.html#doubling-dbl-2007-bl
	xx := fpSquare(a.x)
	yy := fpSquare(a.y)
	yyyy := fpSquare(yy)
	zz := fpSquare(a.z)
	s := fpDouble(fpSub(fpSub(fpSquare(fpAdd(a.x, yy)), xx), yyyy))
	m := fpTriple(xx) // 3*x² (a=0 for BN curves)
	t := fpSub(fpSquare(m), fpDouble(s))
	x3 := t
	y3 := fpSub(fpMul(m, fpSub(s, t)), fpMul(big.NewInt(8), yyyy))
	z3 := fpSub(fpSquare(fpAdd(a.y, a.z)), fpAdd(yy, zz))

	return &G1{x: x3, y: y3, z: z3}
}

// g1ScalarMult returns k*a on G1
func g1ScalarMult(a *G1, k *big.Int) *G1 {
	result := G1Identity()
	temp := new(G1).Set(a)

	for i := 0; i < k.BitLen(); i++ {
		if k.Bit(i) == 1 {
			result = g1Add(result, temp)
		}
		temp = g1Double(temp)
	}
	return result
}

// G1Neg returns -a on G1
func g1Neg(a *G1) *G1 {
	if a.IsIdentity() {
		return G1Identity()
	}
	return &G1{
		x: new(big.Int).Set(a.x),
		y: fpNeg(a.y),
		z: new(big.Int).Set(a.z),
	}
}

// ---------- G2: Points on E'(Fp2) (twist) ----------

// G2 represents a point on the twisted curve E': y² = x³ + b/ξ over Fp2
type G2 struct {
	x, y, z *fp2 // Jacobian coordinates in Fp2
}

// G2Generator returns the generator of G2
func G2Generator() *G2 {
	return &G2{
		x: fp2FromInt(g2x0, g2x1),
		y: fp2FromInt(g2y0, g2y1),
		z: &fp2{big.NewInt(1), big.NewInt(0)},
	}
}

// G2Identity returns the identity element
func G2Identity() *G2 {
	return &G2{
		x: &fp2{big.NewInt(0), big.NewInt(0)},
		y: &fp2{big.NewInt(1), big.NewInt(0)},
		z: &fp2{big.NewInt(0), big.NewInt(0)},
	}
}

// IsIdentity returns true if g is the identity
func (g *G2) IsIdentity() bool {
	return g.z.IsZero()
}

// Set copies a to g
func (g *G2) Set(a *G2) *G2 {
	g.x = newFp2().Set(a.x)
	g.y = newFp2().Set(a.y)
	g.z = newFp2().Set(a.z)
	return g
}

// g2Add returns a + b on G2
func g2Add(a, b *G2) *G2 {
	if a.IsIdentity() {
		return new(G2).Set(b)
	}
	if b.IsIdentity() {
		return new(G2).Set(a)
	}

	z1z1 := fp2Square(a.z)
	z2z2 := fp2Square(b.z)
	u1 := fp2Mul(a.x, z2z2)
	u2 := fp2Mul(b.x, z1z1)
	s1 := fp2Mul(a.y, fp2Mul(b.z, z2z2))
	s2 := fp2Mul(b.y, fp2Mul(a.z, z1z1))
	h := fp2Sub(u2, u1)
	i := fp2Square(fp2Add(h, h))
	j := fp2Mul(h, i)
	r := fp2Add(fp2Sub(s2, s1), fp2Sub(s2, s1))

	if h.IsZero() {
		if r.IsZero() {
			return g2Double(a)
		}
		return G2Identity()
	}

	v := fp2Mul(u1, i)
	x3 := fp2Sub(fp2Sub(fp2Square(r), j), fp2Add(v, v))
	y3 := fp2Sub(fp2Mul(r, fp2Sub(v, x3)), fp2Add(fp2Mul(s1, j), fp2Mul(s1, j)))
	z3 := fp2Mul(fp2Sub(fp2Square(fp2Add(a.z, b.z)), fp2Add(z1z1, z2z2)), h)

	return &G2{x: x3, y: y3, z: z3}
}

// g2Double returns 2a on G2
func g2Double(a *G2) *G2 {
	if a.IsIdentity() {
		return G2Identity()
	}

	xx := fp2Square(a.x)
	yy := fp2Square(a.y)
	yyyy := fp2Square(yy)
	zz := fp2Square(a.z)
	s := fp2Add(fp2Sub(fp2Sub(fp2Square(fp2Add(a.x, yy)), xx), yyyy),
		        fp2Sub(fp2Sub(fp2Square(fp2Add(a.x, yy)), xx), yyyy))
	m := fp2Add(fp2Add(xx, xx), xx) // 3*xx (a=0)
	t := fp2Sub(fp2Square(m), fp2Add(s, s))
	x3 := t
	eight := &fp2{big.NewInt(8), big.NewInt(0)}
	y3 := fp2Sub(fp2Mul(m, fp2Sub(s, t)), fp2Mul(eight, yyyy))
	z3 := fp2Sub(fp2Square(fp2Add(a.y, a.z)), fp2Add(yy, zz))

	return &G2{x: x3, y: y3, z: z3}
}

// g2ScalarMult returns k*a on G2
func g2ScalarMult(a *G2, k *big.Int) *G2 {
	result := G2Identity()
	temp := new(G2).Set(a)

	for i := 0; i < k.BitLen(); i++ {
		if k.Bit(i) == 1 {
			result = g2Add(result, temp)
		}
		temp = g2Double(temp)
	}
	return result
}

// ---------- Random scalar generation ----------

// RandomScalar generates a random scalar in [1, n-1]
func RandomScalar(random io.Reader) (*big.Int, error) {
	if random == nil {
		random = rand.Reader
	}

	nMinus1 := new(big.Int).Sub(n, big.NewInt(1))
	k, err := rand.Int(random, nMinus1)
	if err != nil {
		return nil, err
	}
	k.Add(k, big.NewInt(1))
	return k, nil
}

// Order returns the curve order n
func Order() *big.Int {
	return new(big.Int).Set(n)
}
