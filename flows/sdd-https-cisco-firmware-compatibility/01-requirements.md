# Requirements: https-cisco-firmware-compatibility

> Version: 1.2
> Status: DRAFT
> Last Updated: 2026-04-21

## Problem Statement

The HTTPS VPN needs to support Russian national cryptography (GOST) to be compliant with regional regulations and compatible with existing infrastructure, specifically Cisco devices and firmware that employ GOST for secure communications. The goal is to allow the HTTPS VPN server to use GOST certificates for TLS handshakes over HTTP/2 while maintaining its "browser-identical" traffic profile.

## User Stories

### Primary

**As a** Network Administrator in a GOST-regulated environment
**I want** the HTTPS VPN server to support full GOST cryptographic suite over HTTP/2
**So that** I can use compliant cryptography without changing my existing client infrastructure (including Cisco devices with GOST-enabled firmware).

### Secondary

**As a** Security Auditor
**I want** the GOST implementation to be isolated and easily auditable
**So that** it can be certified according to Russian national standards.

## Acceptance Criteria

### Must Have

1. **GOST Cipher Support**: The server must support the following GOST algorithms:
   - **Кузнечик (Grasshopper)** - GOST R 34.12-2015 block cipher (128-bit block, 256-bit key)
   - **Магма (Magma)** - GOST R 34.12-2015 block cipher (64-bit block, 256-bit key)

2. **GOST Hash Support**:
   - **Стрибог (Streebog)** - GOST R 34.11-2012 hash function (256-bit and 512-bit variants)

3. **GOST Signature Support**:
   - **GOST R 34.10-2012 256-bit** - Digital signature with 256-bit key
   - **GOST R 34.10-2012 512-bit** - Digital signature with 512-bit key

4. **Russian Certificate Support**: The server must load and use certificates with GOST keys (issued by Russian CAs or self-signed for testing).

5. **HTTP/2 with GOST TLS**: Full HTTP/2 support over TLS with GOST cipher suites for VPN tunneling.

6. **Cisco Firmware Compatibility**: The TLS handshake and HTTP/2 CONNECT flow must be compatible with Cisco devices running GOST-enabled firmware.

7. **Pluggable Architecture**: GOST support must be implemented as a separate crypto provider (`crypto/ru`), following the existing `Provider` interface.

8. **No Core Changes**: The core VPN logic should remain unchanged; only the crypto provider should be added.

### Should Have

1. **Auto-detection**: The server should be able to negotiate either standard AES or GOST depending on the client's capabilities (if allowed by config).
2. **Performance**: GOST encryption/decryption should not significantly degrade performance compared to standard AES.
3. **TLS 1.3 GOST**: Support GOST cipher suites within TLS 1.3 (as per Russian standards extensions).

### Won't Have (This Iteration)

1. **Non-GOST Cisco Protocols**: Support for legacy non-HTTPS Cisco protocols (like IKEv2/IPsec) is out of scope.
2. **DTLS with GOST**: Focus is on TLS (TCP) only for now.
3. **GOST 28147-89**: Legacy GOST cipher replaced by Kuznyechik/Magma is not required.

## GOST Cipher Suite Requirements

The following TLS cipher suites must be supported (per RFC 9189 - GOST Cipher Suites for TLS 1.3):

| Cipher Suite | Key Exchange | Cipher | MAC | OID |
|--------------|--------------|--------|-----|-----|
| TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC | GOST R 34.10-2012 (256) | Kuznyechik-CTR | OMAC | - |
| TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC | GOST R 34.10-2012 (256) | Magma-CTR | OMAC | - |
| TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L | GOST R 34.10-2012 (256) | Kuznyechik-MGM | - | - |
| TLS_GOSTR341112_256_WITH_MAGMA_MGM_L | GOST R 34.10-2012 (256) | Magma-MGM | - | - |

For GOST R 34.10-2012 512-bit keys, corresponding cipher suites with `GOSTR341112_512` prefix.

## Constraints

- **Technology**: Must use Go-compatible GOST libraries:
  - Primary: `github.com/bi-zone/gost` or `github.com/ftomza/gogost`
  - Alternative: Custom implementation if needed for TLS integration
- **TLS Stack**: May require `utls` or custom TLS implementation for GOST cipher suite support (Go standard `crypto/tls` does not support GOST).
- **Standard Compliance**: Must adhere to:
  - RFC 4491 (GOST in PKIX)
  - RFC 9189 (GOST Cipher Suites for TLS 1.3)
  - GOST R 34.10-2012, GOST R 34.11-2012, GOST R 34.12-2015
- **Portability**: Should work on standard Linux/Windows environments where Cisco devices are typically deployed.

## Open Questions

- [ ] Does the Go `crypto/tls` package need to be patched or replaced to support GOST cipher suites? (Likely YES - need custom TLS implementation)
- [ ] What specific GOST cipher suites are required for compatibility with common Cisco firmware versions?
- [ ] Can we use `utls` (used in `xray-core`) to help with specialized TLS handshakes, or do we need full GOST TLS implementation?
- [ ] Which Go GOST library is most complete and actively maintained?

## References

- [RFC 4491 - Using the GOST Algorithms with Public Key Infrastructure (PKI)](https://tools.ietf.org/html/rfc4491)
- [RFC 9189 - GOST Cipher Suites for TLS 1.3](https://www.rfc-editor.org/rfc/rfc9189.html)
- [Cisco Documentation on GOST Support](https://www.cisco.com/c/en/us/td/docs/ios-xml/ios/sec_conn_pki/configuration/15-mt/sec-pki-15-mt-book/sec-pki-gost.html)
- [GOST R 34.10-2012 Standard](https://tc26.ru/standard/gost/GOST_R_3410-2012.pdf)
- [GOST R 34.11-2012 Standard](https://tc26.ru/standard/gost/GOST_R_3411-2012.pdf)
- [GOST R 34.12-2015 Standard](https://tc26.ru/standard/gost/GOST_R_34.12-2015.pdf)

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-21
- [x] Notes: Approved with full GOST suite (Kuznyechik, Magma, Streebog, GOST R 34.10-2012)
