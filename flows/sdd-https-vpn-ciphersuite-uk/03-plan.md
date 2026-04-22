# Implementation Plan: UK NCSC Cryptography Compliance

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-22
> Requirements: [01-requirements.md](01-requirements.md)
> Specifications: [02-specifications.md](02-specifications.md)

## Phase 1: Infrastructure

1. [ ] Create `crypto/uk/` directory.
2. [ ] Implement `crypto/uk/provider.go`.
3. [ ] Register the provider in `init()`.

## Phase 2: Verification

1. [ ] Add unit tests for UK provider in `crypto/uk/provider_test.go`.
2. [ ] Test registry integration.
3. [ ] Verify TLS config values match NCSC recommendations.

## Phase 3: Documentation

1. [ ] Update `README.md` "Supported Cryptography Standards" table.
2. [ ] Update `CLAUDE.md` if necessary (any new build tags).

## Phase 4: Integration

1. [ ] Verify server starts with `cipherSuites: "uk"` in `config.json`.
2. [ ] Perform a test connection using `openssl s_client` or similar to verify negotiated parameters.
