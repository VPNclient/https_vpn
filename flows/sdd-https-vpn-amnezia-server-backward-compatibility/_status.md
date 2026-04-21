# Status: sdd-https-vpn-amnezia-server-backward-compatibility

## Current Phase

REQUIREMENTS

## Phase Status

DRAFTING

## Last Updated

2026-04-21 by Gemini

## Blockers

- [None]

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

- **Pivot**: The goal is to integrate OUR protocols and transport (specifically `h2`) INTO the Amnezia ecosystem, rather than just being backward compatible with their old protocols.
- We need to define how `https-vpn` acts as a backend for Amnezia's deployment scripts and client configs.

## Next Actions

1. Refine `01-requirements.md` to focus on `h2` transport integration into Amnezia.
2. Research Amnezia's "Custom Protocol" or "Plugin" architecture.
