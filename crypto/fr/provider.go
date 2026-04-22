package fr

import (
	"crypto/tls"
	"github.com/nativemind/https-vpn/crypto"
)

type Provider struct{}

func (p *Provider) Name() string { return "fr" }

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	cfg.MinVersion = tls.VersionTLS13
	cfg.CipherSuites = p.SupportedCipherSuites()
	return nil
}

func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
	}
}

func init() {
	crypto.Register(&Provider{})
}
