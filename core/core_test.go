package core

import (
	"testing"

	"github.com/nativemind/https-vpn/infra/conf"
	_ "github.com/nativemind/https-vpn/crypto/us"
)

func TestGetProviderName_NilConfig(t *testing.T) {
	name, deprecated := getProviderName(nil)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false for nil config")
	}
}

func TestGetProviderName_EmptyConfig(t *testing.T) {
	cfg := &conf.TLSConfig{}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false for empty config")
	}
}

func TestGetProviderName_CipherSuites_SingleProvider(t *testing.T) {
	cfg := &conf.TLSConfig{
		CipherSuites: "us",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false when using CipherSuites")
	}
}

func TestGetProviderName_CipherSuites_WithCipherNames(t *testing.T) {
	// Provider identifier mixed with standard cipher names
	cfg := &conf.TLSConfig{
		CipherSuites: "us,TLS_AES_256_GCM_SHA384",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false")
	}
}

func TestGetProviderName_CipherSuites_OnlyCipherNames(t *testing.T) {
	// Only standard cipher names, no provider identifier
	cfg := &conf.TLSConfig{
		CipherSuites: "TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us' (default), got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false")
	}
}

func TestGetProviderName_DeprecatedCryptoProvider(t *testing.T) {
	cfg := &conf.TLSConfig{
		CryptoProvider: "us",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if !deprecated {
		t.Error("Expected deprecated=true when using CryptoProvider field")
	}
}

func TestGetProviderName_CipherSuites_TakesPrecedence(t *testing.T) {
	// CipherSuites should take precedence over CryptoProvider
	cfg := &conf.TLSConfig{
		CipherSuites:   "us",
		CryptoProvider: "us", // This should be ignored
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false when CipherSuites has valid provider")
	}
}

func TestGetProviderName_CaseInsensitive(t *testing.T) {
	cfg := &conf.TLSConfig{
		CipherSuites: "US",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false")
	}
}

func TestGetProviderName_WhitespaceHandling(t *testing.T) {
	cfg := &conf.TLSConfig{
		CipherSuites: "  us  ,  TLS_AES_256  ",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false")
	}
}

func TestGetProviderName_UnknownProvider(t *testing.T) {
	cfg := &conf.TLSConfig{
		CipherSuites: "xyz",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us' (default), got '%s'", name)
	}
	if deprecated {
		t.Error("Expected deprecated=false")
	}
}

func TestGetProviderName_FallbackToDeprecated(t *testing.T) {
	// CipherSuites has no valid provider, fallback to CryptoProvider
	cfg := &conf.TLSConfig{
		CipherSuites:   "TLS_AES_256_GCM_SHA384",
		CryptoProvider: "us",
	}
	name, deprecated := getProviderName(cfg)
	if name != "us" {
		t.Errorf("Expected 'us', got '%s'", name)
	}
	if !deprecated {
		t.Error("Expected deprecated=true when falling back to CryptoProvider")
	}
}
