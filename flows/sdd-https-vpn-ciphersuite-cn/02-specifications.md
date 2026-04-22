# Specifications: Chinese National Cryptography (SM Series)

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-04-22
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Реализация китайской национальной криптографии по аналогии с существующей GOST-реализацией (`crypto/ru/`). Создаётся пакет `crypto/cn/` с подпакетами для каждого алгоритма.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `crypto/cn/sm2/curve.go` | Create | SM2 эллиптическая кривая (SM2-P256) |
| `crypto/cn/sm2/sm2.go` | Create | SM2 подписи и шифрование |
| `crypto/cn/sm3/sm3.go` | Create | SM3 хэш-функция (256 бит) |
| `crypto/cn/sm4/sm4.go` | Create | SM4 блочный шифр (128 бит) |
| `crypto/cn/sm9/sm9.go` | Create | SM9 identity-based криптография |
| `crypto/cn/tls/cipher_suites.go` | Create | TLS cipher suites для SM |
| `crypto/cn/provider.go` | Create | Регистрация провайдера "cn" |

## Architecture

### Component Diagram

```
crypto/cn/
├── provider.go              # Provider "cn" registration
├── sm2/
│   ├── curve.go             # SM2-P256 curve parameters
│   ├── sm2.go               # Sign/Verify, Encrypt/Decrypt
│   └── sm2_test.go
├── sm3/
│   ├── sm3.go               # Hash implementation
│   └── sm3_test.go
├── sm4/
│   ├── sm4.go               # Block cipher
│   ├── modes.go             # GCM, CCM modes
│   └── sm4_test.go
├── sm9/
│   ├── sm9.go               # ID-based crypto
│   ├── bn256.go             # BN256 pairing curve
│   └── sm9_test.go
└── tls/
    ├── cipher_suites.go     # TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3
    └── handshake.go         # SM TLS handshake support
```

### Data Flow

```
Config: cipherSuites="cn"
         │
         ▼
crypto.Registry["cn"] → CNProvider
         │
         ▼
TLS Config: SM2 certificate + SM cipher suites
         │
         ▼
Handshake: SM2 key exchange + SM3 hash + SM4 encryption
```

## Interfaces

### New Interfaces

```go
// crypto/cn/sm2/sm2.go

// PrivateKey represents an SM2 private key.
type PrivateKey struct {
    PublicKey
    D *big.Int
}

// PublicKey represents an SM2 public key.
type PublicKey struct {
    Curve elliptic.Curve
    X, Y  *big.Int
}

// GenerateKey generates a new SM2 key pair.
func GenerateKey(rand io.Reader) (*PrivateKey, error)

// Sign signs digest using SM2 with SM3.
func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error)

// Verify verifies signature.
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool

// Encrypt encrypts plaintext using SM2 public key.
func Encrypt(rand io.Reader, pub *PublicKey, plaintext []byte) ([]byte, error)

// Decrypt decrypts ciphertext using SM2 private key.
func Decrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error)
```

```go
// crypto/cn/sm3/sm3.go

// New returns a new SM3 hash.Hash.
func New() hash.Hash

// Sum returns the SM3 checksum of data.
func Sum(data []byte) [32]byte
```

```go
// crypto/cn/sm4/sm4.go

// NewCipher creates an SM4 cipher.Block.
// Key must be 16 bytes.
func NewCipher(key []byte) (cipher.Block, error)

// BlockSize is 16 bytes (128 bits).
const BlockSize = 16

// KeySize is 16 bytes (128 bits).
const KeySize = 16
```

```go
// crypto/cn/sm9/sm9.go

// MasterKey represents SM9 master key (for KGC).
type MasterKey struct {
    MasterSecret *big.Int
    MasterPublic *G2Point  // For signing
}

// UserKey represents SM9 user private key.
type UserKey struct {
    ID  []byte
    Key *G1Point  // For signing
}

// GenerateMasterKey generates SM9 master key pair.
func GenerateMasterKey(rand io.Reader) (*MasterKey, error)

// GenerateUserKey derives user key from master key and ID.
func GenerateUserKey(master *MasterKey, id []byte) (*UserKey, error)

// Sign signs message with user's private key.
func Sign(rand io.Reader, key *UserKey, message []byte) (*Signature, error)

// Verify verifies signature using user ID and master public key.
func Verify(masterPub *G2Point, id []byte, message []byte, sig *Signature) bool
```

