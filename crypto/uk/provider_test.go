package uk_test

import (
	"crypto/tls"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	"github.com/nativemind/https-vpn/crypto/uk"
)

func TestUKProviderName(t *testing.T) {
	p := &uk.Provider{}
	if p.Name() != "uk" {
		t.Errorf("Expected provider name 'uk', got '%s'", p.Name())
	}
}

func TestUKProviderConfigureTLS(t *testing.T) {
	p := &uk.Provider{}
	cfg := &tls.Config{}

	err := p.ConfigureTLS(cfg)
	if err != nil {
		t.Fatalf("ConfigureTLS returned an error: %v", err)
	}

	// Verify MinVersion and MaxVersion are TLS13
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("Expected MinVersion to be TLS13 (%d), got %d", tls.VersionTLS13, cfg.MinVersion)
	}
	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("Expected MaxVersion to be TLS13 (%d), got %d", tls.VersionTLS13, cfg.MaxVersion)
	}

	// Verify SupportedCipherSuites
	expectedCipherSuites := []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
	}
	if len(cfg.CipherSuites) != len(expectedCipherSuites) {
		t.Errorf("Expected %d cipher suites, got %d", len(expectedCipherSuites), len(cfg.CipherSuites))
	} else {
		for i, cs := range cfg.CipherSuites {
			if cs != expectedCipherSuites[i] {
				t.Errorf("Cipher suite at index %d: Expected %x, got %x", i, expectedCipherSuites[i], cs)
			}
		}
	}

	// Verify CurvePreferences
	expectedCurvePreferences := []tls.CurveID{
		tls.CurveP384,
		tls.CurveP256,
	}
	if len(cfg.CurvePreferences) != len(expectedCurvePreferences) {
		t.Errorf("Expected %d curve preferences, got %d", len(expectedCurvePreferences), len(cfg.CurvePreferences))
	} else {
		for i, curve := range cfg.CurvePreferences {
			if curve != expectedCurvePreferences[i] {
				t.Errorf("Curve preference at index %d: Expected %d, got %d", i, expectedCurvePreferences[i], curve)
			}
		}
	}

	// Verify PreferServerCipherSuites
	if !cfg.PreferServerCipherSuites {
		t.Error("Expected PreferServerCipherSuites to be true")
	}
}

func TestUKProviderRegistration(t *testing.T) {
	// After init() is called (which happens automatically when importing the package),
	// the provider should be registered.
	p, ok := crypto.Get("uk")
	if !ok {
		t.Fatal("Expected 'uk' provider to be registered, but it was not found")
	}
	if p == nil {
		t.Fatal("Expected 'uk' provider to be registered, but it was nil")
	}

	ukProvider, ok := p.(*uk.Provider)
	if !ok {
		t.Fatalf("Expected registered provider to be of type *uk.Provider, got %T", p)
	}

	// Perform a simple check to ensure it's the expected provider
	if ukProvider.Name() != "uk" {
		t.Errorf("Expected registered provider name 'uk', got '%s'", ukProvider.Name())
	}
}
