# Implementation Plan: https-vpn-amnezia-server-backward-compatibility

> Version: 1.0  
> Status: DRAFT  
> Last Updated: 2026-04-21  
> Specifications: [link to 02-specifications.md](02-specifications.md)

## Summary

Phase 1 will focus on researching the Amnezia handshake protocol. Phase 2 will implement the transport handler. Phase 3 will integrate it with the main server core. Phase 4 will be validation.

## Task Breakdown

### Phase 1: Research & Discovery

#### Task 1.1: Document Amnezia Handshake Protocol
- **Description**: Analyze the Amnezia client source code to understand the handshake.
- **Files**: None (internal documentation/spec update)
- **Dependencies**: None
- **Verification**: Accurate protocol specification in 02-specifications.md.
- **Complexity**: Medium

### Phase 2: Transport Layer Updates

#### Task 2.1: Add Amnezia Handshake Detection
- **Description**: Add logic to `transport/server.go` to detect incoming Amnezia sessions.
- **Files**: 
  - `transport/server.go` - Modify
- **Dependencies**: Task 1.1
- **Verification**: Unit tests in `transport/transport_test.go`.
- **Complexity**: Medium

### Phase 3: Core Integration

#### Task 3.1: Implement Protocol Translator
- **Description**: Create the mapping between Amnezia session and `xray-core` context.
- **Files**: 
  - `core/core.go` - Modify
- **Dependencies**: Task 2.1
- **Verification**: Successful connection in mock test environment.
- **Complexity**: High

### Phase 4: Configuration Support

#### Task 4.1: Update Config Loader
- **Description**: Support `amnezia_compat` settings in server JSON.
- **Files**: 
  - `infra/conf/config.go` - Modify
- **Dependencies**: Task 3.1
- **Verification**: New config parameters are correctly loaded.
- **Complexity**: Low

## Dependency Graph

```
Task 1.1 ──→ Task 2.1 ──→ Task 3.1 ──→ Task 4.1
```

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `transport/amnezia.go` | Create | Amnezia protocol handler logic. |
| `transport/server.go` | Modify | Detect incoming Amnezia handshakes. |
| `core/core.go` | Modify | Support routing translated sessions. |
| `infra/conf/config.go` | Modify | Config loader for Amnezia parameters. |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Incomplete protocol spec | High | Failed connections | Extensive packet capture and analysis. |
| Breaking existing connections | Low | Server downtime | Comprehensive regression testing. |

## Rollback Strategy

Standard `git revert` on the affected files.

## Checkpoints

- [ ] Amnezia handshake correctly parsed.
- [ ] Session established without encryption errors.
- [ ] Data flowing through the tunnel.

## Open Implementation Questions

- [ ] Should we support all Amnezia client versions or target a specific range?

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
