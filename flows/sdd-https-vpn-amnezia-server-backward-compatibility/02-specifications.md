# Specifications: https-vpn-amnezia-server-backward-compatibility

> Version: 1.0  
> Status: DRAFT  
> Last Updated: 2026-04-21  
> Requirements: [link to 01-requirements.md](01-requirements.md)

## Overview

Backward compatibility for `amnezia-server` will be achieved by implementing a protocol translation layer or a handler that can understand the Amnezia-specific handshakes and map them to the underlying `https-vpn` transport. This will leverage `xray-core`'s flexibility where possible.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `transport` | Modify | Need to add handlers for Amnezia-compatible protocols. |
| `core` | Modify | Core logic needs to understand the new transport options. |
| `infra/conf` | Modify | Configuration loader must support Amnezia-specific parameters. |

## Architecture

### Component Diagram

```
[ Amnezia Client ] --(Amnezia Protocol/Handshake)--> [ https-vpn Server ]
                                                      |
                                                      v
                                              [ Protocol Translator ]
                                                      |
                                                      v
                                              [ Core VPN Engine ]
```

### Data Flow

1. Incoming connection from Amnezia client.
2. `https-vpn` identifies the handshake as Amnezia-style.
3. Translator maps Amnezia session parameters to `https-vpn` internals.
4. Tunnel is established and data starts flowing.

## Interfaces

### New Interfaces

[TBD]

### Modified Interfaces

- `transport.Server`: Needs to accept and process Amnezia-style incoming connections.

## Data Models

### New Types

- `AmneziaConfig`: Structure to represent Amnezia-compatible configuration parameters.

### Schema Changes

- Updated server configuration to include an `amnezia_compat` section.

## Behavior Specifications

### Happy Path

1. User imports Amnezia config into an Amnezia client.
2. Client initiates connection to `https-vpn` server.
3. Server detects the Amnezia protocol.
4. Server completes handshake using Amnezia parameters.
5. VPN tunnel is active.

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Invalid Amnezia Secret | Client uses wrong secret | Server refuses connection with Amnezia error code. |
| Protocol Version Mismatch | Client uses outdated version | Server sends version unsupported message. |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| Authentication Failed | Wrong credentials | Log error and drop connection. |
| Unsupported Sub-Protocol | Client requests ShadowSocksR but only ShadowSocks is supported | Inform client of available protocols. |

## Dependencies

### Requires

- `3rdparty/xray-core`

### Blocks

- None

## Integration Points

### External Systems

- Amnezia Client software

### Internal Systems

- `transport`, `core`, `infra/conf`

## Testing Strategy

### Unit Tests

- [ ] `transport/amnezia_test.go` - Test Amnezia handshake logic.

### Integration Tests

- [ ] Connect a real Amnezia client to a test `https-vpn` server instance.

### Manual Verification

- [ ] Verify connectivity using the official Amnezia VPN application.

## Migration / Rollout

[TBD]

## Open Design Questions

- [ ] How to handle Amnezia's dynamic port selection if applicable?

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
