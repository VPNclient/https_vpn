# Specifications: Multi-Provider Certificate Selection

> Version: 1.0
> Status: REVIEW
> Last Updated: 2026-04-23
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Implement automatic certificate selection based on client TLS capabilities. The server loads multiple certificates (SM2, GOST, RSA/ECDSA) and uses Go's `GetCertificate` callback to select the appropriate one during TLS handshake.

## Architecture

### Current State

```
Config → Load Certificates[0] → tlsConfig.Certificates
                                      ↓
                           Static certificate for all clients
```

### Target State

```
Config → Load All Certificates → CertificateStore
                                      ↓
                              GetCertificate callback
                                      ↓
                          Inspect ClientHello cipher suites
                                      ↓
                          Match to certificate by provider
                                      ↓
                          Return matching certificate
```

## Components

### 1. CertificateStore

New type to hold categorized certificates.

**Location**: `crypto/certstore.go`

```go
// CertificateStore holds certificates categorized by crypto provider.
type CertificateStore struct {
    // Certificates indexed by provider name ("cn", "ru", "us")
    byProvider map[string]*tls.Certificate

    // Priority order for selection (from cipherSuites config)
    priority []string

    // Default certificate (first loaded, for fallback)
    defaultCert *tls.Certificate
}

// NewCertificateStore creates store from config.
func NewCertificateStore(certs []conf.CertConfig, priority []string) (*CertificateStore, error)

// GetCertificate implements tls.Config.GetCertificate callback.
func (cs *CertificateStore) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

// detectProvider determines provider from certificate key type.
func detectProvider(cert *tls.Certificate) string
```

### 2. Provider Detection

Map certificate key algorithm to provider:

| Key Algorithm | OID / Curve | Provider |
|--------------|-------------|----------|
| RSA | - | "us" |
| ECDSA P-256 | 1.2.840.10045.3.1.7 | "us" |
| ECDSA P-384 | 1.3.132.0.34 | "us" |
| SM2 | 1.2.156.10197.1.301 | "cn" |
| GOST R 34.10-2012 | 1.2.643.7.1.1.1.1 | "ru" |

**Implementation**:

```go
func detectProvider(cert *tls.Certificate) string {
    leaf := cert.Leaf
    if leaf == nil {
        // Parse certificate if Leaf not populated
        leaf, _ = x509.ParseCertificate(cert.Certificate[0])
    }

    switch pub := leaf.PublicKey.(type) {
    case *rsa.PublicKey:
        return "us"
    case *ecdsa.PublicKey:
        switch pub.Curve {
        case sm2.P256Sm2():
            return "cn"
        default:
            return "us"
        }
    case *gost.PublicKey:
        return "ru"
    default:
        return "us"
    }
}
```

### 3. Client Capability Detection

Determine client's crypto capabilities from ClientHello:

```go
func (cs *CertificateStore) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
    // Build set of providers client supports based on cipher suites
    clientProviders := make(map[string]bool)

    for _, suite := range hello.CipherSuites {
        switch {
        case isSM4Suite(suite):
            clientProviders["cn"] = true
        case isGOSTSuite(suite):
            clientProviders["ru"] = true
        default:
            clientProviders["us"] = true
        }
    }

    // Select certificate using priority order
    for _, provider := range cs.priority {
        if clientProviders[provider] {
            if cert, ok := cs.byProvider[provider]; ok {
                return cert, nil
            }
        }
    }

    // Fallback to default
    return cs.defaultCert, nil
}
```

### 4. Cipher Suite Classification

```go
// isSM4Suite returns true for Chinese SM4 cipher suites (RFC 8998)
func isSM4Suite(suite uint16) bool {
    return suite == 0x00C6 || // TLS_SM4_GCM_SM3
           suite == 0x00C7    // TLS_SM4_CCM_SM3
}

// isGOSTSuite returns true for Russian GOST cipher suites
func isGOSTSuite(suite uint16) bool {
    // GOST cipher suite IDs (based on implementation)
    return suite >= 0xFF85 && suite <= 0xFF88
}
```

## Integration Points

### core/core.go Changes

