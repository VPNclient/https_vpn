# Requirements: UK NCSC Cryptography Compliance

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22

## Problem Statement

To provide a secure and compliant VPN solution for UK-based users and organizations, the system must support a cryptographic profile that follows the UK National Cyber Security Centre (NCSC) recommendations. While standard "US" (NIST) cryptography is similar, a dedicated "uk" provider allows for a more restrictive and optimized configuration that prioritizes NCSC-preferred algorithms (e.g., P-384 and AES-256-GCM-SHA384).

## User Stories

### Primary

**As a** UK-based VPN administrator
**I want** to configure the VPN to use NCSC-compliant cryptography by setting `cipherSuites: "uk"`
**So that** my infrastructure complies with national security guidelines.

**As a** user in a high-assurance environment
**I want** to use the strongest NCSC-recommended algorithms (AES-256, P-384)
**So that** my communications are protected according to the highest standards.

### Secondary

**As a** developer
**I want** a dedicated UK provider
**So that** I can easily update UK-specific cryptographic policies independently of the global/US defaults.

## Acceptance Criteria

### Must Have

1. **Given** a server config with `cipherSuites: "uk"`
   **When** the server starts
   **Then** TLS connections use NCSC-compliant cryptography (TLS 1.3 only).

2. **Given** the crypto registry
   **When** `crypto.List()` is called
   **Then** "uk" appears in the list of available providers.

3. **Given** a TLS 1.3 handshake
   **When** the "uk" provider is active
   **Then** `TLS_AES_256_GCM_SHA384` is preferred over `TLS_AES_128_GCM_SHA256`.

4. **Given** a TLS 1.3 handshake
   **When** the "uk" provider is active
   **Then** `CurveP384` is preferred for key exchange.

5. **Given** a client attempting to use TLS 1.2 or lower
   **When** the "uk" provider is active
   **Then** the connection is rejected.

### Should Have

- Automated tests verifying that "uk" provider correctly filters and orders cipher suites.
- Documentation in README updating the supported countries table status for UK.

### Won't Have (This Iteration)

- Support for "Foundation Grade" vs "High Assurance" toggle (will default to High Assurance recommendations).
- Custom UK-specific CA or PKI tools.

## Constraints

- **Technical**: Must integrate with existing `crypto.Provider` interface.
- **Standards**: Must comply with NCSC "Using TLS to protect data" guidance.
- **Pattern**: Must follow the existing provider pattern (`crypto/us/`, `crypto/ru/`).

## Open Questions

- [x] Should we support ChaCha20-Poly1305? → **NCSC recommends AES-GCM; ChaCha20 is acceptable but AES is preferred if hardware acceleration is available. We will omit it for "uk" to be more restrictive.**
- [x] Is RSA supported? → **NCSC prefers ECDSA for new deployments. We will support ECDSA with P-384/P-256.**

## References

- [NCSC: Using TLS to protect data](https://www.ncsc.gov.uk/collection/tls-pfs-guidance)
- [NCSC: Cryptographic mechanisms for information protection](https://www.ncsc.gov.uk/guidance/cryptographic-mechanisms-for-information-protection)

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
