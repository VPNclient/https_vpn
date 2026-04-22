# Implementation Plan: https-cisco-firmware-compatibility

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-21
> Specifications: [02-specifications.md](02-specifications.md)

## Summary

Implement GOST crypto provider in 4 phases: primitives, TLS layer, provider integration, and testing. Each phase builds on the previous and can be independently verified.

## Task Breakdown

### Phase 1: GOST Primitives

Foundation cryptographic implementations with test vectors from GOST standards.

#### Task 1.1: Kuznyechik (Grasshopper) Block Cipher
- **Description**: Implement GOST R 34.12-2015 Kuznyechik block cipher (128-bit block, 256-bit key)
- **Files**:
  - `crypto/ru/gost/kuznyechik.go` - Create
  - `crypto/ru/gost/kuznyechik_test.go` - Create
- **Dependencies**: None
- **Verification**: Test vectors from GOST R 34.12-2015 Appendix A
- **Complexity**: Medium

#### Task 1.2: Magma Block Cipher
- **Description**: Implement GOST R 34.12-2015 Magma block cipher (64-bit block, 256-bit key)
- **Files**:
  - `crypto/ru/gost/magma.go` - Create
  - `crypto/ru/gost/magma_test.go` - Create
- **Dependencies**: None
- **Verification**: Test vectors from GOST R 34.12-2015 Appendix B
- **Complexity**: Medium

#### Task 1.3: CTR and MGM Modes
- **Description**: Implement CTR mode and MGM (Multilinear Galois Mode) AEAD for Kuznyechik/Magma
- **Files**:
  - `crypto/ru/gost/ctr.go` - Create
  - `crypto/ru/gost/mgm.go` - Create
  - `crypto/ru/gost/mgm_test.go` - Create
- **Dependencies**: Task 1.1, Task 1.2
- **Verification**: Test vectors from GOST R 34.13-2015
- **Complexity**: High

#### Task 1.4: Streebog Hash Function
- **Description**: Implement GOST R 34.11-2012 Streebog hash (256-bit and 512-bit)
- **Files**:
  - `crypto/ru/gost/streebog.go` - Create
  - `crypto/ru/gost/streebog_test.go` - Create
- **Dependencies**: None
- **Verification**: Test vectors from GOST R 34.11-2012 Appendix
- **Complexity**: Medium

#### Task 1.5: GOST Elliptic Curves
- **Description**: Implement GOST R 34.10-2012 elliptic curves (256-bit and 512-bit parameter sets)
- **Files**:
  - `crypto/ru/gost/curves.go` - Create
  - `crypto/ru/gost/curves_test.go` - Create
- **Dependencies**: None
- **Verification**: Point multiplication test vectors
- **Complexity**: High

#### Task 1.6: GOST R 34.10-2012 Signatures
- **Description**: Implement GOST digital signature algorithm
- **Files**:
  - `crypto/ru/gost/gost3410.go` - Create
  - `crypto/ru/gost/gost3410_test.go` - Create
- **Dependencies**: Task 1.4, Task 1.5
- **Verification**: Signature test vectors from standard
- **Complexity**: High

#### Task 1.7: HMAC-Streebog and OMAC
- **Description**: Implement HMAC with Streebog and OMAC (CMAC variant) for TLS PRF
- **Files**:
  - `crypto/ru/gost/hmac.go` - Create
  - `crypto/ru/gost/omac.go` - Create
  - `crypto/ru/gost/mac_test.go` - Create
- **Dependencies**: Task 1.1, Task 1.4
- **Verification**: Test vectors
- **Complexity**: Medium

---

### Phase 2: GOST TLS Layer

Custom TLS 1.3 implementation with GOST cipher suites per RFC 9189.

#### Task 2.1: TLS Record Layer
- **Description**: Implement TLS 1.3 record protocol with GOST AEAD
- **Files**:
  - `crypto/ru/tls/record.go` - Create
  - `crypto/ru/tls/record_test.go` - Create
