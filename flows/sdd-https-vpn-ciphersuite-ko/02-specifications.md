# Specifications: Korean Ciphersuite (ARIA/SEED)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-28
> Requirements: [01-requirements.md](./01-requirements.md)

## Overview

Реализация корейского криптографического провайдера для HTTPS VPN с поддержкой национальных стандартов:
- **ARIA** - блочный шифр (KS X 1213:2004), аналог AES
- **SEED** - блочный шифр (KISA), legacy совместимость
- **LSH** - хеш-функция (опционально)

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `crypto/ko/` | Create | Новый пакет корейской криптографии |
| `crypto/ko/aria/` | Create | Реализация ARIA-128/256 |
| `crypto/ko/seed/` | Create | Реализация SEED-128 |
| `crypto/ko/tls/` | Create | TLS константы и cipher suites |
| `crypto/provider.go` | Modify | Добавить IsKOCryptoSuite() |

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    crypto/ko/                           │
├──────────────┬──────────────┬──────────────────────────┤
│   aria/      │    seed/     │         tls/             │
│  ┌────────┐  │  ┌────────┐  │  ┌────────────────────┐  │
│  │ ARIA   │  │  │ SEED   │  │  │ CipherSuites       │  │
│  │ 128/256│  │  │ 128    │  │  │ 0xC06A, 0xC06B...  │  │
│  └────────┘  │  └────────┘  │  └────────────────────┘  │
├──────────────┴──────────────┴──────────────────────────┤
│                    provider.go                          │
│           implements crypto.Provider                    │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

```
Client                                Server
  │                                     │
  │ ──── ClientHello ─────────────────► │
  │      cipher_suites: [0xC06A, ...]   │
  │                                     │
  │ ◄──── ServerHello ────────────────  │
  │       selected: 0xC06A              │
  │       (TLS_ARIA_256_GCM_SHA384)     │
  │                                     │
  │ ◄════ ARIA-256-GCM encrypted ═════► │
```

## Algorithms

### ARIA (Academy Research Institute Agency)

| Параметр | ARIA-128 | ARIA-192 | ARIA-256 |
|----------|----------|----------|----------|
| Блок | 128 бит | 128 бит | 128 бит |
| Ключ | 128 бит | 192 бит | 256 бит |
| Раунды | 12 | 14 | 16 |
| Структура | SPN | SPN | SPN |
| Стандарт | KS X 1213 | KS X 1213 | KS X 1213 |
| RFC | 5794, 6209, 9367 | 5794, 6209, 9367 | 5794, 6209, 9367 |

**Структура раунда ARIA:**
1. AddRoundKey - XOR с раундовым ключом
2. SubLayer - два типа S-box (S1, S2 чередуются)
3. DiffusionLayer - матричное преобразование

### SEED

| Параметр | Значение |
|----------|----------|
| Блок | 128 бит |
| Ключ | 128 бит |
| Раунды | 16 |
| Структура | Feistel |
| Стандарт | KISA |
| RFC | 4162, 4269 |

## Interfaces

### Provider Interface

```go
// crypto/ko/provider.go
package ko

type Provider struct{}

func (p *Provider) Name() string
func (p *Provider) ConfigureTLS(cfg *tls.Config) error
func (p *Provider) SupportedCipherSuites() []uint16
func (p *Provider) Description() string
func (p *Provider) Algorithms() []string
func (p *Provider) IsPostQuantum() bool
func (p *Provider) SecurityLevel() int
```

### ARIA Interface

```go
// crypto/ko/aria/aria.go
package aria

// NewCipher128 creates ARIA-128 cipher
func NewCipher128(key []byte) (cipher.Block, error)

// NewCipher192 creates ARIA-192 cipher
func NewCipher192(key []byte) (cipher.Block, error)

// NewCipher256 creates ARIA-256 cipher
func NewCipher256(key []byte) (cipher.Block, error)

// NewCipher auto-selects based on key size
func NewCipher(key []byte) (cipher.Block, error)
```

### SEED Interface

