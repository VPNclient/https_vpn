// Certificate store with automatic provider-based selection.
package crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/nativemind/https-vpn/infra/conf"
)

// CertificateStore holds certificates indexed by crypto provider.
// It implements tls.Config.GetCertificate for automatic selection.
type CertificateStore struct {
	// byProvider maps provider name to certificate
	byProvider map[string]*tls.Certificate

	// priority is the order to try providers (from cipherSuites config)
	priority []string

	// defaultCert is used when no provider matches
	defaultCert *tls.Certificate

	// all contains all loaded certificates
	all []tls.Certificate
}

// NewCertificateStore loads certificates and categorizes by provider.
func NewCertificateStore(certs []conf.CertConfig, priority []string) (*CertificateStore, error) {
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates configured")
	}

	cs := &CertificateStore{
		byProvider: make(map[string]*tls.Certificate),
		priority:   priority,
		all:        make([]tls.Certificate, 0, len(certs)),
	}

	for i, certCfg := range certs {
		cert, err := tls.LoadX509KeyPair(certCfg.CertificateFile, certCfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate %d: %w", i, err)
		}

		// Parse leaf certificate for key type detection
		if len(cert.Certificate) > 0 {
			leaf, err := x509.ParseCertificate(cert.Certificate[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate %d: %w", i, err)
			}
			cert.Leaf = leaf
		}

		cs.all = append(cs.all, cert)

		// Set default to first certificate
		if cs.defaultCert == nil {
			cs.defaultCert = &cs.all[0]
		}

		// Categorize by provider
		provider := detectProvider(&cert)
		if _, exists := cs.byProvider[provider]; !exists {
			cs.byProvider[provider] = &cs.all[len(cs.all)-1]
			log.Printf("Certificate loaded: provider=%s file=%s", provider, certCfg.CertificateFile)
		}
	}

	log.Printf("CertificateStore: %d certificates, providers=%v, priority=%v",
		len(cs.all), keys(cs.byProvider), cs.priority)

	return cs, nil
}

// GetCertificate implements tls.Config.GetCertificate.
// It selects certificate based on client's supported cipher suites.
func (cs *CertificateStore) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if hello == nil {
		return cs.defaultCert, nil
	}

	// Determine which providers the client supports
	clientProviders := make(map[string]bool)

	for _, suite := range hello.CipherSuites {
		switch {
		case IsSM4Suite(suite):
			clientProviders["cn"] = true
		case IsGOSTSuite(suite):
			clientProviders["ru"] = true
		default:
			clientProviders["us"] = true
		}
	}

	// Select certificate using priority order
	for _, provider := range cs.priority {
		if clientProviders[provider] {
			if cert, ok := cs.byProvider[provider]; ok {
				log.Printf("Certificate selected: provider=%s sni=%s", provider, hello.ServerName)
				return cert, nil
			}
		}
	}

	// Fallback to default
	log.Printf("Certificate fallback: default sni=%s", hello.ServerName)
	return cs.defaultCert, nil
}

// AllCertificates returns all loaded certificates.
// Used to populate tls.Config.Certificates as fallback.
func (cs *CertificateStore) AllCertificates() []tls.Certificate {
	return cs.all
}

// detectProvider determines crypto provider from certificate key type.
func detectProvider(cert *tls.Certificate) string {
	if cert.Leaf == nil {
		return "us"
	}

	switch pub := cert.Leaf.PublicKey.(type) {
	case *rsa.PublicKey:
		return "us"

	case *ecdsa.PublicKey:
		// Check curve name for SM2
		if pub.Curve != nil {
			params := pub.Curve.Params()
			if params != nil && params.Name == "SM2-P256" {
				return "cn"
			}
		}
		return "us"

	default:
		// Check for GOST by examining OID in certificate
		// GOST keys use OID 1.2.643.7.1.1.1.1 or 1.2.643.7.1.1.1.2
		if cert.Leaf.PublicKeyAlgorithm == x509.UnknownPublicKeyAlgorithm {
			// Try to detect GOST from signature algorithm OID
			sigAlg := cert.Leaf.SignatureAlgorithm.String()
			if containsGOST(sigAlg) {
				return "ru"
			}
		}
		return "us"
	}
}

// containsGOST checks if algorithm string contains GOST identifiers.
func containsGOST(s string) bool {
	// GOST signature algorithms contain "GOST" in name
	for i := 0; i+4 <= len(s); i++ {
		if s[i:i+4] == "GOST" || s[i:i+4] == "gost" {
			return true
		}
	}
	return false
}

// IsSM4Suite returns true for Chinese SM4 cipher suites (RFC 8998).
func IsSM4Suite(suite uint16) bool {
	return suite == 0x00C6 || // TLS_SM4_GCM_SM3
		suite == 0x00C7 // TLS_SM4_CCM_SM3
}

// IsGOSTSuite returns true for Russian GOST cipher suites.
func IsGOSTSuite(suite uint16) bool {
	// GOST cipher suites (implementation-defined range)
	// TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC = 0xFF85
	// TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC = 0xFF86
	// TLS_GOSTR341112_256_WITH_28147_CNT_IMIT = 0xFF87
	return suite >= 0xFF85 && suite <= 0xFF88
}

// keys returns map keys as slice.
func keys(m map[string]*tls.Certificate) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
