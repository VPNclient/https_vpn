package ua

import (
	"crypto/tls"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	uatls "github.com/nativemind/https-vpn/crypto/ua/tls"
)

func TestProviderRegistration(t *testing.T) {
	// Перевіряємо, що провайдер зареєстрований
	provider, ok := crypto.Get("ua")
	if !ok || provider == nil {
		t.Fatal("UA провайдер не зареєстрований")
	}

	if provider.Name() != "ua" {
		t.Errorf("Name() = %q, очікувано %q", provider.Name(), "ua")
	}
}

func TestProviderConfigureTLS(t *testing.T) {
	provider := &Provider{}
	cfg := &tls.Config{}

	err := provider.ConfigureTLS(cfg)
	if err != nil {
		t.Fatalf("ConfigureTLS() помилка: %v", err)
	}

	// Перевіряємо TLS версії
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("MinVersion = %d, очікувано TLS 1.3 (%d)", cfg.MinVersion, tls.VersionTLS13)
	}

	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %d, очікувано TLS 1.3 (%d)", cfg.MaxVersion, tls.VersionTLS13)
	}

	// Перевіряємо, що є cipher suites
	if len(cfg.CipherSuites) == 0 {
		t.Error("CipherSuites порожній")
	}

	// Перевіряємо, що є криві
	if len(cfg.CurvePreferences) == 0 {
		t.Error("CurvePreferences порожній")
	}
}

func TestProviderSupportedCipherSuites(t *testing.T) {
	provider := &Provider{}
	suites := provider.SupportedCipherSuites()

	if len(suites) == 0 {
		t.Error("SupportedCipherSuites() порожній")
	}

	// Перевіряємо наявність українських suites
	hasUASuite := false
	for _, suite := range suites {
		if uatls.IsUACipherSuite(suite) {
			hasUASuite = true
			break
		}
	}

	if !hasUASuite {
		t.Error("Немає українських cipher suites")
	}
}

func TestProviderDescription(t *testing.T) {
	provider := &Provider{}
	desc := provider.Description()

	if desc == "" {
		t.Error("Description() порожній")
	}
}

func TestProviderAlgorithms(t *testing.T) {
	provider := &Provider{}
	algs := provider.Algorithms()

	if len(algs) == 0 {
		t.Error("Algorithms() порожній")
	}

	// Перевіряємо наявність основних алгоритмів
	expectedAlgs := []string{"Калина", "Купина", "Мальва", "Сокіл"}
	for _, expected := range expectedAlgs {
		found := false
		for _, alg := range algs {
			if contains(alg, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Алгоритм %q не знайдено в списку", expected)
		}
	}
}

func TestProviderIsPostQuantum(t *testing.T) {
	provider := &Provider{}
	if !provider.IsPostQuantum() {
		t.Error("IsPostQuantum() має повертати true")
	}
}

func TestProviderSecurityLevel(t *testing.T) {
	provider := &Provider{}
	level := provider.SecurityLevel()

	if level < 128 {
		t.Errorf("SecurityLevel() = %d, очікувано >= 128", level)
	}
}

func TestCipherSuiteDetection(t *testing.T) {
	tests := []struct {
		suite    uint16
		expected bool
		name     string
	}{
		{uatls.TLS_UA_KALYNA_512_GCM_KUPYNA_512, true, "Калина-512"},
		{uatls.TLS_UA_KALYNA_256_GCM_KUPYNA_256, true, "Калина-256"},
		{tls.TLS_AES_256_GCM_SHA384, false, "AES-256"},
		{0x0000, false, "Нульовий"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := uatls.IsUACipherSuite(tc.suite)
			if result != tc.expected {
				t.Errorf("IsUACipherSuite(0x%04X) = %v, очікувано %v",
					tc.suite, result, tc.expected)
			}
		})
	}
}

func TestCurveDetection(t *testing.T) {
	tests := []struct {
		curve    tls.CurveID
		expected bool
		name     string
	}{
		{uatls.CurveX25519Malva, true, "X25519+Malva"},
		{uatls.CurveDSTU4145_512, true, "ДСТУ 4145-512"},
		{tls.X25519, false, "X25519"},
		{tls.CurveP256, false, "P-256"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := uatls.IsUACurve(tc.curve)
			if result != tc.expected {
				t.Errorf("IsUACurve(0x%04X) = %v, очікувано %v",
					tc.curve, result, tc.expected)
			}
		})
	}
}

func TestSignatureDetection(t *testing.T) {
	tests := []struct {
		sig      uint16
		expected bool
		name     string
	}{
		{uatls.SignatureSokil512, true, "Сокіл-512"},
		{uatls.SignatureDSTU4145_512, true, "ДСТУ 4145-512"},
		{0x0401, false, "RSA-PKCS1-SHA256"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := uatls.IsUASignature(tc.sig)
			if result != tc.expected {
				t.Errorf("IsUASignature(0x%04X) = %v, очікувано %v",
					tc.sig, result, tc.expected)
			}
		})
	}
}

func TestCipherSuiteInfo(t *testing.T) {
	info := uatls.GetCipherSuiteInfo(uatls.TLS_UA_KALYNA_512_GCM_KUPYNA_512)
	if info == nil {
		t.Fatal("GetCipherSuiteInfo() повернув nil для Калина-512")
	}

	if info.KeySize != 64 {
		t.Errorf("KeySize = %d, очікувано 64", info.KeySize)
	}

	if info.HashSize != 64 {
		t.Errorf("HashSize = %d, очікувано 64", info.HashSize)
	}

	if !info.IsPostQuatum {
		t.Error("IsPostQuantum має бути true")
	}
}

// contains перевіряє, чи містить рядок підрядок
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