**Before** (lines 98-106):
```go
if len(tlsSettings.Certificates) > 0 {
    cert := tlsSettings.Certificates[0]
    certPair, err := tls.LoadX509KeyPair(cert.CertificateFile, cert.KeyFile)
    if err != nil {
        return fmt.Errorf("failed to load certificate: %w", err)
    }
    tlsConfig.Certificates = []tls.Certificate{certPair}
}
```

**After**:
```go
if len(tlsSettings.Certificates) > 0 {
    // Parse priority from cipherSuites
    priority := crypto.ParseProviderPriority(tlsSettings.CipherSuites)

    // Create certificate store with all certificates
    certStore, err := crypto.NewCertificateStore(tlsSettings.Certificates, priority)
    if err != nil {
        return fmt.Errorf("failed to load certificates: %w", err)
    }

    // Set callback for dynamic selection
    tlsConfig.GetCertificate = certStore.GetCertificate

    // Also set Certificates for clients that don't trigger callback
    tlsConfig.Certificates = certStore.AllCertificates()
}
```

## Data Structures

### CertConfig (unchanged)

```go
type CertConfig struct {
    CertificateFile string `json:"certificateFile"`
    KeyFile         string `json:"keyFile"`
}
```

### Priority Parsing

```go
// ParseProviderPriority extracts provider names from cipherSuites.
// Input: "cn,ru,us" or "cn,TLS_AES_128_GCM_SHA256,ru"
// Output: ["cn", "ru", "us"]
func ParseProviderPriority(cipherSuites string) []string {
    var priority []string
    seen := make(map[string]bool)

    for _, part := range strings.Split(cipherSuites, ",") {
        name := strings.TrimSpace(strings.ToLower(part))
        if _, ok := Registry[name]; ok && !seen[name] {
            priority = append(priority, name)
            seen[name] = true
        }
    }

    // Default fallback
    if len(priority) == 0 {
        priority = []string{"us"}
    }

    return priority
}
```

## Edge Cases

### 1. Single Certificate Config

If only one certificate is configured, behavior is unchanged - that certificate is used for all connections.

### 2. No Matching Certificate

If client doesn't support any provider for which we have a certificate, use the first certificate loaded (defaultCert).

### 3. Certificate Without Matching Provider

If a certificate's key type doesn't map to any known provider, categorize as "us" (default).

### 4. Empty CipherSuites

If `cipherSuites` is empty, priority defaults to `["us"]`.

### 5. Unsupported Key Type in Certificate

Log warning and skip certificate during store initialization.

## Logging

```go
// On certificate selection
log.Printf("Certificate selected: provider=%s sni=%s", provider, hello.ServerName)

// On store initialization
log.Printf("Loaded %d certificates: %v", len(store.byProvider), providerNames)
```

## Error Handling

| Condition | Behavior |
|-----------|----------|
| Certificate file not found | Return error on Start() |
| Invalid certificate format | Return error on Start() |
| No certificates configured | Allow (use default TLS) |
| Key type detection fails | Categorize as "us", log warning |
| GetCertificate returns nil | TLS library uses Certificates[0] |

## Testing Strategy

### Unit Tests

1. `TestDetectProvider` - verify key type → provider mapping
2. `TestCertificateStoreSelection` - verify selection logic
3. `TestParseProviderPriority` - verify priority parsing
4. `TestCipherSuiteClassification` - verify SM4/GOST detection

### Integration Tests

1. Mock ClientHello with SM4 suites → verify SM2 cert selected
2. Mock ClientHello with GOST suites → verify GOST cert selected
3. Mock ClientHello with standard suites → verify RSA/ECDSA cert selected
4. Single certificate fallback → verify backward compatibility

## Files to Modify

| File | Change |
|------|--------|
| `crypto/certstore.go` | **New** - CertificateStore type |
| `crypto/provider.go` | Add ParseProviderPriority() |
| `core/core.go` | Use CertificateStore instead of single cert |
| `crypto/certstore_test.go` | **New** - Unit tests |

## Backward Compatibility

- Single certificate configs work unchanged
- Empty `cipherSuites` defaults to "us" provider
- `cryptoProvider` deprecated field still works

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
