// Package tls містить константи та типи для TLS з українською криптографією.
package tls

import "crypto/tls"

// Cipher Suites для українського провайдера ДСТУ-ПК 2026
const (
	// TLS_UA_KALYNA_512_GCM_KUPYNA_512 використовує:
	// - AEAD: Kalyna-512-GCM (512-біт ключ)
	// - Hash: Kupyna-512
	// - KEM: X25519 + Malva-1024 (гібрид)
	TLS_UA_KALYNA_512_GCM_KUPYNA_512 uint16 = 0xD001

	// TLS_UA_KALYNA_256_GCM_KUPYNA_256 - полегшений варіант
	// - AEAD: Kalyna-256-GCM
	// - Hash: Kupyna-256
	TLS_UA_KALYNA_256_GCM_KUPYNA_256 uint16 = 0xD002

	// Резервні ID для майбутнього розширення
	TLS_UA_RESERVED_01 uint16 = 0xD003
	TLS_UA_RESERVED_02 uint16 = 0xD004
)

// Ідентифікатори кривих та груп ключового обміну
const (
	// CurveX25519Malva - гібридна група: X25519 + Malva-1024
	CurveX25519Malva tls.CurveID = 0x6D01

	// CurveDSTU4145_512 - українська еліптична крива 512 біт (ДСТУ 4145)
	CurveDSTU4145_512 tls.CurveID = 0x6D02

	// CurveDSTU4145_431 - українська еліптична крива 431 біт
	CurveDSTU4145_431 tls.CurveID = 0x6D03

	// CurveDSTU4145_307 - українська еліптична крива 307 біт
	CurveDSTU4145_307 tls.CurveID = 0x6D04
)

// Ідентифікатори алгоритмів підпису
const (
	// SignatureSokil512 - постквантовий підпис Сокіл (аналог Dilithium-5)
	SignatureSokil512 uint16 = 0x0720

	// SignatureSokil256 - полегшений Сокіл (аналог Dilithium-3)
	SignatureSokil256 uint16 = 0x0721

	// SignatureDSTU4145_512 - ДСТУ 4145 з кривою 512 біт
	SignatureDSTU4145_512 uint16 = 0x0722

	// SignatureDSTU4145_431 - ДСТУ 4145 з кривою 431 біт
	SignatureDSTU4145_431 uint16 = 0x0723

	// SignatureHybridSokilDSTU - гібридний підпис (Сокіл + ДСТУ 4145)
	SignatureHybridSokilDSTU uint16 = 0x0724
)

// Імена Cipher Suites для логування та відображення
var CipherSuiteNames = map[uint16]string{
	TLS_UA_KALYNA_512_GCM_KUPYNA_512: "TLS_UA_KALYNA_512_GCM_KUPYNA_512",
	TLS_UA_KALYNA_256_GCM_KUPYNA_256: "TLS_UA_KALYNA_256_GCM_KUPYNA_256",
}

// CurveNames - імена кривих для логування
var CurveNames = map[tls.CurveID]string{
	CurveX25519Malva:  "X25519_Malva",
	CurveDSTU4145_512: "DSTU4145_512",
	CurveDSTU4145_431: "DSTU4145_431",
	CurveDSTU4145_307: "DSTU4145_307",
}

// SignatureNames - імена алгоритмів підпису
var SignatureNames = map[uint16]string{
	SignatureSokil512:        "Sokil_512",
	SignatureSokil256:        "Sokil_256",
	SignatureDSTU4145_512:    "DSTU4145_512",
	SignatureDSTU4145_431:    "DSTU4145_431",
	SignatureHybridSokilDSTU: "Hybrid_Sokil_DSTU4145",
}

// IsUACipherSuite перевіряє, чи є suite українським
func IsUACipherSuite(suite uint16) bool {
	return suite >= 0xD001 && suite <= 0xD0FF
}

// IsUACurve перевіряє, чи є крива українською
func IsUACurve(curve tls.CurveID) bool {
	return curve >= 0x6D01 && curve <= 0x6DFF
}

// IsUASignature перевіряє, чи є алгоритм підпису українським
func IsUASignature(sig uint16) bool {
	return sig >= 0x0720 && sig <= 0x072F
}

// SupportedCipherSuites повертає список підтримуваних UA cipher suites
func SupportedCipherSuites() []uint16 {
	return []uint16{
		TLS_UA_KALYNA_512_GCM_KUPYNA_512,
		TLS_UA_KALYNA_256_GCM_KUPYNA_256,
	}
}

// SupportedCurves повертає список підтримуваних UA кривих
func SupportedCurves() []tls.CurveID {
	return []tls.CurveID{
		CurveX25519Malva,
		CurveDSTU4145_512,
	}
}

// CipherSuiteInfo містить інформацію про cipher suite
type CipherSuiteInfo struct {
	ID           uint16
	Name         string
	KeySize      int  // Розмір ключа в байтах
	BlockSize    int  // Розмір блоку в байтах
	HashSize     int  // Розмір хешу в байтах
	IsPostQuatum bool // Чи є постквантовим
}

// GetCipherSuiteInfo повертає інформацію про cipher suite
func GetCipherSuiteInfo(suite uint16) *CipherSuiteInfo {
	switch suite {
	case TLS_UA_KALYNA_512_GCM_KUPYNA_512:
		return &CipherSuiteInfo{
			ID:           suite,
			Name:         "TLS_UA_KALYNA_512_GCM_KUPYNA_512",
			KeySize:      64,
			BlockSize:    64,
			HashSize:     64,
			IsPostQuatum: true,
		}
	case TLS_UA_KALYNA_256_GCM_KUPYNA_256:
		return &CipherSuiteInfo{
			ID:           suite,
			Name:         "TLS_UA_KALYNA_256_GCM_KUPYNA_256",
			KeySize:      32,
			BlockSize:    32,
			HashSize:     32,
			IsPostQuatum: true,
		}
	default:
		return nil
	}
}
