// Package us provides the US/NIST crypto provider using Go standard library.
// This is the default provider, using standard TLS 1.3 cipher suites.
package us

import (
	"crypto/tls"

	"github.com/nativemind/https-vpn/crypto"
)

func init() {
	crypto.Register(&Provider{})
}

// Provider implements US/NIST cryptography using Go stdlib.
type Provider struct{}

// Name returns the provider identifier.
func (p *Provider) Name() string { return "us" }

// ConfigureTLS configures tls.Config for standard TLS 1.3.
func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	cfg.MinVersion = tls.VersionTLS13
	cfg.CipherSuites = p.SupportedCipherSuites()
	return nil
}

// SupportedCipherSuites returns standard TLS 1.3 cipher suites.
func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
	}
}
