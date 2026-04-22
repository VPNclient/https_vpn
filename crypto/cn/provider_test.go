package cn

import (
	"crypto/tls"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	smtls "github.com/nativemind/https-vpn/crypto/cn/tls"
)

func TestProviderRegistration(t *testing.T) {
	// Provider should be registered via init()
	p, ok := crypto.Get("cn")
	if !ok {
		t.Fatal("CN provider not registered")
	}

	if p.Name() != "cn" {
		t.Errorf("Name() = %s, want cn", p.Name())
	}
}

func TestProviderInList(t *testing.T) {
	providers := crypto.List()
	found := false
	for _, name := range providers {
		if name == "cn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("CN provider not in crypto.List()")
	}
}

func TestSupportedCipherSuites(t *testing.T) {
	p, _ := crypto.Get("cn")

	suites := p.SupportedCipherSuites()
	if len(suites) != 2 {
		t.Errorf("len(SupportedCipherSuites()) = %d, want 2", len(suites))
	}

	// Check for expected cipher suites
	expected := map[uint16]bool{
		smtls.TLS_SM4_GCM_SM3: false,
		smtls.TLS_SM4_CCM_SM3: false,
	}

	for _, suite := range suites {
		if _, ok := expected[suite]; ok {
			expected[suite] = true
		}
	}

	for suite, found := range expected {
		if !found {
			t.Errorf("cipher suite 0x%04X not in SupportedCipherSuites()", suite)
		}
	}
}

func TestConfigureTLS(t *testing.T) {
	p, _ := crypto.Get("cn")

	cfg := &tls.Config{}
	err := p.ConfigureTLS(cfg)
	if err != nil {
		t.Fatalf("ConfigureTLS failed: %v", err)
	}

	// SM cipher suites require TLS 1.3
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("MinVersion = 0x%04X, want TLS 1.3 (0x%04X)", cfg.MinVersion, tls.VersionTLS13)
	}
}
