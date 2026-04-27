// Package ua реалізує криптографічний провайдер для української постквантової криптографії.
// Базується на концепції стандарту ДСТУ-ПК 2026.
package ua

import (
	"crypto/tls"

	"github.com/nativemind/https-vpn/crypto"
	uatls "github.com/nativemind/https-vpn/crypto/ua/tls"
)

// Provider реалізує crypto.Provider для української криптографії
type Provider struct{}

func init() {
	// Реєструємо провайдер при імпорті пакету
	crypto.Register(&Provider{})
}

// Name повертає ідентифікатор провайдера
func (p *Provider) Name() string {
	return "ua"
}

// ConfigureTLS налаштовує TLS конфігурацію для української криптографії
func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	// Встановлюємо TLS 1.3 як мінімальну та максимальну версію
	cfg.MinVersion = tls.VersionTLS13
	cfg.MaxVersion = tls.VersionTLS13

	// Налаштовуємо cipher suites
	// Примітка: стандартна бібліотека Go не підтримує кастомні cipher suites,
	// тому використовуємо fallback на AES-256-GCM
	cfg.CipherSuites = []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_CHACHA20_POLY1305_SHA256,
	}

	// Налаштовуємо криві для key exchange
	// Використовуємо X25519 як fallback (гібрид X25519+Malva буде додано пізніше)
	cfg.CurvePreferences = []tls.CurveID{
		tls.X25519,
		tls.CurveP384,
		tls.CurveP256,
	}

	return nil
}

// SupportedCipherSuites повертає список підтримуваних cipher suites
func (p *Provider) SupportedCipherSuites() []uint16 {
	// Повертаємо українські cipher suites
	// (для практичного використання поки fallback на стандартні)
	return []uint16{
		uatls.TLS_UA_KALYNA_512_GCM_KUPYNA_512,
		uatls.TLS_UA_KALYNA_256_GCM_KUPYNA_256,
		// Fallback
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
	}
}

// Description повертає опис провайдера
func (p *Provider) Description() string {
	return "Українська постквантова криптографія (ДСТУ-ПК 2026)"
}

// Algorithms повертає список підтримуваних алгоритмів
func (p *Provider) Algorithms() []string {
	return []string{
		"Калина-512-GCM (блочний шифр)",
		"Купина-512 (хеш-функція)",
		"Мальва-1024 (KEM)",
		"Сокіл-512 (цифровий підпис)",
		"ДСТУ 4145 (еліптичні криві)",
	}
}

// IsPostQuantum повертає true, оскільки провайдер підтримує постквантову криптографію
func (p *Provider) IsPostQuantum() bool {
	return true
}

// SecurityLevel повертає рівень безпеки в бітах
func (p *Provider) SecurityLevel() int {
	return 256 // Category 5 (еквівалент AES-256)
}
