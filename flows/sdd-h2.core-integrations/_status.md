# Status: sdd-h2.core-integrations

## Current Phase

REQUIREMENTS

## Phase Status

DRAFTING

## Last Updated

2026-05-14 by Claude

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

- h2.core is currently CLI-only (cmd/https-vpn)
- No C-API/CGO exports exist - pure Go binary
- Has both server (H2Server) and client (H2Client) components
- Client implements net.Dialer interface via DialContext
- Need to add integration layer for external consumers

## Integration Targets Identified

1. **vpnclient_engine_flutter** - Flutter VPN client engine
2. **gomobile** - iOS/Android library via gomobile
3. **C-API** - For native integration (CGO exports)
4. **HTTP API** - For remote control/management

## Fork History

- Not forked
- Created fresh for integration planning

## Next Actions

1. Complete requirements elicitation
2. Get user approval on requirements
3. Draft specifications for each integration type
