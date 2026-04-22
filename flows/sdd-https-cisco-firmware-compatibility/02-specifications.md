# Specifications: https-cisco-firmware-compatibility

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-21
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Implement a GOST crypto provider (`crypto/ru`) that enables HTTP/2 VPN over TLS with Russian national cryptographic algorithms. This requires a custom TLS implementation since Go's standard `crypto/tls` does not support GOST cipher suites.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `crypto/ru/` | Create | New GOST crypto provider package |
| `crypto/ru/gost/` | Create | GOST primitives (Kuznyechik, Magma, Streebog, GOST 34.10) |
| `crypto/ru/tls/` | Create | Custom TLS 1.3 implementation with GOST cipher suites |
| `crypto/provider.go` | Modify | Add `TLSListener` method to Provider interface |
| `transport/server.go` | Modify | Support custom TLS listener from provider |

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTPS VPN                               │
├─────────────────────────────────────────────────────────────────┤
│  core/core.go                                                   │
│    └── selects crypto provider based on config                  │
├─────────────────────────────────────────────────────────────────┤
│  transport/server.go                                            │
│    └── H2Server uses provider.TLSListener() or stdlib tls      │
├───────────────────────┬─────────────────────────────────────────┤
│  crypto/us/           │  crypto/ru/                             │
│  (US/NIST - stdlib)   │  (GOST - custom TLS)                    │
│                       │                                         │
│  - ConfigureTLS()     │  - ConfigureTLS()                       │
│  - TLSListener():nil  │  - TLSListener(): custom                │
│                       │                                         │
│                       │  ┌─────────────────────────────────┐   │
│                       │  │ crypto/ru/tls/                  │   │
│                       │  │  - TLS 1.3 handshake            │   │
│                       │  │  - GOST cipher suites           │   │
│                       │  │  - Certificate handling         │   │
│                       │  └─────────────────────────────────┘   │
│                       │                                         │
│                       │  ┌─────────────────────────────────┐   │
│                       │  │ crypto/ru/gost/                 │   │
│                       │  │  - kuznyechik.go (Grasshopper)  │   │
│                       │  │  - magma.go                     │   │
│                       │  │  - streebog.go (hash)           │   │
│                       │  │  - gost3410.go (signatures)     │   │
│                       │  │  - curves.go (elliptic curves)  │   │
│                       │  └─────────────────────────────────┘   │
└───────────────────────┴─────────────────────────────────────────┘
```

### Data Flow

```
Client (Cisco/OpenConnect with GOST)
    │
    ▼
TCP Connection
    │
    ▼
┌─────────────────────────────────────┐
│ TLS 1.3 Handshake (GOST)            │
│  - ClientHello with GOST suites     │
│  - ServerHello selects GOST suite   │
│  - Certificate (GOST R 34.10-2012)  │
│  - Key Exchange (GOST DH/ECDH)      │
│  - Finished (Streebog MAC)          │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│ HTTP/2 over encrypted channel       │
│  - CONNECT request                  │
│  - Bidirectional stream             │
│  - Data encrypted with Kuznyechik   │
└─────────────────────────────────────┘
    │
    ▼
Target Server
```

## Interfaces

### New Interfaces

#### Extended Provider Interface

```go
// crypto/provider.go
type Provider interface {
    Name() string
    ConfigureTLS(cfg *tls.Config) error
    SupportedCipherSuites() []uint16

    // New method for custom TLS implementations
    // Returns nil if provider uses stdlib tls
    TLSListener(inner net.Listener, config *tls.Config) (net.Listener, error)
}
```

#### GOST Cipher Interface

```go
// crypto/ru/gost/cipher.go
type BlockCipher interface {
    cipher.Block
    BlockSize() int  // 16 for Kuznyechik, 8 for Magma
    KeySize() int    // 32 for both
}

// Kuznyechik (Grasshopper) - GOST R 34.12-2015
func NewKuznyechik(key []byte) (BlockCipher, error)

