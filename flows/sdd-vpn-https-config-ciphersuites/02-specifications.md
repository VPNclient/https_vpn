# Specifications: Crypto Provider Selection via CipherSuites

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-21
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

This specification defines how the `tlsSettings.cipherSuites` string field is repurposed to select national cryptography providers while maintaining full backward compatibility with standard Xray/V2Ray clients and GUIs.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `infra/conf/config.go` | Modify | TLSConfig struct, CipherSuites field usage |
| `core/core.go` | Modify | Add provider selection logic |
| Crypto registry | Read | Query available providers via `crypto.List()` |

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     JSON Config                              │
│  { "tlsSettings": { "cipherSuites": "ru" } }                │
└─────────────────────────────────┬───────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│               infra/conf/config.go                          │
│  TLSConfig.CipherSuites → parsed string                     │
└─────────────────────────────────┬───────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                   core/core.go                               │
│  1. Parse CipherSuites (comma-split)                        │
│  2. Check each part against crypto.List()                   │
│  3. Initialize matched crypto.Provider                      │
│  4. Fallback to cryptoProvider field if no match            │
│  5. Default to "us" if all else fails                       │
└─────────────────────────────────┬───────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│               crypto.Provider                                │
│  "ru" → GOST   |   "cn" → SM   |   "us" → RSA/ECDSA        │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Config Load → Parse cipherSuites → Match Provider → Initialize TLS
```

## Interfaces

### Modified Interfaces

```go
// infra/conf/config.go
type TLSConfig struct {
    // ... existing fields ...

    // CipherSuites - now dual-purpose:
    // 1. Original: comma-separated cipher suite names
    // 2. New: crypto provider identifier ("ru", "cn", "us")
    // First matching provider identifier wins
    CipherSuites   string `json:"cipherSuites,omitempty"`

    // Deprecated: use CipherSuites instead
    CryptoProvider string `json:"cryptoProvider,omitempty"`

    // ... existing fields ...
}
```

### New Functions

```go
// core/core.go

// selectCryptoProvider parses the cipherSuites string and returns
// the first matching crypto provider, or falls back to cryptoProvider
// field, or defaults to "us".
func selectCryptoProvider(cipherSuites, cryptoProvider string) crypto.Provider
```

## Data Models

No new data types required. Reusing existing configuration structures.

## Behavior Specifications

### Happy Path

1. User specifies `cipherSuites: "ru"` in config
2. System parses the string
3. System finds "ru" in `crypto.List()`
4. System initializes GOST crypto provider
5. TLS connections use GOST cryptography

### Parsing Algorithm

```
Input: cipherSuites string
Output: crypto.Provider

1. If cipherSuites is empty, goto step 6
2. Split cipherSuites by comma
3. For each part (trimmed):
   a. If part is in crypto.List():
      - Return corresponding crypto.Provider
4. No provider found in cipherSuites
5. (Implicit: cipherSuites might contain standard cipher names, ignore them)
6. If cryptoProvider is not empty:
   a. If cryptoProvider is in crypto.List():
      - Return corresponding crypto.Provider
   b. Log warning: deprecated field used
7. Return default "us" provider
```

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Empty cipherSuites | `cipherSuites: ""` | Use cryptoProvider fallback, then "us" |
| Mixed content | `cipherSuites: "ru,TLS_AES_256"` | Use "ru" (first valid provider) |
| Unknown provider | `cipherSuites: "xyz"` | Treat as cipher name, use fallback |
| Provider not compiled | `cipherSuites: "cn"` but SM not in build | Skip, try next, or fallback |
| Multiple providers | `cipherSuites: "ru,cn"` | Use first one ("ru") |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| No valid provider | All specified providers unavailable | Default to "us" with warning log |
| Deprecated field used | cryptoProvider specified | Log deprecation warning, still honor |

## Dependencies

### Requires

- `crypto` package with `List()` function returning available providers
- Provider implementations registered at init time

### Blocks

- Nothing (this is a configuration enhancement)

## Integration Points

### External Systems

- Xray/V2Ray GUI tools (must not break compatibility)
- Standard V2Ray clients (must accept the config format)

### Internal Systems

- `crypto` package provider registry
- TLS initialization code

## Testing Strategy

### Unit Tests

- [ ] `selectCryptoProvider()` - parse various cipherSuites strings
- [ ] Fallback logic - cryptoProvider deprecation path
- [ ] Default behavior - empty config

### Integration Tests

- [ ] Server starts with `cipherSuites: "ru"` and uses GOST
- [ ] Server starts with `cipherSuites: "cn"` and uses SM
- [ ] Server starts with empty config and uses RSA/ECDSA

### Manual Verification

- [ ] Load config in standard V2Ray GUI - no schema errors
- [ ] Connect client with each crypto provider

## Migration / Rollout

1. Existing configs with `cryptoProvider` continue to work (deprecated)
2. New configs should use `cipherSuites` for provider selection
3. Documentation should recommend `cipherSuites` over `cryptoProvider`

## Open Design Questions

- [ ] Should invalid provider names generate an error or just a warning?
- [ ] Is logging sufficient for deprecation, or should we also emit metrics?

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
