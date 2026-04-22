# Status: sdd-https-vpn-ciphersuite-cn

## Current Phase

IMPLEMENTATION

## Phase Status

PARTIAL COMPLETE (SM9 deferred)

## Last Updated

2026-04-22 by Claude

## Blockers

- None

## Progress

- [x] Requirements drafted
- [x] Requirements approved
- [x] Specifications drafted
- [x] Specifications approved
- [x] Plan drafted
- [x] Plan approved
- [x] Implementation started
- [x] Implementation complete (core: SM2/SM3/SM4, SM9 deferred)

## Context Notes

Key decisions and context for resuming:

- Chinese national cryptography (SM series) implementation
- Will follow the same pattern as GOST (Russian) crypto implementation
- Provider identifier: "cn"
- Algorithms: SM2 (signatures), SM3 (hash), SM4 (symmetric encryption), SM9 (identity-based crypto)
- TLS cipher suites: TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3 (all per RFC 8998)
- Part of the crypto provider system defined in `crypto/provider.go`
- SM2 curve: Only standard SM2-P256

## Related Flows

- `sdd-https-vpn-multi-cert` - Multi-provider certificate selection (parallel)
- `sdd-vpn-https-config-ciphersuites` - Provider selection (complete)

## Fork History

N/A - New flow

## Next Actions

1. (Optional) Implement SM9 (BN256 pairings) for identity-based crypto
2. Use provider with `cipherSuites: "cn"` in config
