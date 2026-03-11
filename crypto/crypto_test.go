package crypto_test

import (
	"crypto/tls"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	_ "github.com/nativemind/https-vpn/crypto/us"
)

// TestRegistry_Get tests provider registration and retrieval
func TestRegistry_Get(t *testing.T) {
	// US provider should be registered via init()
	provider, ok := crypto.Get("us")
	if !ok {
		t.Fatal("US provider not found")
	}

	if provider.Name() != "us" {
		t.Errorf("Expected name 'us', got '%s'", provider.Name())
	}
}

// TestRegistry_GetNotFound tests getting non-existent provider
func TestRegistry_GetNotFound(t *testing.T) {
	_, ok := crypto.Get("nonexistent")
	if ok {
		t.Error("Expected provider not found")
	}
}

// TestRegistry_List tests listing providers
func TestRegistry_List(t *testing.T) {
	names := crypto.List()
	if len(names) == 0 {
		t.Error("Expected at least one provider")
	}

	found := false
	for _, name := range names {
		if name == "us" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'us' provider in list")
	}
}

// TestUSProvider_ConfigureTLS tests US provider TLS configuration
func TestUSProvider_ConfigureTLS(t *testing.T) {
	provider, ok := crypto.Get("us")
	if !ok {
		t.Fatal("US provider not found")
	}

	cfg := &tls.Config{}
	if err := provider.ConfigureTLS(cfg); err != nil {
		t.Fatalf("ConfigureTLS failed: %v", err)
	}

	if cfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("Expected MinVersion TLS13, got %d", cfg.MinVersion)
	}

	suites := provider.SupportedCipherSuites()
	if len(suites) == 0 {
		t.Error("Expected cipher suites")
	}
}