// Magma - GOST R 34.12-2015
func NewMagma(key []byte) (BlockCipher, error)
```

#### GOST Hash Interface

```go
// crypto/ru/gost/hash.go
// Streebog - GOST R 34.11-2012
func NewStreebog256() hash.Hash
func NewStreebog512() hash.Hash
```

#### GOST Signature Interface

```go
// crypto/ru/gost/signature.go
type PrivateKey struct {
    D      *big.Int
    Curve  *Curve
    Public *PublicKey
}

type PublicKey struct {
    X, Y  *big.Int
    Curve *Curve
}

// GOST R 34.10-2012
func Sign(rand io.Reader, priv *PrivateKey, hash []byte) ([]byte, error)
func Verify(pub *PublicKey, hash, sig []byte) bool

// Key generation
func GenerateKey(curve *Curve, rand io.Reader) (*PrivateKey, error)
```

### Modified Interfaces

#### ServerConfig Extension

```go
// transport/server.go
type ServerConfig struct {
    Addr           string
    TLSConfig      *tls.Config
    CryptoProvider string
    Handler        http.Handler
    // New: allows provider to supply custom TLS listener
    CustomListener net.Listener
}
```

## Data Models

### New Types

#### GOST Elliptic Curves

```go
// crypto/ru/gost/curves.go
// Per GOST R 34.10-2012

// 256-bit curves
var (
    CurveIdGostR34102001CryptoProA *Curve  // id-GostR3410-2001-CryptoPro-A-ParamSet
    CurveIdGostR34102001CryptoProB *Curve  // id-GostR3410-2001-CryptoPro-B-ParamSet
    CurveIdtc26gost341012256A      *Curve  // id-tc26-gost-3410-12-256-paramSetA
)

// 512-bit curves
var (
    CurveIdtc26gost341012512A *Curve  // id-tc26-gost-3410-12-512-paramSetA
    CurveIdtc26gost341012512B *Curve  // id-tc26-gost-3410-12-512-paramSetB
    CurveIdtc26gost341012512C *Curve  // id-tc26-gost-3410-12-512-paramSetC
)
```

#### TLS Cipher Suite IDs

```go
// crypto/ru/tls/cipher_suites.go
// Per RFC 9189

const (
    // TLS 1.3 GOST cipher suites
    TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L uint16 = 0xC103
    TLS_GOSTR341112_256_WITH_MAGMA_MGM_L      uint16 = 0xC104
    TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_S uint16 = 0xC105
    TLS_GOSTR341112_256_WITH_MAGMA_MGM_S      uint16 = 0xC106

    // With 512-bit signatures
    TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L uint16 = 0xC107
    TLS_GOSTR341112_512_WITH_MAGMA_MGM_L      uint16 = 0xC108
    TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_S uint16 = 0xC109
    TLS_GOSTR341112_512_WITH_MAGMA_MGM_S      uint16 = 0xC10A
)
```

#### Certificate Types

```go
// crypto/ru/tls/cert.go
// GOST certificate with OIDs per RFC 4491

type GOSTCertificate struct {
    Raw          []byte
    PublicKey    *gost.PublicKey
    SignatureAlg x509.SignatureAlgorithm  // GOST R 34.10-2012
}

