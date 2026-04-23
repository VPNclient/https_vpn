# Requirements: European Cryptography (Brainpool ECC)

> Version: 1.1
> Status: APPROVED
> Last Updated: 2026-04-22

## Problem Statement

Добавить поддержку европейских криптографических стандартов (Brainpool ECC) в систему HTTPS VPN для обеспечения совместимости с требованиями ЕС (ETSI/BSI) и расширения географического охвата сервиса.

Система уже поддерживает:
- "us" - стандартная криптография (RSA/ECDSA/NIST Curves)
- "ru" - российская криптография (ГОСТ)
- "cn" - китайская криптография (SM)

Необходимо добавить:
- "eu" - европейская криптография (Brainpool ECC curves)

## User Stories

### Primary

**As a** VPN service operator in Europe
**I want** to configure European cryptography (Brainpool curves)
**So that** my service complies with EU cryptographic recommendations (BSI TR-02102)

**As a** user in the EU
**I want** to connect using Brainpool ECC curves
**So that** my VPN traffic uses approved European standards

### Secondary

**As a** developer
**I want** the EU implementation to follow the same pattern as US/RU/CN
**So that** code is consistent and maintainable

## Acceptance Criteria

### Must Have

1. **Given** a server config with `cipherSuites: "eu"`
   **When** the server starts
   **Then** TLS connections use Brainpool ECC curves for key exchange

2. **Given** the crypto registry
   **When** `crypto.List()` is called
   **Then** "eu" appears in the list of available providers

3. **Given** a Brainpool ECC key pair
   **When** signing and verifying data
   **Then** signatures are valid per RFC 5639

4. **Given** TLS connection with "eu" provider
   **When** using AES-GCM and SHA-256
   **Then** handshake and data transfer succeed using Brainpool curves

### Should Have

- Unit tests for all Brainpool curve implementations
- Support for multiple Brainpool curves: P256r1, P384r1, P512r1
- Integration with standard Go `crypto/tls` (via custom curve implementation)

### Won't Have (This Iteration)

- Hardware security module (HSM) integration
- Custom European symmetric ciphers (reusing standard AES)
- Custom European hash functions (reusing standard SHA-2)

## Constraints

- **Technical**: Must integrate with existing `crypto.Provider` interface
- **Standards**: Must comply with RFC 5639 (Brainpool Elliptic Curve Cryptography)
- **Pattern**: Must follow existing provider structure (`crypto/us/`, `crypto/ru/`, `crypto/cn/`)
- **Dependencies**: No external CGO dependencies preferred (pure Go)

## Open Questions

- **[x] Кривые Brainpool**: Поддержка всех основных кривых: P256r1, P384r1, P512r1.
- **[x] Версии TLS**: Поддержка TLS 1.2 и TLS 1.3.
- **[x] Шифры TLS**: Достаточно стандартных AES-GCM с кривыми Brainpool.

## References

- RFC 5639 - Elliptic Curve Cryptography (ECC) Brainpool Standard Curves and Curve Generation
- RFC 7027 - Elliptic Curve Cryptography (ECC) Brainpool Curves for Transport Layer Security (TLS)
- BSI TR-02102 - Cryptographic Mechanisms: Recommendations and Key Lengths
- Existing CN implementation: `crypto/cn/`

---

## Approval

- Reviewed by: User
- Approved on: 2026-04-22
- Notes: User confirmed choices for curves, TLS versions, and cipher suites.