- **Dependencies**: Phase 1 complete
- **Verification**: Encrypt/decrypt test vectors
- **Complexity**: High

#### Task 2.2: GOST Key Exchange (VKO)
- **Description**: Implement VKO GOST R 34.10-2012 key agreement
- **Files**:
  - `crypto/ru/tls/vko.go` - Create
  - `crypto/ru/tls/vko_test.go` - Create
- **Dependencies**: Task 1.5, Task 1.6
- **Verification**: Key exchange test vectors
- **Complexity**: High

#### Task 2.3: TLS Handshake State Machine
- **Description**: Implement TLS 1.3 handshake with GOST extensions
- **Files**:
  - `crypto/ru/tls/handshake.go` - Create
  - `crypto/ru/tls/handshake_client.go` - Create (for testing)
  - `crypto/ru/tls/handshake_server.go` - Create
  - `crypto/ru/tls/handshake_test.go` - Create
- **Dependencies**: Task 2.1, Task 2.2
- **Verification**: Full handshake with test client
- **Complexity**: High

#### Task 2.4: GOST Certificate Parsing
- **Description**: Parse X.509 certificates with GOST keys and signatures
- **Files**:
  - `crypto/ru/tls/cert.go` - Create
  - `crypto/ru/tls/cert_test.go` - Create
- **Dependencies**: Task 1.6
- **Verification**: Parse real GOST certificates
- **Complexity**: Medium

#### Task 2.5: TLS Connection Wrapper
- **Description**: Implement net.Conn wrapper for GOST TLS
- **Files**:
  - `crypto/ru/tls/conn.go` - Create
  - `crypto/ru/tls/listener.go` - Create
- **Dependencies**: Task 2.3, Task 2.4
- **Verification**: HTTP/2 over GOST TLS connection
- **Complexity**: Medium

#### Task 2.6: Cipher Suite Definitions
- **Description**: Define GOST cipher suite constants and negotiation
- **Files**:
  - `crypto/ru/tls/cipher_suites.go` - Create
- **Dependencies**: None
- **Verification**: Constants match RFC 9189
- **Complexity**: Low

---

### Phase 3: Provider Integration

Integrate GOST TLS with existing provider system.

#### Task 3.1: Extend Provider Interface
- **Description**: Add TLSListener method to crypto.Provider interface
- **Files**:
  - `crypto/provider.go` - Modify
- **Dependencies**: None
- **Verification**: Existing US provider still works
- **Complexity**: Low

#### Task 3.2: Update US Provider
- **Description**: Add nil TLSListener() to US provider (uses stdlib)
- **Files**:
  - `crypto/us/provider.go` - Modify
- **Dependencies**: Task 3.1
- **Verification**: US provider compiles and works
- **Complexity**: Low

#### Task 3.3: Implement RU Provider
- **Description**: Create GOST provider implementing extended interface
- **Files**:
  - `crypto/ru/provider.go` - Create
- **Dependencies**: Phase 2 complete, Task 3.1
- **Verification**: Provider registers and returns custom TLS listener
- **Complexity**: Medium

#### Task 3.4: Update Transport Server
- **Description**: Support custom TLS listener from provider
- **Files**:
  - `transport/server.go` - Modify
- **Dependencies**: Task 3.1
- **Verification**: Server uses provider's TLSListener when available
- **Complexity**: Medium

#### Task 3.5: Configuration Support
- **Description**: Ensure config correctly loads GOST certificates
- **Files**:
  - `infra/conf/config.go` - Modify (if needed)
- **Dependencies**: Task 2.4
- **Verification**: GOST cert loads from config file
- **Complexity**: Low

---

### Phase 4: Testing & Documentation

End-to-end testing and verification.

#### Task 4.1: Generate Test Certificates
- **Description**: Create self-signed GOST certificates for testing
- **Files**:
  - `crypto/ru/testdata/gost_cert.pem` - Create
  - `crypto/ru/testdata/gost_key.pem` - Create
  - `crypto/ru/tools/gencert.go` - Create (certificate generator)
