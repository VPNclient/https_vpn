# Specifications: European Cryptography (Brainpool ECC)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Реализация европейской криптографии на базе кривых Brainpool (RFC 5639). Создаётся пакет `crypto/eu/` с реализацией кривых и провайдером для TLS.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `crypto/eu/brainpool/curves.go` | Create | Реализация кривых Brainpool P256r1, P384r1, P512r1 |
| `crypto/eu/brainpool/brainpool_test.go` | Create | Тесты для кривых с официальными векторами |
| `crypto/eu/provider.go` | Create | Регистрация провайдера "eu" |

## Architecture

### Component Diagram

```
crypto/eu/
├── provider.go              # Provider "eu" registration
└── brainpool/
    ├── curves.go            # Elliptic curve implementations (elliptic.Curve)
    ├── brainpool.go         # ECDSA and Key Exchange support
    └── brainpool_test.go
```

### Data Flow

```
Config: cipherSuites="eu"
         │
         ▼
crypto.Registry["eu"] → EUProvider
         │
         ▼
TLS Config: Brainpool certificate + AES-GCM cipher suites
         │
         ▼
Handshake: Brainpool key exchange + SHA-256/384 hash + AES encryption
```

## Interfaces

### New Interfaces / Components

```go
// crypto/eu/brainpool/curves.go

// P256r1 returns an elliptic.Curve implementing brainpoolP256r1.
func P256r1() elliptic.Curve

// P384r1 returns an elliptic.Curve implementing brainpoolP384r1.
func P384r1() elliptic.Curve

// P512r1 returns an elliptic.Curve implementing brainpoolP512r1.
func P512r1() elliptic.Curve
```

```go
// crypto/eu/provider.go

// Provider implements crypto.Provider for European cryptography.
type Provider struct{}

func (p *Provider) Name() string { return "eu" }

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
    // Configure Brainpool curves for TLS
    // Note: Standard Go crypto/tls may need extension to support these IDs
}

func (p *Provider) SupportedCipherSuites() []uint16 {
    // Reusing standard AES-GCM suites
    return []uint16{
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_AES_256_GCM_SHA384,
    }
}
```

## Data Models

### Brainpool P256r1 Parameters (RFC 5639)

```
p = A9FB57DBA1EEA9BC3E660A909D838D726E3BF623D52620282013481D1F6E5377
a = 7D5A0975FC2C3057EEF67530417AFFE7FB8055C126DC5C6CE94A4B44F330B5D9
b = 26DC5C6CE94A4B44F330B5D9BBD77CBF958416295CF7E748287AF60FACCB4A8D
x = 8BD2AEB9CB7E57CB2C4B482FFC81B7AFB9DE27E1E3BD23C23A4453BD9AD14695
y = 4803A05FD420F9D7AD5965A2158866B57388703798544A396BB8E1A169A1C81D
q = A9FB57DBA1EEA9BC3E660A909D838D718C397AA3B561A6F7D36E1085332497AF
h = 1
```

## Behavior Specifications

### Happy Path

1. Server loads config with `cipherSuites: "eu"`
2. `crypto.Get("eu")` returns EU provider
3. Provider configures TLS with Brainpool certificate
4. Client connects with TLS 1.3 + Brainpool curve
5. Handshake uses Brainpool P256r1 for key exchange
6. Data encrypted with AES-128-GCM

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Client doesn't support Brainpool | Handshake without BP curves | Handshake failure (if strict) or fallback |
| Unsupported Brainpool curve | Use of P160r1 | Return error |

## Testing Strategy

### Unit Tests

- [ ] `brainpool/brainpool_test.go` - Verify curve points, additions, and scalar multiplication using RFC 5639 test vectors.
- [ ] ECDSA Sign/Verify with Brainpool curves.

### Integration Tests

- [ ] Provider registration - `crypto.Get("eu")` returns valid provider.
- [ ] TLS configuration check - verify `CurvePreferences` include Brainpool IDs.

## Open Design Questions

- [ ] **TLS Support**: Go's `crypto/tls` has hardcoded list of supported curves in `tls13.go`. We might need to use a fork of `crypto/tls` or `github.com/google/go-tpm` style wrappers if we want full TLS 1.3 support for custom curves.
- [ ] **Curve Implementation**: Should we use `math/big` (slow) or optimized assembly for Brainpool? For Phase 3, `math/big` is acceptable.

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
