package gost

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestCurve256ABasePoint(t *testing.T) {
	c := TC26256A()
	if !c.IsOnCurve(c.Gx, c.Gy) {
		t.Error("Base point is not on curve")
	}
}

func TestCurve512ABasePoint(t *testing.T) {
	c := TC26512A()
	if !c.IsOnCurve(c.Gx, c.Gy) {
		t.Error("Base point is not on curve")
	}
}

func TestCurve512BBasePoint(t *testing.T) {
	c := TC26512B()
	if !c.IsOnCurve(c.Gx, c.Gy) {
		t.Error("Base point is not on curve")
	}
}

func TestCurve512CBasePoint(t *testing.T) {
	c := TC26512C()
	if !c.IsOnCurve(c.Gx, c.Gy) {
		t.Error("Base point is not on curve")
	}
}

func TestCurve256ADouble(t *testing.T) {
	c := TC26256A()
	x2, y2 := c.Double(c.Gx, c.Gy)
	if !c.IsOnCurve(x2, y2) {
		t.Error("2G is not on curve")
	}
}

func TestCurve256AAdd(t *testing.T) {
	c := TC26256A()
	// 2G = G + G
	x2a, y2a := c.Double(c.Gx, c.Gy)
	x2b, y2b := c.Add(c.Gx, c.Gy, c.Gx, c.Gy)

	if x2a.Cmp(x2b) != 0 || y2a.Cmp(y2b) != 0 {
		t.Error("Double and Add don't match for G+G")
	}
}

func TestCurve256AScalarMult(t *testing.T) {
	c := TC26256A()

	// Test that n*G = O (identity)
	x, y := c.ScalarBaseMult(c.N.Bytes())
	if x.Sign() != 0 || y.Sign() != 0 {
		t.Error("n*G should be identity")
	}

	// Test 2*G
	two := big.NewInt(2)
	x2a, y2a := c.ScalarBaseMult(two.Bytes())
	x2b, y2b := c.Double(c.Gx, c.Gy)

	if x2a.Cmp(x2b) != 0 || y2a.Cmp(y2b) != 0 {
		t.Error("ScalarBaseMult(2) != Double(G)")
	}
}

func TestCurve256AScalarMultAssociativity(t *testing.T) {
	c := TC26256A()

	// (a*b)*G = a*(b*G)
	a := big.NewInt(7)
	b := big.NewInt(13)
	ab := new(big.Int).Mul(a, b)

	// (a*b)*G
	x1, y1 := c.ScalarBaseMult(ab.Bytes())

	// b*G
	xb, yb := c.ScalarBaseMult(b.Bytes())
	// a*(b*G)
	x2, y2 := c.ScalarMult(xb, yb, a.Bytes())

	if x1.Cmp(x2) != 0 || y1.Cmp(y2) != 0 {
		t.Error("Scalar multiplication not associative")
	}
}

func TestCurve512AScalarMult(t *testing.T) {
	c := TC26512A()

	// Test 3*G
	three := big.NewInt(3)
	x3, y3 := c.ScalarBaseMult(three.Bytes())
	if !c.IsOnCurve(x3, y3) {
		t.Error("3G is not on curve")
	}

	// Alternative: G + 2G
	x2, y2 := c.Double(c.Gx, c.Gy)
	x3b, y3b := c.Add(c.Gx, c.Gy, x2, y2)

	if x3.Cmp(x3b) != 0 || y3.Cmp(y3b) != 0 {
		t.Error("3G != G + 2G")
	}
}

func TestCurveRandomPoint(t *testing.T) {
	c := TC26256A()

	// Generate random scalar
	k := make([]byte, 32)
	rand.Read(k)

	x, y := c.ScalarBaseMult(k)
	if !c.IsOnCurve(x, y) {
		t.Error("Random point k*G is not on curve")
	}
}

func TestCurveIdentity(t *testing.T) {
	c := TC26256A()

	// G + O = G
	zero := new(big.Int)
	x, y := c.Add(c.Gx, c.Gy, zero, zero)

	if x.Cmp(c.Gx) != 0 || y.Cmp(c.Gy) != 0 {
		t.Error("G + O != G")
	}

	// O + G = G
	x, y = c.Add(zero, zero, c.Gx, c.Gy)
	if x.Cmp(c.Gx) != 0 || y.Cmp(c.Gy) != 0 {
		t.Error("O + G != G")
	}
}

func TestCurveInverse(t *testing.T) {
	c := TC26256A()

	// G + (-G) = O
	negY := new(big.Int).Sub(c.P, c.Gy)
	x, y := c.Add(c.Gx, c.Gy, c.Gx, negY)

	if x.Sign() != 0 || y.Sign() != 0 {
		t.Error("G + (-G) != O")
	}
}

func BenchmarkCurve256AScalarMult(b *testing.B) {
	c := TC26256A()
	k := make([]byte, 32)
	rand.Read(k)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ScalarBaseMult(k)
	}
}

func BenchmarkCurve512AScalarMult(b *testing.B) {
	c := TC26512A()
	k := make([]byte, 64)
	rand.Read(k)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ScalarBaseMult(k)
	}
}