```go
// crypto/ko/seed/seed.go
package seed

// NewCipher creates SEED-128 cipher
func NewCipher(key []byte) (cipher.Block, error)
```

## TLS Cipher Suites

### TLS 1.3 (RFC 9367)

| ID | Name | AEAD | Hash |
|----|------|------|------|
| 0x1306 | TLS_ARIA_128_GCM_SHA256 | ARIA-128-GCM | SHA-256 |
| 0x1307 | TLS_ARIA_256_GCM_SHA384 | ARIA-256-GCM | SHA-384 |

### TLS 1.2 (RFC 6209)

| ID | Name | Key Exchange | Cipher | MAC |
|----|------|--------------|--------|-----|
| 0xC06A | TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256 | ECDHE | ARIA-128-GCM | SHA-256 |
| 0xC06B | TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384 | ECDHE | ARIA-256-GCM | SHA-384 |
| 0xC06C | TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256 | ECDHE | ARIA-128-GCM | SHA-256 |
| 0xC06D | TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384 | ECDHE | ARIA-256-GCM | SHA-384 |

### SEED (RFC 4162)

| ID | Name | Key Exchange | Cipher | MAC |
|----|------|--------------|--------|-----|
| 0x0096 | TLS_RSA_WITH_SEED_CBC_SHA | RSA | SEED-CBC | SHA-1 |

## Configuration

### config.ko.json

```json
{
  "inbounds": [{
    "port": 443,
    "protocol": "https-vpn",
    "streamSettings": {
      "security": "tls",
      "tlsSettings": {
        "cipherSuites": "ko",
        "certificates": [{
          "certificateFile": "/path/to/cert.pem",
          "keyFile": "/path/to/key.pem"
        }]
      }
    }
  }],
  "outbounds": [{"protocol": "freedom"}]
}
```

## Test Vectors

### ARIA Test Vectors (RFC 5794)

```
ARIA-128:
Key:       00112233445566778899aabbccddeeff
Plaintext: 11111111aaaaaaaa11111111bbbbbbbb
Ciphertext: c6ecd08e22c30abdb215cf74e2075e6e

ARIA-256:
Key:       00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff
Plaintext: 11111111aaaaaaaa11111111bbbbbbbb
Ciphertext: 8d1470625f59ebacb0e55b534b3e462b
```

### SEED Test Vectors (RFC 4269)

```
SEED-128:
Key:       00000000000000000000000000000000
Plaintext: 000102030405060708090A0B0C0D0E0F
Ciphertext: 5EBAC6E0054E166819AFF1CC6D346CDB
```

## Testing Strategy

### Unit Tests

- [ ] ARIA-128 encrypt/decrypt with RFC vectors
- [ ] ARIA-256 encrypt/decrypt with RFC vectors
- [ ] SEED encrypt/decrypt with RFC vectors
- [ ] Provider registration
- [ ] TLS configuration

### Integration Tests

- [ ] TLS handshake with ARIA cipher suite
- [ ] Data encryption/decryption through VPN tunnel

## File Structure

```
crypto/ko/
├── provider.go          # KO provider implementation
├── provider_test.go     # Provider tests
├── README.md            # Documentation (Korean/English)
├── aria/
│   ├── aria.go          # ARIA cipher implementation
│   ├── aria_test.go     # ARIA tests with RFC vectors
│   ├── consts.go        # S-boxes and constants
│   └── tables.go        # Precomputed tables (optional)
├── seed/
│   ├── seed.go          # SEED cipher implementation
│   ├── seed_test.go     # SEED tests with RFC vectors
│   └── consts.go        # S-boxes and constants
└── tls/
    └── cipher_suites.go # TLS cipher suite constants
```

## Open Design Questions

- [x] Use external library or implement from scratch? → Implement from scratch for audit
- [x] Support TLS 1.2 or only TLS 1.3? → Both (TLS 1.3 primary, TLS 1.2 for legacy)
- [ ] Include LSH hash function? → Optional, SHA-2/SHA-3 preferred

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
