// GOST R 34.10-2012 elliptic curves.
// Defines curve parameters for 256-bit and 512-bit key sizes.
package gost

import (
	"crypto/elliptic"
	"math/big"
	"sync"
)

// Curve represents a GOST elliptic curve.
type Curve struct {
	elliptic.CurveParams
	A *big.Int // Curve coefficient a (standard elliptic has a=-3, GOST uses custom)
}

// Params returns the curve parameters.
func (c *Curve) Params() *elliptic.CurveParams {
	return &c.CurveParams
}

// IsOnCurve reports whether the point (x,y) is on the curve.
func (c *Curve) IsOnCurve(x, y *big.Int) bool {
	// y² = x³ + ax + b (mod p)
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, c.P)

	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)

	ax := new(big.Int).Mul(c.A, x)

	rhs := new(big.Int).Add(x3, ax)
	rhs.Add(rhs, c.B)
	rhs.Mod(rhs, c.P)

	return y2.Cmp(rhs) == 0
}

// Add returns the sum of (x1,y1) and (x2,y2).
func (c *Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	// Handle identity
	if x1.Sign() == 0 && y1.Sign() == 0 {
		return new(big.Int).Set(x2), new(big.Int).Set(y2)
	}
	if x2.Sign() == 0 && y2.Sign() == 0 {
		return new(big.Int).Set(x1), new(big.Int).Set(y1)
	}

	// Check if points are the same
	if x1.Cmp(x2) == 0 {
		if y1.Cmp(y2) == 0 {
			return c.Double(x1, y1)
		}
		// Points are inverse, return identity
		return new(big.Int), new(big.Int)
	}

	// λ = (y2 - y1) / (x2 - x1)
	dy := new(big.Int).Sub(y2, y1)
	dx := new(big.Int).Sub(x2, x1)
	dx.ModInverse(dx, c.P)
	lambda := new(big.Int).Mul(dy, dx)
	lambda.Mod(lambda, c.P)

	// x3 = λ² - x1 - x2
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, c.P)

	// y3 = λ(x1 - x3) - y1
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, c.P)

	return x3, y3
}

// Double returns 2*(x,y).
func (c *Curve) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	if y1.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}

	// λ = (3x² + a) / 2y
	x2 := new(big.Int).Mul(x1, x1)
	x2.Mul(x2, big.NewInt(3))
	x2.Add(x2, c.A)

	y2 := new(big.Int).Mul(y1, big.NewInt(2))
	y2.ModInverse(y2, c.P)

	lambda := new(big.Int).Mul(x2, y2)
	lambda.Mod(lambda, c.P)

	// x3 = λ² - 2x1
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x1)
	x3.Mod(x3, c.P)

	// y3 = λ(x1 - x3) - y1
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, c.P)

	return x3, y3
}

// ScalarMult returns k*(x,y).
func (c *Curve) ScalarMult(x1, y1 *big.Int, k []byte) (*big.Int, *big.Int) {
	// Double-and-add
	Bx, By := new(big.Int).Set(x1), new(big.Int).Set(y1)
	Rx, Ry := new(big.Int), new(big.Int)

	for _, b := range k {
		for i := 7; i >= 0; i-- {
			Rx, Ry = c.Double(Rx, Ry)
			if (b>>i)&1 == 1 {
				Rx, Ry = c.Add(Rx, Ry, Bx, By)
			}
		}
	}
	return Rx, Ry
}

// ScalarBaseMult returns k*G where G is the base point.
func (c *Curve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return c.ScalarMult(c.Gx, c.Gy, k)
}

// Curve instances
var (
	initOnce sync.Once

	// 256-bit curves
	CurveIdtc26gost34102012256paramSetA *Curve

	// 512-bit curves
	CurveIdtc26gost34102012512paramSetA *Curve
	CurveIdtc26gost34102012512paramSetB *Curve
	CurveIdtc26gost34102012512paramSetC *Curve
)

