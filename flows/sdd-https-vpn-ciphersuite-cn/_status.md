# Status: sdd-https-vpn-ciphersuite-cn

## Current Phase

REQUIREMENTS

## Phase Status

DRAFTING

## Last Updated

2026-04-22 by Claude

## Blockers

- None

## Progress

- [ ] Requirements drafted
- [ ] Requirements approved
- [ ] Specifications drafted
- [ ] Specifications approved
- [ ] Plan drafted
- [ ] Plan approved
- [ ] Implementation started
- [ ] Implementation complete

## Context Notes

Key decisions and context for resuming:

- Chinese national cryptography (SM series) implementation
- Will follow the same pattern as GOST (Russian) crypto implementation
- Provider identifier: "cn"
- Algorithms: SM2 (signatures), SM3 (hash), SM4 (symmetric encryption), SM9 (identity-based crypto)
- TLS cipher suites: TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3 (all per RFC 8998)
- Part of the crypto provider system defined in `crypto/provider.go`

## Fork History

N/A - New flow

## Next Actions

1. Elicit detailed requirements from user
2. Define scope and constraints
3. Get requirements approved
