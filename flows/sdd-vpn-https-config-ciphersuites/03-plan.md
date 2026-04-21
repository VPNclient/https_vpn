# Implementation Plan: Crypto Provider Selection via CipherSuites

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-21
> Specifications: [02-specifications.md](02-specifications.md)

## Summary

Implement crypto provider selection by parsing the `cipherSuites` field in TLS configuration. The implementation involves modifying the config parser and core initialization logic to recognize provider identifiers ("ru", "cn", "us") within the existing cipherSuites string.

## Task Breakdown

### Phase 1: Configuration Structure

#### Task 1.1: Add Deprecated Comment to CryptoProvider
- **Description**: Mark the `cryptoProvider` field as deprecated in TLSConfig
- **Files**:
  - `infra/conf/config.go` - Modify (add deprecation comment)
- **Dependencies**: None
- **Verification**: Code review, go build succeeds
- **Complexity**: Low

#### Task 1.2: Document CipherSuites Dual Purpose
- **Description**: Add documentation comment explaining the new dual-purpose usage of cipherSuites
- **Files**:
  - `infra/conf/config.go` - Modify (update comments)
- **Dependencies**: None
- **Verification**: Code review
- **Complexity**: Low

### Phase 2: Core Provider Selection Logic

#### Task 2.1: Implement selectCryptoProvider Function
- **Description**: Create function to parse cipherSuites and select crypto provider
- **Files**:
  - `core/core.go` - Modify (add new function)
- **Dependencies**: Task 1.1, Task 1.2
- **Verification**: Unit tests pass
- **Complexity**: Medium

```go
func selectCryptoProvider(cipherSuites, cryptoProvider string) crypto.Provider {
    // 1. Parse cipherSuites
    // 2. Check against crypto.List()
    // 3. Fallback to cryptoProvider
    // 4. Default to "us"
}
```

#### Task 2.2: Integrate selectCryptoProvider into TLS Init
- **Description**: Call selectCryptoProvider during TLS initialization
- **Files**:
  - `core/core.go` - Modify (update TLS init)
- **Dependencies**: Task 2.1
- **Verification**: Server starts with correct provider
- **Complexity**: Medium

### Phase 3: Logging and Warnings

#### Task 3.1: Add Deprecation Warning for CryptoProvider
- **Description**: Log warning when deprecated cryptoProvider field is used
- **Files**:
  - `core/core.go` - Modify (add logging)
- **Dependencies**: Task 2.1
- **Verification**: Warning appears in logs when using deprecated field
- **Complexity**: Low

#### Task 3.2: Add Provider Selection Logging
- **Description**: Log which crypto provider was selected at startup
- **Files**:
  - `core/core.go` - Modify (add logging)
- **Dependencies**: Task 2.2
- **Verification**: Provider name logged at startup
- **Complexity**: Low

### Phase 4: Testing

#### Task 4.1: Unit Tests for selectCryptoProvider
- **Description**: Write unit tests covering all parsing scenarios
- **Files**:
  - `core/core_test.go` - Create/Modify
- **Dependencies**: Task 2.1
- **Verification**: All tests pass
- **Complexity**: Medium

#### Task 4.2: Integration Test with Config
- **Description**: Test end-to-end config loading with new cipherSuites usage
- **Files**:
  - Test config files or integration tests
- **Dependencies**: Task 2.2
- **Verification**: Server starts with each provider type
- **Complexity**: Medium

## Dependency Graph

```
Task 1.1 ──┬──→ Task 2.1 ──┬──→ Task 2.2 ──→ Task 4.2
Task 1.2 ──┘              │
                          ├──→ Task 3.1
                          ├──→ Task 3.2
                          └──→ Task 4.1
```

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `infra/conf/config.go` | Modify | Add deprecation comment, document dual-purpose |
| `core/core.go` | Modify | Add selectCryptoProvider, integrate into TLS init |
| `core/core_test.go` | Modify | Add unit tests for new function |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing configs | Low | High | Fallback ensures backward compatibility |
| crypto.List() returns empty | Low | Medium | Default to "us" provider |
| Standard cipher names conflict with providers | Low | Low | Provider names are short (2 letters), unlikely collision |

## Rollback Strategy

If implementation fails or needs to be reverted:

1. Remove selectCryptoProvider function
2. Restore original TLS initialization
3. Keep deprecated cryptoProvider field working
4. Git revert commits

## Checkpoints

After each phase, verify:

- [ ] All tests pass
- [ ] go build succeeds
- [ ] Existing configs still work
- [ ] New cipherSuites provider selection works

## Open Implementation Questions

- [ ] Exact location of selectCryptoProvider in core/core.go
- [ ] Existing crypto.List() function signature and return type
- [ ] Current logging mechanism/library used

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
