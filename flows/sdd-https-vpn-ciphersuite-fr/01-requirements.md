# Requirements: French National Cryptography (ANSSI)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22

## Problem Statement

Добавить поддержку французских криптографических рекомендаций (ANSSI) в систему HTTPS VPN для обеспечения соответствия требованиям безопасности, установленным ANSSI, и расширения географического охвата сервиса.

Система уже поддерживает:
- "us" - стандартная криптография (RSA/ECDSA)
- "ru" - российская криптография (ГОСТ)
- "cn" - китайская криптография (SM)
- "eu" - европейская криптография (Brainpool)
- "uk" - британская криптография (NCSC)

Необходимо добавить:
- "fr" - французская криптография (ANSSI-compliant parameters)

## User Stories

### Primary

**As a** VPN service operator in France
**I want** to configure French cryptography (ANSSI-approved parameters)
**So that** my service complies with ANSSI security recommendations.

**As a** user in a high-security French environment
**I want** to connect using ANSSI-recommended cryptographic algorithms
**So that** my VPN traffic uses approved national standards.

### Secondary

**As a** developer
**I want** the FR implementation to follow the same pattern as existing providers (US/RU/CN/EU/UK)
**So that** code is consistent and maintainable.

## Acceptance Criteria

### Must Have

1. **Given** a server config with `cipherSuites: "fr"`
   **When** the server starts
   **Then** TLS connections use ANSSI-recommended cryptographic primitives.

2. **Given** the crypto registry
   **When** `crypto.List()` is called
   **Then** "fr" appears in the list of available providers.

3. **Given** a TLS 1.3 connection with "fr" provider
   **When** performing a handshake
   **Then** only approved elliptic curves (e.g., Brainpool or high-security NIST/SEC curves as recommended by ANSSI) and ciphers (AES-GCM-256) are allowed.

4. **Given** an attempt to connect using legacy protocols (TLS 1.2 or below)
   **When** using "fr" provider
   **Then** the connection is rejected, in accordance with ANSSI "hardened" profile recommendations.

### Should Have

- Automated tests verifying "fr" provider compliance.
- Documentation updating the supported countries list in README.

### Won't Have (This Iteration)

- Hardware security module (HSM) integration.
- Custom French symmetric ciphers (reusing standard AES).

## Constraints

- **Technical**: Must integrate with existing `crypto.Provider` interface.
- **Standards**: Must comply with ANSSI guidance for TLS/VPN.
- **Pattern**: Must follow the existing provider pattern.

## Open Questions

- [ ] Какие специфические параметры ANSSI наиболее приоритетны для реализации?
- [ ] Требуется ли строгая изоляция FR от EU профиля?

## References

- [ANSSI: Recommandations de sécurité relatives à la mise en œuvre de TLS](https://www.ssi.gouv.fr/)
- Existing providers: `crypto/us/`, `crypto/ru/`, `crypto/cn/`

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