func initCurves() {
	initOnce.Do(func() {
		initCurve256A()
		initCurve512A()
		initCurve512B()
		initCurve512C()
	})
}

// TC26256A returns the tc26-gost-3410-2012-256-paramSetA curve.
func TC26256A() *Curve {
	initCurves()
	return CurveIdtc26gost34102012256paramSetA
}

// TC26512A returns the tc26-gost-3410-2012-512-paramSetA curve.
func TC26512A() *Curve {
	initCurves()
	return CurveIdtc26gost34102012512paramSetA
}

// TC26512B returns the tc26-gost-3410-2012-512-paramSetB curve.
func TC26512B() *Curve {
	initCurves()
	return CurveIdtc26gost34102012512paramSetB
}

// TC26512C returns the tc26-gost-3410-2012-512-paramSetC curve.
func TC26512C() *Curve {
	initCurves()
	return CurveIdtc26gost34102012512paramSetC
}

func initCurve256A() {
	// id-tc26-gost-3410-2012-256-paramSetA
	CurveIdtc26gost34102012256paramSetA = &Curve{}
	CurveIdtc26gost34102012256paramSetA.P, _ = new(big.Int).SetString(
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFD97", 16)
	CurveIdtc26gost34102012256paramSetA.N, _ = new(big.Int).SetString(
		"400000000000000000000000000000000FD8CDDFC87B6635C115AF556C360C67", 16)
	CurveIdtc26gost34102012256paramSetA.A, _ = new(big.Int).SetString(
		"C2173F1513981673AF4892C23035A27CE25E2013BF95AA33B22C656F277E7335", 16)
	CurveIdtc26gost34102012256paramSetA.B, _ = new(big.Int).SetString(
		"295F9BAE7428ED9CCC20E7C359A9D41A22FCCD9108E17BF7BA9337A6F8AE9513", 16)
	CurveIdtc26gost34102012256paramSetA.Gx, _ = new(big.Int).SetString(
		"91E38443A5E82C0D880923425712B2BB658B9196932E02C78B2582FE742DAA28", 16)
	CurveIdtc26gost34102012256paramSetA.Gy, _ = new(big.Int).SetString(
		"32879423AB1A0375895786C4BB46E9565FDE0B5344766740AF268ADB32322E5C", 16)
	CurveIdtc26gost34102012256paramSetA.BitSize = 256
	CurveIdtc26gost34102012256paramSetA.Name = "id-tc26-gost-3410-2012-256-paramSetA"
}

func initCurve512A() {
	// id-tc26-gost-3410-2012-512-paramSetA
	CurveIdtc26gost34102012512paramSetA = &Curve{}
	CurveIdtc26gost34102012512paramSetA.P, _ = new(big.Int).SetString(
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFDC7", 16)
	CurveIdtc26gost34102012512paramSetA.N, _ = new(big.Int).SetString(
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF27E69532F48D89116FF22B8D4E0560609B4B38ABFAD2B85DCACDB1411F10B275", 16)
	CurveIdtc26gost34102012512paramSetA.A, _ = new(big.Int).SetString(
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFDC4", 16)
	CurveIdtc26gost34102012512paramSetA.B, _ = new(big.Int).SetString(
		"E8C2505DEDFC86DDC1BD0B2B6667F1DA34B82574761CB0E879BD081CFD0B6265EE3CB090F30D27614CB4574010DA90DD862EF9D4EBEE4761503190785A71C760", 16)
	CurveIdtc26gost34102012512paramSetA.Gx, _ = new(big.Int).SetString(
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003", 16)
	CurveIdtc26gost34102012512paramSetA.Gy, _ = new(big.Int).SetString(
		"7503CFE87A836AE3A61B8816E25450E6CE5E1C93ACF1ABC1778064FDCBEFA921DF1626BE4FD036E93D75E6A50E3A41E98028FE5FC235F5B889A589CB5215F2A4", 16)
	CurveIdtc26gost34102012512paramSetA.BitSize = 512
	CurveIdtc26gost34102012512paramSetA.Name = "id-tc26-gost-3410-2012-512-paramSetA"
}

func initCurve512B() {
	// id-tc26-gost-3410-2012-512-paramSetB
	CurveIdtc26gost34102012512paramSetB = &Curve{}
	CurveIdtc26gost34102012512paramSetB.P, _ = new(big.Int).SetString(
		"8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006F", 16)
	CurveIdtc26gost34102012512paramSetB.N, _ = new(big.Int).SetString(
		"800000000000000000000000000000000000000000000000000000000000000149A1EC142565A545ACFDB77BD9D40CFA8B996712101BEA0EC6346C54374F25BD", 16)
	CurveIdtc26gost34102012512paramSetB.A, _ = new(big.Int).SetString(
		"8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006C", 16)
	CurveIdtc26gost34102012512paramSetB.B, _ = new(big.Int).SetString(
		"687D1B459DC841457E3E06CF6F5E2517B97C7D614AF138BCBF85DC806C4B289F3E965D2DB1416D217F8B276FAD1AB69C50F78BEE1FA3106EFB8CCBC7C5140116", 16)
	CurveIdtc26gost34102012512paramSetB.Gx, _ = new(big.Int).SetString(
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002", 16)
	CurveIdtc26gost34102012512paramSetB.Gy, _ = new(big.Int).SetString(
		"1A8F7EDA389B094C2C071E3647A8940F3C123B697578C213BE6DD9E6C8EC7335DCB228FD1EDF4A39152CBCAAF8C0398828041055F94CEEEC7E21340780FE41BD", 16)
	CurveIdtc26gost34102012512paramSetB.BitSize = 512
	CurveIdtc26gost34102012512paramSetB.Name = "id-tc26-gost-3410-2012-512-paramSetB"
}

func initCurve512C() {
	// id-tc26-gost-3410-2012-512-paramSetC (same as paramSetA with different cofactor)
	CurveIdtc26gost34102012512paramSetC = &Curve{}
	CurveIdtc26gost34102012512paramSetC.P, _ = new(big.Int).SetString(
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFDC7", 16)
	CurveIdtc26gost34102012512paramSetC.N, _ = new(big.Int).SetString(
		"3FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC98CDBA46506AB004C33A9FF5147502CC8EDA9E7A769A12694623CEF47F023ED", 16)
	CurveIdtc26gost34102012512paramSetC.A, _ = new(big.Int).SetString(
		"DC9203E514A721875485A529D2C722FB187BC8980EB866644DE41C68E143064546E861C0E2C9EDD92ADE71F46FCF50FF2AD97F951FDA9F2A2EB6546F39689BD3", 16)
	CurveIdtc26gost34102012512paramSetC.B, _ = new(big.Int).SetString(
		"B4C4EE28CEBC6C2C8AC12952CF37F16AC7EFB6A9F69F4B57FFDA2E4F0DE5ADE038CBC2FFF719D2C18DE0284B8BFEF3B52B8CC7A5F5BF0A3C8D2319A5312557E1", 16)
	CurveIdtc26gost34102012512paramSetC.Gx, _ = new(big.Int).SetString(
		"E2E31EDFC23DE7BDEBE241CE593EF5DE2295B7A9CBAEF021D385F7074CEA043AA27272A7AE602BF2A7B9033DB9ED3610C6FB85487EAE97AAC5BC7928C1950148", 16)
	CurveIdtc26gost34102012512paramSetC.Gy, _ = new(big.Int).SetString(
		"F5CE40D95B5EB899ABBCCFF5911CB8577939804D6527378B8C108C3D2090FF9BE18E2D33E3021ED2EF32D85822423B6304F726AA854BAE07D0396E9A9ADDC40F", 16)
	CurveIdtc26gost34102012512paramSetC.BitSize = 512
	CurveIdtc26gost34102012512paramSetC.Name = "id-tc26-gost-3410-2012-512-paramSetC"
}