```go
// crypto/cn/tls/cipher_suites.go

const (
    // RFC 8998 TLS 1.3 cipher suites
    TLS_SM4_GCM_SM3 uint16 = 0x00C6
    TLS_SM4_CCM_SM3 uint16 = 0x00C7
)

// SM2 signature algorithm for TLS
const (
    SignatureSM2_SM3 uint16 = 0x0708
)

// SM2 named curve
const (
    CurveSM2 uint16 = 0x0029  // curveSM2 from RFC 8998
)
```

```go
// crypto/cn/provider.go

// Provider implements crypto.Provider for Chinese cryptography.
type Provider struct{}

func (p *Provider) Name() string { return "cn" }

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
    // Configure SM cipher suites, curves, etc.
}

func (p *Provider) SupportedCipherSuites() []uint16 {
    return []uint16{TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3}
}

func init() {
    crypto.Register(&Provider{})
}
```

## Data Models

### SM2 Curve Parameters (GB/T 32918.5)

```go
// SM2-P256 recommended curve
var SM2P256 = &elliptic.CurveParams{
    Name:    "SM2-P256",
    BitSize: 256,
    P:  "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF",
    N:  "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123",
    B:  "28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93",
    Gx: "32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7",
    Gy: "BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0",
}
// Note: SM2 uses a=-3 (same as P-256), so standard elliptic.Curve works
```

### SM3 Constants

```go
const (
    sm3BlockSize = 64   // 512 bits
    sm3DigestSize = 32  // 256 bits
)

// Initial values (IV)
var sm3IV = [8]uint32{
    0x7380166F, 0x4914B2B9, 0x172442D7, 0xDA8A0600,
    0xA96F30BC, 0x163138AA, 0xE38DEE4D, 0xB0FB0E4E,
}
```

### SM4 Constants

```go
const (
    sm4BlockSize = 16   // 128 bits
    sm4KeySize   = 16   // 128 bits
    sm4Rounds    = 32
)
```

### SM9 BN256 Curve

```go
// SM9 uses BN256 pairing-friendly curve (per GB/T 38635)
// G1: E(Fp), G2: E'(Fp²), GT: Fp¹²
```

## Behavior Specifications

### Happy Path

1. Server loads config with `cipherSuites: "cn"`
2. `crypto.Get("cn")` returns CN provider
3. Provider configures TLS with SM2 certificate
4. Client connects with TLS_SM4_GCM_SM3
5. Handshake uses SM2 for key exchange, SM3 for hash
6. Data encrypted with SM4-GCM

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Invalid SM4 key size | key != 16 bytes | Return error |
| Invalid SM2 signature | malformed r,s | Verify returns false |
| Empty SM3 input | hash of "" | Return valid 32-byte hash |
| SM9 unknown ID | ID not in system | GenerateUserKey creates new key |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| `ErrInvalidKeySize` | SM4 key not 16 bytes | Return error immediately |
| `ErrInvalidPublicKey` | SM2 point not on curve | Return error |
| `ErrDecryptionFailed` | SM2 decrypt with wrong key | Return error |
| `ErrInvalidSignature` | Signature verification failed | Return false |

## Dependencies

### Requires

- `crypto/provider.go` - Provider interface (exists)
- `math/big` - Big integer operations
- `crypto/elliptic` - Elliptic curve interface

### Blocks

- `sdd-https-vpn-multi-cert` - Can use CN provider once ready

## Integration Points

### External Systems

- SM2/SM3/SM4 compatible clients
- Chinese government systems requiring SM crypto

### Internal Systems

- `crypto.Registry` - Provider registration
- TLS configuration pipeline

## Testing Strategy

### Unit Tests

- [ ] `sm2/sm2_test.go` - Key generation, sign/verify, encrypt/decrypt
- [ ] `sm3/sm3_test.go` - Hash with official test vectors
- [ ] `sm4/sm4_test.go` - Encrypt/decrypt with official test vectors
- [ ] `sm9/sm9_test.go` - Sign/verify with test vectors

### Test Vectors

Использовать официальные тест-векторы из:
- GB/T 32918 Appendix A (SM2)
- GB/T 32905 Appendix A (SM3)
- GB/T 32907 Appendix A (SM4)
- GB/T 38635 Appendix A (SM9)

### Integration Tests

- [ ] Provider registration - `crypto.Get("cn")` returns valid provider
- [ ] TLS handshake with SM cipher suite

### Manual Verification

- [ ] Server starts with `cipherSuites: "cn"`
- [ ] Client with SM support connects successfully

## Migration / Rollout

1. Implement primitives (SM2, SM3, SM4) first
2. Add SM9 (more complex due to pairings)
3. Implement TLS cipher suites
4. Register provider
5. Test end-to-end

## Open Design Questions

- [x] Все решены на этапе требований

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
