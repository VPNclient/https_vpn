// SM2 elliptic curve parameters per GB/T 32918.5-2017.
package sm2

import (
	"crypto/elliptic"
	"math/big"
	"sync"
)

var initonce sync.Once
var sm2Curve *elliptic.CurveParams

// P256 returns the SM2 P-256 elliptic curve (also known as sm2p256v1).
// This is the recommended curve from GB/T 32918.5-2017.
func P256() elliptic.Curve {
	initonce.Do(initSM2P256)
	return sm2Curve
}

func initSM2P256() {
	sm2Curve = &elliptic.CurveParams{Name: "SM2-P256"}
	sm2Curve.P, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	sm2Curve.N, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)
	sm2Curve.B, _ = new(big.Int).SetString("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93", 16)
	sm2Curve.Gx, _ = new(big.Int).SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	sm2Curve.Gy, _ = new(big.Int).SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)
	sm2Curve.BitSize = 256
}

// sm2A returns the 'a' coefficient for SM2 curve.
// SM2 uses a = p - 3 (same as NIST P-256).
func sm2A() *big.Int {
	a, _ := new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)
	return a
}
