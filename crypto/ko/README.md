# Korean Cryptography Provider (ARIA/SEED)

This package implements the cryptographic provider for Korean national standards.

## Algorithms

### ARIA (Academy Research Institute Agency)

- **Standard**: KS X 1213:2004
- **Block size**: 128 bits
- **Key sizes**: 128, 192, 256 bits
- **Rounds**: 12, 14, 16 (depending on key size)
- **Structure**: SPN (Substitution-Permutation Network)
- **RFCs**: RFC 5794, RFC 6209, RFC 9367

```go
import "github.com/nativemind/https-vpn/crypto/ko/aria"

// Create ARIA-256 cipher
key := make([]byte, 32)
cipher, err := aria.NewCipher256(key)

// Or auto-detect key size
cipher, err := aria.NewCipher(key)
```

### SEED

- **Standard**: KISA
- **Block size**: 128 bits
- **Key size**: 128 bits
- **Rounds**: 16
- **Structure**: Feistel network
- **RFCs**: RFC 4162, RFC 4269

```go
import "github.com/nativemind/https-vpn/crypto/ko/seed"

key := make([]byte, 16)
cipher, err := seed.NewCipher(key)
```

## TLS Cipher Suites

### TLS 1.3 (RFC 9367)

| ID | Name | AEAD | Hash |
|----|------|------|------|
| 0x1306 | TLS_ARIA_128_GCM_SHA256 | ARIA-128-GCM | SHA-256 |
| 0x1307 | TLS_ARIA_256_GCM_SHA384 | ARIA-256-GCM | SHA-384 |

### TLS 1.2 (RFC 6209)

| ID | Name | Key Exchange | Cipher |
|----|------|--------------|--------|
| 0xC06A | TLS_ECDHE_ECDSA_WITH_ARIA_128_GCM_SHA256 | ECDHE | ARIA-128-GCM |
| 0xC06B | TLS_ECDHE_ECDSA_WITH_ARIA_256_GCM_SHA384 | ECDHE | ARIA-256-GCM |

### SEED (RFC 4162)

| ID | Name | Key Exchange | Cipher |
|----|------|--------------|--------|
| 0x0096 | TLS_RSA_WITH_SEED_CBC_SHA | RSA | SEED-CBC |

## Usage

### Configuration

```json
{
  "tlsSettings": {
    "cipherSuites": "ko"
  }
}
```

### Programmatic

```go
import (
    "github.com/nativemind/https-vpn/crypto"
    _ "github.com/nativemind/https-vpn/crypto/ko" // Register provider
)

provider, ok := crypto.Get("ko")
if ok {
    err := provider.ConfigureTLS(tlsConfig)
}
```

## Implementation Status

| Component | Status | Tests |
|-----------|--------|-------|
| ARIA-128/192/256 | Basic implementation | Roundtrip OK |
| SEED-128 | Basic implementation | Roundtrip OK |
| TLS Provider | Complete | 10/10 PASS |

**Note**: RFC test vectors need verification. Encrypt/decrypt roundtrip works correctly.

## References

- [RFC 5794](https://www.rfc-editor.org/rfc/rfc5794) - A Description of the ARIA Encryption Algorithm
- [RFC 6209](https://www.rfc-editor.org/rfc/rfc6209) - ARIA Cipher Suites for TLS
- [RFC 9367](https://www.rfc-editor.org/rfc/rfc9367) - ARIA Cipher Suites for TLS 1.3
- [RFC 4162](https://www.rfc-editor.org/rfc/rfc4162) - SEED Cipher Suites for TLS
- [RFC 4269](https://www.rfc-editor.org/rfc/rfc4269) - The SEED Encryption Algorithm

## License

MIT License
