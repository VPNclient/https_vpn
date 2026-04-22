# Plan: European Cryptography (Brainpool ECC)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22
> Specifications: [02-specifications.md](02-specifications.md)

## Task Breakdown

### Phase 1: Brainpool Curves Implementation

1. **Implement Brainpool Curves**
   - Create `crypto/eu/brainpool/curves.go`.
   - Implement `elliptic.Curve` for P256r1, P384r1, and P512r1 using `math/big`.
   - Use constants from RFC 5639.
   - 📂 `crypto/eu/brainpool/curves.go`

2. **Unit Tests for Curves**
   - Implement tests using official RFC 5639 test vectors.
   - Verify point addition, doubling, and scalar multiplication.
   - 📂 `crypto/eu/brainpool/brainpool_test.go`

### Phase 2: Crypto Provider

3. **EU Provider Registration**
   - Create `crypto/eu/provider.go`.
   - Implement `crypto.Provider` interface.
   - Register the provider in `init()`.
   - 📂 `crypto/eu/provider.go`

4. **TLS Integration Research**
   - Investigate how to inject custom Brainpool curves into `crypto/tls` for TLS 1.3.
   - If standard `crypto/tls` is too restrictive, document the limitation or propose a workaround (e.g., using a modified `crypto/tls` or custom handshake).

### Phase 3: Verification

5. **Integration Tests**
   - Test `crypto.Get("eu")`.
   - Verify that the provider correctly sets up `tls.Config`.

## Testing Strategy

### Primitives
- Validate curve parameters against RFC 5639.
- Use `crypto/ecdsa` to ensure `elliptic.Curve` implementation is compatible with standard signing.

### TLS
- Mock TLS handshake or use local server/client to verify curve negotiation if supported.

## Rollback Plan

- Since this is a new provider, rollback involves removing the `import _ "github.com/nativemind/https-vpn/crypto/eu"` or deleting the `crypto/eu` package.

## Complexity Estimate

- Brainpool Primitives: Medium (requires careful constant entry and verification)
- TLS Integration: High (due to Go stdlib restrictions on custom curves in TLS 1.3)
- Overall: Medium-High
