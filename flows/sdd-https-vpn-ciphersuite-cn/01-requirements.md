# Requirements: Chinese National Cryptography (SM Series)

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-04-22

## Problem Statement

Добавить поддержку китайской национальной криптографии (серия SM) в систему HTTPS VPN для обеспечения совместимости с китайскими требованиями к шифрованию и расширения географического охвата сервиса.

Система уже поддерживает:
- "us" - стандартная криптография (RSA/ECDSA)
- "ru" - российская криптография (ГОСТ)

Необходимо добавить:
- "cn" - китайская криптография (SM2/SM3/SM4/SM9)

## User Stories

### Primary

**As a** VPN service operator
**I want** to configure Chinese national cryptography (SM series)
**So that** my service complies with Chinese cryptographic requirements

**As a** user in China
**I want** to connect using SM2/SM3/SM4 cryptography
**So that** my VPN traffic uses approved national standards

### Secondary

**As a** developer
**I want** the SM implementation to follow the same pattern as GOST
**So that** code is consistent and maintainable

## Acceptance Criteria

### Must Have

1. **Given** a server config with `cipherSuites: "cn"`
   **When** the server starts
   **Then** TLS connections use SM2/SM3/SM4 cryptography

2. **Given** the crypto registry
   **When** `crypto.List()` is called
   **Then** "cn" appears in the list of available providers

3. **Given** an SM2 key pair
   **When** signing and verifying data
   **Then** signatures are valid per GB/T 32918.2

4. **Given** data to hash
   **When** using SM3 algorithm
   **Then** output matches SM3 specification (GB/T 32905)

5. **Given** data to encrypt
   **When** using SM4 cipher
   **Then** encryption/decryption works per GB/T 32907

6. **Given** SM9 identity and master keys
   **When** signing/verifying or encrypting/decrypting
   **Then** operations work per GB/T 38635

7. **Given** TLS connection with SM cipher suite
   **When** using TLS_SM4_GCM_SM3 or TLS_SM4_CCM_SM3
   **Then** handshake and data transfer succeed per RFC 8998

### Should Have

- Unit tests for all cryptographic primitives
- Test vectors from official Chinese standards
- TLS cipher suite definitions for SM
- SM9 identity-based encryption support

### Won't Have (This Iteration)

- Hardware security module (HSM) integration
- Certificate generation tooling (use external tools)
- Backward compatibility with non-SM clients on SM-configured servers
- SM9 key distribution infrastructure (KGC - Key Generation Center)
- Multi-provider certificate selection (see: `sdd-https-vpn-multi-cert`)

## Constraints

- **Technical**: Must integrate with existing `crypto.Provider` interface
- **Standards**: Must comply with Chinese national standards (GB/T series)
- **Pattern**: Must follow existing GOST implementation structure (`crypto/ru/gost/`)
- **Dependencies**: No external CGO dependencies preferred (pure Go)

## Open Questions

- [x] Какие конкретные эллиптические кривые SM2 нужны? → **Только стандартная SM2-P256**
- [x] Нужна ли поддержка SM9 (identity-based encryption)? → **Да, нужна**
- [x] Какие TLS cipher suites нужны? → **Все (TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3)**
- [x] Нужна ли поддержка двойного сертификата (SM2 + RSA)? → **Отдельный flow: `sdd-https-vpn-multi-cert`**

## References

- GB/T 32918 - SM2 Elliptic Curve Cryptography
- GB/T 32905 - SM3 Cryptographic Hash Algorithm
- GB/T 32907 - SM4 Block Cipher Algorithm
- GB/T 38635 - SM9 Identity-Based Cryptography
- RFC 8998 - TLS 1.3 with SM Cipher Suites
- Existing GOST implementation: `crypto/ru/gost/`

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
