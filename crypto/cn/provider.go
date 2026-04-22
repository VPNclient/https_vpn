// Package cn provides Chinese national cryptography (SM series) implementation.
package cn

import (
	"crypto/tls"

	"github.com/nativemind/https-vpn/crypto"
	smtls "github.com/nativemind/https-vpn/crypto/cn/tls"
)

// Provider implements crypto.Provider for Chinese cryptography (SM series).
type Provider struct{}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "cn"
}

// ConfigureTLS applies SM cryptography settings to tls.Config.
func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	// Note: Standard Go TLS doesn't support SM cipher suites natively.
	// This configuration is a placeholder for custom TLS implementations.
	//
	// For actual SM TLS support, you would need:
	// 1. Custom TLS implementation with SM cipher suites
	// 2. SM2 certificates
	// 3. SM3 for handshake hashing
	// 4. SM4-GCM/CCM for record encryption
	//
	// The cipher suite IDs are defined in crypto/cn/tls/cipher_suites.go
	// per RFC 8998.

	// Set minimum TLS version (SM cipher suites are TLS 1.3 only per RFC 8998)
	if cfg.MinVersion < tls.VersionTLS13 {
		cfg.MinVersion = tls.VersionTLS13
	}

	return nil
}

// SupportedCipherSuites returns the list of supported SM cipher suite IDs.
func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		smtls.TLS_SM4_GCM_SM3,
		smtls.TLS_SM4_CCM_SM3,
	}
}

func init() {
	crypto.Register(&Provider{})
}
