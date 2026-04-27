// Package crypto defines the interface for cryptographic providers.
// Each provider implements TLS configuration for specific national cryptography standards.
package crypto

import (
	"crypto/tls"
	"strings"
)

// Provider configures TLS with specific cryptographic algorithms.
type Provider interface {
	// Name returns provider identifier (e.g., "us", "ru", "cn").
	Name() string

	// ConfigureTLS applies crypto settings to tls.Config.
	ConfigureTLS(cfg *tls.Config) error

	// SupportedCipherSuites returns list of supported cipher suite IDs.
	SupportedCipherSuites() []uint16
}

// Registry holds available crypto providers.
var Registry = make(map[string]Provider)

// Register adds a provider to the global registry.
// Called automatically by provider packages via init().
func Register(p Provider) {
	Registry[p.Name()] = p
}

// Get returns a provider by name.
// Returns nil, false if provider not found.
func Get(name string) (Provider, bool) {
	p, ok := Registry[name]
	return p, ok
}

// List returns names of all registered providers.
func List() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}

// IsUACryptoSuite checks if suite ID is Ukrainian (ДСТУ-ПК 2026).
// UA suites use range 0xD001-0xD0FF.
func IsUACryptoSuite(suite uint16) bool {
	return suite >= 0xD001 && suite <= 0xD0FF
}

// ParseProviderPriority extracts provider names from cipherSuites config.
// Input examples: "cn,ru,us" or "cn,TLS_AES_128_GCM_SHA256,ru"
// Returns list of valid provider names in order, or ["us"] if none found.
func ParseProviderPriority(cipherSuites string) []string {
	var priority []string
	seen := make(map[string]bool)

	for _, part := range strings.Split(cipherSuites, ",") {
		name := strings.TrimSpace(strings.ToLower(part))
		if name == "" {
			continue
		}
		if _, ok := Registry[name]; ok && !seen[name] {
			priority = append(priority, name)
			seen[name] = true
		}
	}

	// Default fallback
	if len(priority) == 0 {
		priority = []string{"us"}
	}

	return priority
}