- **Dependencies**: Task 1.6, Task 2.4
- **Verification**: Certificates parse correctly
- **Complexity**: Medium

#### Task 4.2: Integration Tests
- **Description**: Full VPN flow over GOST TLS
- **Files**:
  - `crypto/ru/integration_test.go` - Create
- **Dependencies**: Phase 3 complete, Task 4.1
- **Verification**: HTTP/2 CONNECT works over GOST TLS
- **Complexity**: Medium

#### Task 4.3: Interoperability Testing
- **Description**: Test with external GOST TLS clients
- **Files**:
  - `docs/gost-testing.md` - Create
- **Dependencies**: Task 4.2
- **Verification**: OpenConnect or similar client connects
- **Complexity**: High (requires external tools)

---

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `crypto/provider.go` | Modify | Add TLSListener method |
| `crypto/us/provider.go` | Modify | Implement TLSListener (nil) |
| `crypto/ru/provider.go` | Create | GOST provider |
| `crypto/ru/gost/kuznyechik.go` | Create | Grasshopper cipher |
| `crypto/ru/gost/magma.go` | Create | Magma cipher |
| `crypto/ru/gost/ctr.go` | Create | CTR mode |
| `crypto/ru/gost/mgm.go` | Create | MGM AEAD mode |
| `crypto/ru/gost/streebog.go` | Create | Streebog hash |
| `crypto/ru/gost/curves.go` | Create | GOST elliptic curves |
| `crypto/ru/gost/gost3410.go` | Create | GOST signatures |
| `crypto/ru/gost/hmac.go` | Create | HMAC-Streebog |
| `crypto/ru/gost/omac.go` | Create | OMAC/CMAC |
| `crypto/ru/tls/record.go` | Create | TLS record layer |
| `crypto/ru/tls/vko.go` | Create | Key exchange |
| `crypto/ru/tls/handshake.go` | Create | TLS handshake |
| `crypto/ru/tls/handshake_server.go` | Create | Server handshake |
| `crypto/ru/tls/cert.go` | Create | GOST certificates |
| `crypto/ru/tls/conn.go` | Create | TLS connection |
| `crypto/ru/tls/listener.go` | Create | TLS listener |
| `crypto/ru/tls/cipher_suites.go` | Create | Cipher suite IDs |
| `transport/server.go` | Modify | Custom TLS support |
| `crypto/ru/testdata/*` | Create | Test certificates |
| `crypto/ru/*_test.go` | Create | Unit tests |

**Total: ~25 new files, 3 modified files**

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| GOST implementation bugs | Medium | High | Extensive test vectors from standards |
| TLS handshake incompatibility | Medium | High | Test with multiple clients |
| Performance issues | Low | Medium | Benchmark and optimize hot paths |
| Certificate parsing edge cases | Medium | Medium | Test with real-world GOST certs |
| Memory leaks in crypto | Low | High | Careful buffer management, fuzzing |

## Rollback Strategy

1. GOST provider is additive - no changes to existing functionality
2. If issues found, remove `crypto/ru/` directory
3. Revert `crypto/provider.go` and `transport/server.go` changes
4. US provider continues to work unchanged

## Checkpoints

### After Phase 1
- [ ] All GOST primitives pass standard test vectors
- [ ] No compiler warnings
- [ ] Benchmarks show acceptable performance

### After Phase 2
- [ ] TLS handshake completes between test client/server
- [ ] HTTP/2 negotiation works
- [ ] Encrypted data transfers correctly

### After Phase 3
- [ ] `cipherSuites: "ru"` selects GOST provider
- [ ] Server starts with GOST certificates
- [ ] US provider unaffected

### After Phase 4
- [ ] Integration tests pass
- [ ] External client connects successfully
- [ ] Documentation complete

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-21
- [x] Notes: Approved - proceed with 4-phase implementation
