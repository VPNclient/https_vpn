# Specifications: French National Cryptography (ANSSI)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22

## Architectural Overview

The "fr" provider will be implemented as a new package under `crypto/fr/`, following the structure of the existing `crypto/ru/gost/` and `crypto/cn/` implementations.

## Implementation Details

### Provider Interface
The `crypto/fr` package must implement the `crypto.Provider` interface:

```go
package fr

type Provider struct{}

func (p *Provider) Name() string { return "fr" }
// ... other interface methods
```

### Configuration
The configuration loader in `infra/conf/loader.go` needs to be updated to recognize "fr" in the `cipherSuites` field.

### TLS Configuration
The TLS configuration will restrict algorithms to those approved by ANSSI (typically high-security AES-GCM and approved ECC curves).

## Security Policy

- Enforce TLS 1.3.
- Disable weak cipher suites and curves.
- Prioritize high-strength key exchanges.

## Dependencies

- None (Standard Go crypto library usage).

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
