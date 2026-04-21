# Requirements: https-vpn-amnezia-server-backward-compatibility

> Version: 1.0  
> Status: DRAFT  
> Last Updated: 2026-04-21

## Problem Statement

The `https-vpn` system needs to maintain backward compatibility with `amnezia-server` protocols and configurations. This ensures that users transitioning from Amnezia or using Amnezia clients can connect to the `https-vpn` infrastructure without issues.

## User Stories

### Primary

**As a** VPN user  
**I want** to use my existing Amnezia client configuration  
**So that** I can connect to the new `https-vpn` server without needing to change my client software.

### Secondary

**As a** system administrator  
**I want** to deploy `https-vpn` as a drop-in replacement or augmentation for `amnezia-server`  
**So that** I can leverage the security and performance benefits of `https-vpn` while supporting legacy clients.

## Acceptance Criteria

### Must Have

1. **Given** a standard Amnezia client configuration  
   **When** the client attempts to connect to `https-vpn` server  
   **Then** the server should authenticate and establish a secure VPN tunnel.

2. **Given** an `amnezia-server` instance  
   **When** `https-vpn` is configured to run alongside or as a replacement  
   **Then** it should be able to handle incoming requests that follow the Amnezia protocol.

### Should Have

- Compatibility with various Amnezia-supported protocols (Shadowsocks, OpenVPN via cloak, etc. if applicable to `https-vpn`).

### Won't Have (This Iteration)

- Full feature parity with all obscure Amnezia server features not related to core VPN connectivity.

## Constraints

- **Technical**: Must interface with existing `xray-core` if that's what `https-vpn` uses for its core engine.
- **Platform**: Must work across the same OS platforms as `https-vpn`.

## Open Questions

- [ ] Which specific Amnezia protocols are most critical for initial backward compatibility?
- [ ] Are there any proprietary Amnezia handshake extensions that need to be implemented?

## References

- [Amnezia Server Documentation](https://github.com/amnezia-vpn/amnezia-server)

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
