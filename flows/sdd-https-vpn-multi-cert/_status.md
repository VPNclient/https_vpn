# Status: sdd-https-vpn-multi-cert

## Current Phase

REQUIREMENTS

## Phase Status

AWAITING APPROVAL

## Last Updated

2026-04-22 by Claude

## Blockers

- None

## Progress

- [x] Requirements drafted
- [ ] Requirements approved
- [ ] Specifications drafted
- [ ] Specifications approved
- [ ] Plan drafted
- [ ] Plan approved
- [ ] Implementation started
- [ ] Implementation complete

## Context Notes

Key decisions and context for resuming:

- Multi-provider certificate auto-selection for TLS
- Allows server to serve different certificates based on client capabilities
- Works with all crypto providers: "us", "ru", "cn"
- Priority determined by order in `cipherSuites` config
- Can be implemented in parallel with `sdd-https-vpn-ciphersuite-cn`

## Related Flows

- `sdd-https-vpn-ciphersuite-cn` - Chinese crypto (parallel)
- `sdd-vpn-https-config-ciphersuites` - Provider selection (complete)

## Fork History

N/A - New flow

## Next Actions

1. Get requirements approved ("requirements approved")
2. Begin specifications phase
