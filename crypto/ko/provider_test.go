package ko

import (
	"crypto/tls"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	kotls "github.com/nativemind/https-vpn/crypto/ko/tls"
)

func TestProviderRegistration(t *testing.T) {
	provider, ok := crypto.Get("ko")
	if !ok {
		t.Fatal("KO provider not registered")
	}

	if provider.Name() != "ko" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "ko")
	}
}

func TestProviderConfigureTLS(t *testing.T) {
	provider := &Provider{}
	cfg := &tls.Config{}

	err := provider.ConfigureTLS(cfg)
	if err != nil {
		t.Fatalf("ConfigureTLS failed: %v", err)
	}

	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d", cfg.MinVersion, tls.VersionTLS12)
	}

	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %d, want %d", cfg.MaxVersion, tls.VersionTLS13)
	}

	if len(cfg.CipherSuites) == 0 {
		t.Error("CipherSuites is empty")
	}

	if len(cfg.CurvePreferences) == 0 {
		t.Error("CurvePreferences is empty")
	}
}

func TestProviderSupportedCipherSuites(t *testing.T) {
	provider := &Provider{}
	suites := provider.SupportedCipherSuites()

	if len(suites) == 0 {
		t.Error("SupportedCipherSuites is empty")
	}

	// Check that ARIA suites are included
	hasARIA := false
	for _, s := range suites {
		if kotls.IsARIACipherSuite(s) {
			hasARIA = true
			break
		}
	}
	if !hasARIA {
		t.Error("No ARIA cipher suites in SupportedCipherSuites")
	}
}

func TestProviderDescription(t *testing.T) {
	provider := &Provider{}
	desc := provider.Description()

	if desc == "" {
		t.Error("Description is empty")
	}

	if len(desc) < 10 {
		t.Error("Description is too short")
	}
}

func TestProviderAlgorithms(t *testing.T) {
	provider := &Provider{}
	algs := provider.Algorithms()

	if len(algs) == 0 {
		t.Error("Algorithms is empty")
	}

	// Should include ARIA and SEED
	hasARIA := false
	hasSEED := false
	for _, a := range algs {
		if len(a) > 4 && a[:4] == "ARIA" {
			hasARIA = true
		}
		if len(a) > 4 && a[:4] == "SEED" {
			hasSEED = true
		}
	}
	if !hasARIA {
		t.Error("ARIA not in Algorithms")
	}
	if !hasSEED {
		t.Error("SEED not in Algorithms")
	}
}

func TestProviderIsPostQuantum(t *testing.T) {
	provider := &Provider{}

	if provider.IsPostQuantum() {
		t.Error("IsPostQuantum should be false")
	}
}

func TestProviderSecurityLevel(t *testing.T) {
	provider := &Provider{}

	level := provider.SecurityLevel()
	if level != 256 {
		t.Errorf("SecurityLevel = %d, want 256", level)
	}
}

func TestCipherSuiteDetection(t *testing.T) {
	tests := []struct {
		name    string
		suite   uint16
		isARIA  bool
		isSEED  bool
		isKO    bool
	}{
		{"ARIA-256-GCM", kotls.TLS_ARIA_256_GCM_SHA384, true, false, true},
		{"ARIA-128-GCM", kotls.TLS_ARIA_128_GCM_SHA256, true, false, true},
		{"SEED-CBC", kotls.TLS_RSA_WITH_SEED_CBC_SHA, false, true, true},
		{"AES-256", tls.TLS_AES_256_GCM_SHA384, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if kotls.IsARIACipherSuite(tt.suite) != tt.isARIA {
				t.Errorf("IsARIACipherSuite(%#x) = %v, want %v", tt.suite, !tt.isARIA, tt.isARIA)
			}
			if kotls.IsSEEDCipherSuite(tt.suite) != tt.isSEED {
				t.Errorf("IsSEEDCipherSuite(%#x) = %v, want %v", tt.suite, !tt.isSEED, tt.isSEED)
			}
			if kotls.IsKOCipherSuite(tt.suite) != tt.isKO {
				t.Errorf("IsKOCipherSuite(%#x) = %v, want %v", tt.suite, !tt.isKO, tt.isKO)
			}
		})
	}
}

func TestCipherSuiteInfo(t *testing.T) {
	info, ok := kotls.GetCipherSuiteInfo(kotls.TLS_ARIA_256_GCM_SHA384)
	if !ok {
		t.Fatal("GetCipherSuiteInfo failed for ARIA-256")
	}

	if info.KeySize != 256 {
		t.Errorf("KeySize = %d, want 256", info.KeySize)
	}

	if !info.IsAEAD {
		t.Error("IsAEAD should be true")
	}

	if !info.IsTLS13 {
		t.Error("IsTLS13 should be true")
	}
}
