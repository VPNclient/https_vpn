# Requirements: Korean Ciphersuite (ARIA/SEED)

> Version: 1.0  
> Status: DRAFT  
> Last Updated: 2026-04-28

## Problem Statement

Users in South Korea often require or prefer to use national cryptographic standards (ARIA and SEED) for compliance with local regulations and security policies. Standard TLS implementations (like Go's `crypto/tls`) often lack built-in support for these ciphersuites.

## User Stories

### Primary

**As a** VPN user in South Korea  
**I want** to use ARIA and SEED encryption algorithms  
**So that** my connection complies with national security standards (KS X 1213, KS X 1213-1).

### Secondary

**As a** system administrator  
**I want** to configure the VPN to support Korean national ciphersuites  
**So that** I can offer a localized and compliant service.

## Acceptance Criteria

### Must Have

1. **Given** a VPN client and server both supporting Korean ciphersuites  
   **When** a connection is established using TLS 1.3 with ARIA-GCM  
   **Then** the handshake should succeed and data should be encrypted using ARIA.

2. **Given** a legacy environment  
   **When** a connection is established using TLS 1.2 with ARIA-CBC or SEED-CBC  
   **Then** the connection should be secure and use the specified algorithm.

3. **Given** the `crypto` package  
   **When** I register the `ko` provider  
   **Then** it should expose the ARIA and SEED ciphersuites.

### Should Have

- High-performance constant-time implementations of ARIA and SEED.
- Comprehensive test suite covering RFC test vectors.

### Won't Have (This Iteration)

- SM series (Chinese) or other national ciphersuites not related to Korea (already covered or out of scope).
- HAS-160 support if it's not strictly required for the TLS handshake (prefer SHA-2/SHA-3).

## Constraints

- **Technical**: Must integrate with the existing `crypto.Provider` architecture.
- **Dependencies**: May require external libraries for ARIA/SEED if they are not implemented from scratch.

## Open Questions

- [ ] Does the current `xray-core` fork support custom cipher suites for TLS 1.3?
- [ ] Should we implement ARIA/SEED from scratch or use an existing Go library?

## References

- RFC 9367 (ARIA in TLS 1.3)
- RFC 6209 (ARIA in TLS 1.2)
- RFC 4162 (SEED in TLS)

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
