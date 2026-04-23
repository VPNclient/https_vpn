# Implementation Plan: Multi-Provider Certificate Selection

> Version: 1.0
> Status: REVIEW
> Last Updated: 2026-04-23
> Specifications: [02-specifications.md](02-specifications.md)

## Summary

Implement CertificateStore for automatic certificate selection based on client TLS capabilities. 4 tasks, 3 files modified/created.

## Task Breakdown

### Task 1: Add ParseProviderPriority to crypto/provider.go

**File**: `crypto/provider.go`

**Changes**:
```go
// ParseProviderPriority extracts provider names from cipherSuites config.
func ParseProviderPriority(cipherSuites string) []string
```

**Acceptance**:
- Parses "cn,ru,us" → ["cn", "ru", "us"]
- Skips unknown names (e.g., "TLS_AES_128_GCM_SHA256")
- Returns ["us"] if empty or no providers found
- Unit test passes

---

### Task 2: Create crypto/certstore.go

**File**: `crypto/certstore.go` (new)

**Components**:

```go
// CertificateStore holds certificates indexed by provider
type CertificateStore struct {
    byProvider  map[string]*tls.Certificate
    priority    []string
    defaultCert *tls.Certificate
    all         []tls.Certificate
}

// NewCertificateStore loads certificates and categorizes by provider
func NewCertificateStore(certs []conf.CertConfig, priority []string) (*CertificateStore, error)

// GetCertificate implements tls.Config.GetCertificate
func (cs *CertificateStore) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

// AllCertificates returns all loaded certificates
func (cs *CertificateStore) AllCertificates() []tls.Certificate

// detectProvider determines provider from certificate key type
func detectProvider(cert *tls.Certificate) string

// isSM4Suite returns true for Chinese SM4 cipher suites
func isSM4Suite(suite uint16) bool

// isGOSTSuite returns true for Russian GOST cipher suites
func isGOSTSuite(suite uint16) bool
```

**Acceptance**:
- Loads all certificates from config
- Detects provider from key type (RSA→us, SM2→cn, GOST→ru)
- GetCertificate selects based on ClientHello cipher suites
- Falls back to defaultCert if no match

---

### Task 3: Create crypto/certstore_test.go

**File**: `crypto/certstore_test.go` (new)

**Tests**:

```go
func TestParseProviderPriority(t *testing.T)
func TestIsSM4Suite(t *testing.T)
func TestIsGOSTSuite(t *testing.T)
func TestDetectProvider(t *testing.T)
func TestCertificateStoreSelection(t *testing.T)
func TestCertificateStoreFallback(t *testing.T)
```

**Acceptance**:
- All unit tests pass
- Coverage for priority parsing, suite detection, provider detection
- Mock certificates for selection tests

---

### Task 4: Integrate CertificateStore in core/core.go

**File**: `core/core.go`

**Change** (lines 98-106):

Before:
```go
if len(tlsSettings.Certificates) > 0 {
    cert := tlsSettings.Certificates[0]
    certPair, err := tls.LoadX509KeyPair(cert.CertificateFile, cert.KeyFile)
    ...
    tlsConfig.Certificates = []tls.Certificate{certPair}
}
```

After:
```go
if len(tlsSettings.Certificates) > 0 {
    priority := crypto.ParseProviderPriority(tlsSettings.CipherSuites)
    certStore, err := crypto.NewCertificateStore(tlsSettings.Certificates, priority)
    if err != nil {
        return fmt.Errorf("failed to load certificates: %w", err)
    }
    tlsConfig.GetCertificate = certStore.GetCertificate
    tlsConfig.Certificates = certStore.AllCertificates()
}
```

**Acceptance**:
- Single certificate configs still work
- Multiple certificates loaded and categorized
- GetCertificate callback set
- Existing tests pass

---

## Execution Order

```
Task 1 ──→ Task 2 ──→ Task 3 ──→ Task 4
  │           │          │          │
  │           │          │          └── Integration
  │           │          └── Unit tests
  │           └── Core implementation
  └── Foundation (priority parsing)
```

## Files Summary

| File | Action | Lines (est.) |
|------|--------|--------------|
| `crypto/provider.go` | Modify | +20 |
| `crypto/certstore.go` | Create | ~150 |
| `crypto/certstore_test.go` | Create | ~200 |
| `core/core.go` | Modify | +10, -8 |

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| SM2/GOST key detection fails | Use OID-based detection as fallback |
| GetCertificate not called | Also populate tlsConfig.Certificates |
| Circular import | CertStore in crypto pkg, uses conf.CertConfig |

## Testing Plan

1. **Unit tests** (Task 3): Core logic isolated
2. **Build test**: `go build ./...`
3. **Existing tests**: `go test ./...`
4. **Manual test**: Config with RSA + test SM4 ClientHello

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
