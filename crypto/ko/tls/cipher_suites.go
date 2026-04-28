// Package tls defines TLS cipher suite constants for Korean cryptography.
package tls

// Korean TLS 1.3 cipher suites (RFC 9367)
const (
	// TLS_ARIA_128_GCM_SHA256 is ARIA-128-GCM with SHA-256
	TLS_ARIA_128_GCM_SHA256 uint16 = 0x1306

	// TLS_ARIA_256_GCM_SHA384 is ARIA-256-GCM with SHA-384
	TLS_ARIA_256_GCM_SHA384 uint16 = 0x1307
)

// Korean TLS 1.2 cipher suites (RFC 6209)
const (
	// ECDHE-ECDSA with ARIA-GCM
	TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256 uint16 = 0xC06A
	TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384 uint16 = 0xC06B

	// ECDHE-RSA with ARIA-GCM
	TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256 uint16 = 0xC06C
	TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384 uint16 = 0xC06D

	// DHE-RSA with ARIA-GCM
	TLS_DHE_RSA_WITH_ARIA_128_GCM_SHA256 uint16 = 0xC068
	TLS_DHE_RSA_WITH_ARIA_256_GCM_SHA384 uint16 = 0xC069

	// RSA with ARIA-GCM
	TLS_RSA_WITH_ARIA_128_GCM_SHA256 uint16 = 0xC050
	TLS_RSA_WITH_ARIA_256_GCM_SHA384 uint16 = 0xC051
)

// SEED cipher suites (RFC 4162)
const (
	// TLS_RSA_WITH_SEED_CBC_SHA is SEED-CBC with SHA-1
	TLS_RSA_WITH_SEED_CBC_SHA uint16 = 0x0096

	// TLS_DHE_DSS_WITH_SEED_CBC_SHA is DHE-DSS with SEED-CBC
	TLS_DHE_DSS_WITH_SEED_CBC_SHA uint16 = 0x0099

	// TLS_DHE_RSA_WITH_SEED_CBC_SHA is DHE-RSA with SEED-CBC
	TLS_DHE_RSA_WITH_SEED_CBC_SHA uint16 = 0x009A

	// TLS_DH_anon_WITH_SEED_CBC_SHA is anonymous DH with SEED-CBC
	TLS_DH_anon_WITH_SEED_CBC_SHA uint16 = 0x009B
)

// IsARIACipherSuite returns true if the cipher suite uses ARIA
func IsARIACipherSuite(suite uint16) bool {
	switch suite {
	case TLS_ARIA_128_GCM_SHA256,
		TLS_ARIA_256_GCM_SHA384,
		TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256,
		TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384,
		TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256,
		TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384,
		TLS_DHE_RSA_WITH_ARIA_128_GCM_SHA256,
		TLS_DHE_RSA_WITH_ARIA_256_GCM_SHA384,
		TLS_RSA_WITH_ARIA_128_GCM_SHA256,
		TLS_RSA_WITH_ARIA_256_GCM_SHA384:
		return true
	}
	return false
}

// IsSEEDCipherSuite returns true if the cipher suite uses SEED
func IsSEEDCipherSuite(suite uint16) bool {
	switch suite {
	case TLS_RSA_WITH_SEED_CBC_SHA,
		TLS_DHE_DSS_WITH_SEED_CBC_SHA,
		TLS_DHE_RSA_WITH_SEED_CBC_SHA,
		TLS_DH_anon_WITH_SEED_CBC_SHA:
		return true
	}
	return false
}

// IsKOCipherSuite returns true if the cipher suite is Korean (ARIA or SEED)
func IsKOCipherSuite(suite uint16) bool {
	return IsARIACipherSuite(suite) || IsSEEDCipherSuite(suite)
}

// CipherSuiteInfo contains information about a cipher suite
type CipherSuiteInfo struct {
	ID       uint16
	Name     string
	KeySize  int  // in bits
	IsAEAD   bool
	IsTLS13  bool
}

// GetCipherSuiteInfo returns information about a Korean cipher suite
func GetCipherSuiteInfo(suite uint16) (CipherSuiteInfo, bool) {
	info, ok := cipherSuiteInfoMap[suite]
	return info, ok
}

var cipherSuiteInfoMap = map[uint16]CipherSuiteInfo{
	TLS_ARIA_128_GCM_SHA256: {
		ID:      TLS_ARIA_128_GCM_SHA256,
		Name:    "TLS_ARIA_128_GCM_SHA256",
		KeySize: 128,
		IsAEAD:  true,
		IsTLS13: true,
	},
	TLS_ARIA_256_GCM_SHA384: {
		ID:      TLS_ARIA_256_GCM_SHA384,
		Name:    "TLS_ARIA_256_GCM_SHA384",
		KeySize: 256,
		IsAEAD:  true,
		IsTLS13: true,
	},
	TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256: {
		ID:      TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256,
		Name:    "TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256",
		KeySize: 128,
		IsAEAD:  true,
		IsTLS13: false,
	},
	TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384: {
		ID:      TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384,
		Name:    "TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384",
		KeySize: 256,
		IsAEAD:  true,
		IsTLS13: false,
	},
	TLS_RSA_WITH_SEED_CBC_SHA: {
		ID:      TLS_RSA_WITH_SEED_CBC_SHA,
		Name:    "TLS_RSA_WITH_SEED_CBC_SHA",
		KeySize: 128,
		IsAEAD:  false,
		IsTLS13: false,
	},
}