// Load GOST certificate from PEM/DER
func LoadGOSTCertificate(certFile, keyFile string) (*GOSTCertificate, *gost.PrivateKey, error)
```

## Behavior Specifications

### Happy Path

1. Server starts with `cipherSuites: "ru"` in config
2. `crypto.Get("ru")` returns GOST provider
3. Provider's `TLSListener()` returns custom GOST TLS listener
4. Client connects with GOST-enabled TLS
5. TLS handshake negotiates GOST cipher suite (e.g., Kuznyechik-MGM)
6. HTTP/2 CONNECT request flows over encrypted channel
7. VPN tunnel established

### TLS Handshake Flow (GOST TLS 1.3)

1. **ClientHello**
   - Supported cipher suites include GOST
   - Supported groups include GOST curves
   - Signature algorithms include GOST R 34.10-2012

2. **ServerHello**
   - Selected cipher suite: `TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L`
   - Key share using GOST curve

3. **EncryptedExtensions**
   - ALPN: `h2`

4. **Certificate**
   - Server certificate with GOST R 34.10-2012 public key

5. **CertificateVerify**
   - Signature using GOST R 34.10-2012 with Streebog hash

6. **Finished**
   - MAC using Streebog-256

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Non-GOST client connects to GOST server | Client sends only AES suites | Connection rejected (no common cipher) OR fallback if dual-mode enabled |
| Invalid GOST certificate | Corrupted cert file | Server fails to start with clear error |
| Unsupported GOST curve | Client requests unavailable curve | Handshake fails with appropriate alert |
| TLS 1.2 client | Client doesn't support TLS 1.3 | Fallback to TLS 1.2 GOST suites if supported |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| `ErrGOSTCertificateRequired` | No certificate configured for GOST | Server startup fails |
| `ErrInvalidGOSTKey` | Key doesn't match GOST parameters | Certificate load fails |
| `ErrCipherSuiteNotSupported` | Client/server cipher mismatch | TLS alert: handshake_failure |
| `ErrStreebogHashFailed` | Hash computation error | Internal error, connection closed |

## Dependencies

### External Libraries

| Library | Purpose | License |
|---------|---------|---------|
| `github.com/bi-zone/gost` (or similar) | GOST primitives reference | Check license |

**Note**: May need to implement GOST primitives from scratch for full control and auditability.

### Requires

- Go 1.21+ (for crypto improvements)
- GOST certificates (for testing)

### Blocks

- Nothing (this is an additive feature)

## Integration Points

### External Systems

- **Cisco Firmware**: Must be tested with actual Cisco devices running GOST-enabled firmware
- **OpenConnect**: Compatible client for testing
- **Russian CA certificates**: For production use

### Internal Systems

- `crypto/provider.go`: Extended interface
- `transport/server.go`: Custom listener support
- `core/core.go`: Provider selection logic

## Testing Strategy

### Unit Tests

- [ ] `crypto/ru/gost/kuznyechik_test.go` - Test vectors from GOST R 34.12-2015
- [ ] `crypto/ru/gost/magma_test.go` - Test vectors from GOST R 34.12-2015
- [ ] `crypto/ru/gost/streebog_test.go` - Test vectors from GOST R 34.11-2012
- [ ] `crypto/ru/gost/gost3410_test.go` - Signature test vectors
- [ ] `crypto/ru/tls/handshake_test.go` - TLS handshake simulation

### Integration Tests

- [ ] Full TLS handshake with GOST cipher suites
- [ ] HTTP/2 CONNECT over GOST TLS
- [ ] Certificate chain validation
- [ ] Mixed environment (GOST server, non-GOST client rejection)

### Manual Verification

- [ ] Test with OpenConnect client configured for GOST
- [ ] Test with actual Cisco device (if available)
- [ ] Verify traffic with Wireshark (check cipher suite negotiation)
- [ ] Performance benchmark vs AES

## Migration / Rollout

1. **Phase 1**: Implement GOST primitives with test vectors
2. **Phase 2**: Implement custom TLS layer
3. **Phase 3**: Integration with existing provider system
4. **Phase 4**: Testing with real clients
5. **Phase 5**: Documentation and release

No data migration required - this is a new feature.

## Open Design Questions

- [x] Custom TLS implementation vs patching stdlib - **Decision: Custom implementation required**
- [ ] Use existing GOST library or implement from scratch? - Recommend implementing from scratch for auditability
- [ ] Support TLS 1.2 GOST suites or TLS 1.3 only? - Recommend TLS 1.3 first, TLS 1.2 as follow-up
- [ ] How to handle dual-mode (GOST + AES on same port)? - Separate providers, user configures which to use

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-21
- [x] Notes: Approved with custom TLS implementation approach
