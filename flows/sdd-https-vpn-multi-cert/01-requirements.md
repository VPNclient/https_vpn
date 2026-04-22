# Requirements: Multi-Provider Certificate Selection

> Version: 1.0
> Status: REVIEW
> Last Updated: 2026-04-22

## Problem Statement

Сервер должен автоматически выбирать подходящий сертификат на основе возможностей клиента. Это позволит одному серверу обслуживать клиентов с разными криптографическими требованиями:

- Китайские клиенты → SM2 сертификат
- Российские клиенты → GOST сертификат
- Стандартные клиенты → RSA/ECDSA сертификат

## User Stories

### Primary

**As a** VPN service operator
**I want** to configure multiple certificates for different crypto providers
**So that** one server can serve clients with different cryptographic requirements

**As a** client
**I want** the server to automatically select a compatible certificate
**So that** I can connect using my preferred/required cryptography

## Acceptance Criteria

### Must Have

1. **Given** server config with multiple certificates (SM2, GOST, RSA)
   **When** SM-capable client connects
   **Then** server uses SM2 certificate

2. **Given** server config with multiple certificates
   **When** GOST-capable client connects
   **Then** server uses GOST certificate

3. **Given** server config with multiple certificates
   **When** standard client connects (no SM/GOST)
   **Then** server uses RSA/ECDSA certificate

4. **Given** `cipherSuites: "cn,ru,us"`
   **When** client supports both SM and GOST
   **Then** server prefers SM (first in list)

5. **Given** `cipherSuites: "ru"`
   **When** standard client connects (no GOST)
   **Then** connection fails (no fallback configured)

### Should Have

- Logging of certificate selection decisions
- Metrics for certificate usage by provider type

### Won't Have (This Iteration)

- Runtime certificate reload without restart
- Per-SNI certificate selection (beyond crypto provider)

## Constraints

- **Technical**: Must work with Go's `crypto/tls` GetCertificate callback
- **Compatibility**: Must not break existing single-certificate configs
- **Dependencies**: Requires crypto providers to be registered (us, ru, cn)

## Configuration

```json
{
  "tlsSettings": {
    "certificates": [
      { "certificateFile": "sm2.crt", "keyFile": "sm2.key" },
      { "certificateFile": "gost.crt", "keyFile": "gost.key" },
      { "certificateFile": "rsa.crt", "keyFile": "rsa.key" }
    ],
    "cipherSuites": "cn,ru,us"
  }
}
```

## Open Questions

- [x] Все вопросы решены

## References

- Go TLS GetCertificate: https://pkg.go.dev/crypto/tls#Config
- Related flow: `sdd-https-vpn-ciphersuite-cn`
- Provider selection: `sdd-vpn-https-config-ciphersuites`

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
