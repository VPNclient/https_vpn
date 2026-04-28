// Package ko implements the cryptographic provider for Korean national standards.
// Supports ARIA (KS X 1213) and SEED (KISA) block ciphers.
package ko

import (
	"crypto/tls"

	"github.com/nativemind/https-vpn/crypto"
	kotls "github.com/nativemind/https-vpn/crypto/ko/tls"
)

// Provider implements crypto.Provider for Korean cryptography
type Provider struct{}

func init() {
	// Register the provider on package import
	crypto.Register(&Provider{})
}

// Name returns the provider identifier
func (p *Provider) Name() string {
	return "ko"
}

// ConfigureTLS configures TLS for Korean cryptography
func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	// Set TLS 1.3 as minimum (TLS 1.2 as fallback if needed)
	cfg.MinVersion = tls.VersionTLS12
	cfg.MaxVersion = tls.VersionTLS13

	// Configure cipher suites
	// Note: Go's standard library doesn't support custom cipher suites,
	// so we use fallback to AES-GCM with preference order
	cfg.CipherSuites = []uint16{
		tls.TLS_AES_256_GCM_SHA384,      // TLS 1.3
		tls.TLS_AES_128_GCM_SHA256,      // TLS 1.3
		tls.TLS_CHACHA20_POLY1305_SHA256, // TLS 1.3
		// TLS 1.2 fallback
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}

	// Configure curve preferences
	cfg.CurvePreferences = []tls.CurveID{
		tls.X25519,
		tls.CurveP384,
		tls.CurveP256,
	}

	return nil
}

// SupportedCipherSuites returns the list of supported cipher suites
func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		// Korean TLS 1.3 suites
		kotls.TLS_ARIA_256_GCM_SHA384,
		kotls.TLS_ARIA_128_GCM_SHA256,
		// Korean TLS 1.2 suites
		kotls.TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384,
		kotls.TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256,
		kotls.TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384,
		kotls.TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256,
		// SEED (legacy)
		kotls.TLS_RSA_WITH_SEED_CBC_SHA,
		// Fallback
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
	}
}

// Description returns a description of the provider
func (p *Provider) Description() string {
	return "Korean National Cryptography (ARIA/SEED - KS X 1213, KISA)"
}

// Algorithms returns the list of supported algorithms
func (p *Provider) Algorithms() []string {
	return []string{
		"ARIA-256-GCM (block cipher)",
		"ARIA-192-GCM (block cipher)",
		"ARIA-128-GCM (block cipher)",
		"SEED-128-CBC (block cipher, legacy)",
		"SHA-256 (hash function)",
		"SHA-384 (hash function)",
	}
}

// IsPostQuantum returns false as Korean standards don't include PQ algorithms yet
func (p *Provider) IsPostQuantum() bool {
	return false
}

// SecurityLevel returns the security level in bits
func (p *Provider) SecurityLevel() int {
	return 256 // ARIA-256 provides 256-bit security
}
