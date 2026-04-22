# Specifications: UK NCSC Cryptography Compliance

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Implementation of the UK NCSC compliant crypto provider. This provider uses standard Go `crypto/tls` but configures it strictly according to NCSC guidelines for TLS 1.3.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `crypto/uk/provider.go` | Create | New provider "uk" implementation |
| `crypto/crypto.go` (registry) | Modify | Implicitly via `init()` in new package |
| `README.md` | Modify | Update UK status to "supported" |

## Architecture

### Component Diagram

```
crypto/
├── provider.go              # Provider interface
├── us/                      # NIST (Go stdlib)
├── ru/                      # GOST provider
├── cn/                      # SM provider
└── uk/                      # NEW: NCSC provider
    └── provider.go          # Strict TLS 1.3 config
```

### Data Flow

```
Config: cipherSuites="uk"
         │
         ▼
crypto.Registry["uk"] → UKProvider
         │
         ▼
TLS Config: 
 - MinVersion: TLS 1.3
 - CipherSuites: [TLS_AES_256_GCM_SHA384, TLS_AES_128_GCM_SHA256]
 - CurvePreferences: [CurveP384, CurveP256]
```

## Interfaces

### New Interfaces

None. Uses existing `crypto.Provider` interface.

```go
type Provider interface {
    Name() string
    ConfigureTLS(cfg *tls.Config) error
    SupportedCipherSuites() []uint16
}
```

## Data Models

### UK Provider implementation

```go
package uk

import (
	"crypto/tls"
	"github.com/nativemind/https-vpn/crypto"
)

type Provider struct{}

func (p *Provider) Name() string { return "uk" }

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
	cfg.MinVersion = tls.VersionTLS13
	cfg.MaxVersion = tls.VersionTLS13
	cfg.CipherSuites = p.SupportedCipherSuites()
	cfg.CurvePreferences = []tls.CurveID{
		tls.CurveP384,
		tls.CurveP256,
	}
	// PreferServerCipherSuites is deprecated in TLS 1.3 but kept for completeness
	cfg.PreferServerCipherSuites = true 
	return nil
}

func (p *Provider) SupportedCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
	}
}

func init() {
	crypto.Register(&Provider{})
}
```

## Behavior Specifications

### Happy Path

1. Server loads config with `cipherSuites: "uk"`
2. `crypto.Get("uk")` returns UK provider
3. Provider configures TLS 1.3 with AES-256-GCM-SHA384 preference
4. Client connects with TLS 1.3
5. Handshake negotiates P-384 and AES-256-GCM-SHA384

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Client supports TLS 1.2 only | Handshake start | Connection rejected (Protocol version mismatch) |
| Client doesn't support P-384 | Handshake start | Negotiates P-256 (fallback per NCSC) |
| Client supports only ChaCha20 | Handshake start | Connection rejected (No overlapping cipher suites) |

### Error Handling

Standard TLS error handling provided by `crypto/tls`.

## Dependencies

### Requires

- `crypto/provider.go` - Provider interface (exists)
- Go `crypto/tls` standard library

## Integration Points

### Internal Systems

- `crypto.Registry` - Provider registration
- `infra/conf` - Config parsing (already supports `cipherSuites` string)

## Testing Strategy

### Unit Tests

- [ ] `crypto/uk/provider_test.go` - Verify `ConfigureTLS` sets expected fields.
- [ ] Registry test - Verify "uk" is registered when imported.

### Integration Tests

- [ ] End-to-end handshake test with a client restricted to AES-256-GCM-SHA384.

## Migration / Rollout

1. Create `crypto/uk/provider.go`.
2. Update `README.md` table.
3. Verify with tests.

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
