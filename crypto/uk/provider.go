package uk

import (
	"crypto/tls"
	"github.com/nativemind/https-vpn/crypto"
)

// Provider implements crypto.Provider for UK NCSC compliant cryptography.
type Provider struct{}

// Name returns the provider identifier.
func (p *Provider) Name() string { return "uk" }

// ConfigureTLS configures tls.Config for NCSC-recommended TLS 1.3 settings.
func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	cfg.MinVersion = tls.VersionTLS13
	cfg.MaxVersion = tls.VersionTLS13 // Strictly TLS 1.3
	cfg.CipherSuites = p.SupportedCipherSuites()
	cfg.CurvePreferences = []tls.CurveID{
		tls.CurveP384, // NCSC prefers P-384 for high assurance
		tls.CurveP256,
	}
	// PreferServerCipherSuites is deprecated in TLS 1.3 but can be set for completeness
	// and for potential compatibility if an older TLS version somehow sneaks through.
	// However, since MinVersion and MaxVersion are TLS13, this setting has no effect.
	cfg.PreferServerCipherSuites = true 
	return nil
}

// SupportedCipherSuites returns NCSC-recommended TLS 1.3 cipher suites.
// Prioritizes AES-256-GCM-SHA384 as per NCSC high assurance guidance.
func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
	}
}

func init() {
	crypto.Register(&Provider{})
}
