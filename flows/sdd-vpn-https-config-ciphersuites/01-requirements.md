# Requirements: Crypto Provider Selection via CipherSuites

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-04-21

## Problem Statement

To maintain compatibility with standard Xray/V2Ray clients and GUI tools, we need a mechanism to select national cryptography providers without adding custom JSON fields. Using a custom field like `cryptoProvider` would break compatibility with existing client applications that have strict schema validation.

The solution is to reuse the standard `tlsSettings.cipherSuites` field to pass national cryptography provider identifiers, since this field already exists in the standard configuration schema.

## User Stories

### Primary

**As a** VPN server administrator
**I want** to configure national crypto providers using standard config fields
**So that** I can use existing GUI tools and clients without modification

### Secondary

**As a** user in a region with national crypto requirements
**I want** to easily switch between GOST, SM, or standard crypto
**So that** I can comply with local regulations while using familiar tools

## Acceptance Criteria

### Must Have

1. **Given** a config with `cipherSuites: "ru"`
   **When** the server starts
   **Then** GOST cryptography is used for TLS

2. **Given** a config with `cipherSuites: "cn"`
   **When** the server starts
   **Then** SM2/SM3/SM4 cryptography is used for TLS

3. **Given** a config with `cipherSuites: "us"` or empty
   **When** the server starts
   **Then** standard RSA/ECDSA cryptography is used (default)

4. **Given** a config with comma-separated values like `"ru,TLS_AES_256_GCM_SHA384"`
   **When** the server parses the config
   **Then** only the first valid provider identifier is used ("ru")

5. **Given** an empty `cipherSuites` but populated `cryptoProvider` field
   **When** the server parses the config
   **Then** the deprecated `cryptoProvider` field is used as fallback

### Should Have

- Clear error messages when invalid provider identifiers are specified
- Logging of which crypto provider was selected at startup

### Won't Have (This Iteration)

- Dynamic crypto provider switching at runtime
- Support for multiple simultaneous crypto providers
- Custom cipher suite selection within a provider

## Constraints

- **Compatibility**: Must work with unmodified Xray/V2Ray GUI clients
- **Schema**: Cannot add new fields to the standard TLS config schema
- **Backward Compatibility**: Must support the deprecated `cryptoProvider` field as fallback
- **Dependencies**: Requires crypto provider implementations to be registered via `crypto.List()`

## Open Questions

- [x] How are crypto providers registered? Via `crypto.List()` function
- [ ] What happens if the specified provider is not compiled in?
- [ ] Should we log a warning when using the deprecated `cryptoProvider` field?

## References

- Existing document: `flows/sdd-vpn-https-config-ciphersuites.md`
- Configuration structure: `infra/conf/config.go`
- Selection logic: `core/core.go`

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
