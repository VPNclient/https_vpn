// Package crypto defines the interface for cryptographic providers.
// Each provider implements TLS configuration for specific national cryptography standards.
package crypto

import "crypto/tls"

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
