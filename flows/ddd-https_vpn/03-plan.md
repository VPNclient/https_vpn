# Implementation Plan: HTTPS VPN

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-03-10
> Specifications: [02-specifications.md](./02-specifications.md)

## Summary

Implementation of ~540 LOC HTTPS VPN in 4 phases:
1. **Foundation** - project structure, crypto interface
2. **Transport** - HTTP/2 CONNECT server/client
3. **Integration** - config parser, core instance
4. **Polish** - CLI, tests, documentation

## Task Breakdown

### Phase 1: Foundation (~80 LOC)

#### Task 1.1: Project Setup
- **Description**: Initialize Go module, create directory structure
- **Files**:
  - `go.mod` - Create
  - `go.sum` - Create
- **Dependencies**: None
- **Verification**: `go mod tidy` succeeds
- **Complexity**: Low

#### Task 1.2: Crypto Provider Interface
- **Description**: Define Provider interface and Registry
- **Files**:
  - `crypto/provider.go` - Create (~30 LOC)
- **Dependencies**: Task 1.1
- **Verification**: Code compiles
- **Complexity**: Low

#### Task 1.3: US Crypto Provider (stdlib)
- **Description**: Implement default US/NIST provider using Go stdlib
- **Files**:
  - `crypto/us/provider.go` - Create (~20 LOC)
- **Dependencies**: Task 1.2
- **Verification**: `go test ./crypto/...` passes
- **Complexity**: Low

### Phase 2: Transport (~210 LOC)

#### Task 2.1: Bidirectional Pipe
- **Description**: Implement io.Copy in both directions with proper cleanup
- **Files**:
  - `transport/pipe.go` - Create (~30 LOC)
- **Dependencies**: Task 1.1
- **Verification**: Unit test with mock connections
- **Complexity**: Low

#### Task 2.2: CONNECT Handler
- **Description**: HTTP handler for CONNECT method, dial target, pipe data
- **Files**:
  - `transport/handler.go` - Create (~50 LOC)
- **Dependencies**: Task 2.1
- **Verification**: Unit test with httptest
- **Complexity**: Medium

#### Task 2.3: HTTP/2 Server
- **Description**: TLS listener with HTTP/2, uses crypto provider
- **Files**:
  - `transport/server.go` - Create (~60 LOC)
- **Dependencies**: Task 1.2, Task 2.2
- **Verification**: Server starts, accepts TLS connection
- **Complexity**: Medium

#### Task 2.4: HTTP/2 Client
- **Description**: Connect to server, send CONNECT, return tunnel connection
- **Files**:
  - `transport/client.go` - Create (~70 LOC)
- **Dependencies**: Task 1.2
- **Verification**: Client connects to test server
- **Complexity**: Medium

### Phase 3: Integration (~230 LOC)

#### Task 3.1: Config Structs
- **Description**: xray-compatible config structures
- **Files**:
  - `infra/conf/config.go` - Create (~80 LOC)
- **Dependencies**: Task 1.1
- **Verification**: JSON unmarshal test
- **Complexity**: Low

#### Task 3.2: Config Loader
- **Description**: Load and validate config from file
- **Files**:
  - `infra/conf/loader.go` - Create (~70 LOC)
- **Dependencies**: Task 3.1, Task 1.2
- **Verification**: Load sample config, validate crypto provider exists
- **Complexity**: Low

#### Task 3.3: Core Instance
- **Description**: Main entry point, xray-compatible New/Start/Close
- **Files**:
  - `core/core.go` - Create (~80 LOC)
- **Dependencies**: Task 2.3, Task 2.4, Task 3.2
- **Verification**: Instance starts server, accepts connection
- **Complexity**: Medium

### Phase 4: Polish (~50 LOC + tests)

#### Task 4.1: CLI Entry Point
- **Description**: Command-line interface with run/init commands
- **Files**:
  - `cmd/https-vpn/main.go` - Create (~50 LOC)
- **Dependencies**: Task 3.3
- **Verification**: `go build ./cmd/https-vpn` succeeds
- **Complexity**: Low

#### Task 4.2: Integration Tests
- **Description**: End-to-end tunnel test
- **Files**:
  - `test/integration_test.go` - Create
- **Dependencies**: All previous tasks
- **Verification**: Full tunnel works: client -> server -> target
- **Complexity**: Medium

#### Task 4.3: Sample Configs
- **Description**: Example config files for server and client
- **Files**:
  - `examples/server.json` - Create
  - `examples/client.json` - Create
- **Dependencies**: Task 3.1
- **Verification**: Configs parse without error
- **Complexity**: Low

## Dependency Graph

```
Task 1.1 (setup)
    │
    ├──► Task 1.2 (crypto interface)
    │        │
    │        ├──► Task 1.3 (us provider)
    │        │
    │        ├──► Task 2.3 (server) ◄── Task 2.2 (handler) ◄── Task 2.1 (pipe)
    │        │        │
    │        │        └──► Task 3.3 (core) ◄── Task 3.2 (loader) ◄── Task 3.1 (config)
    │        │                  │
    │        └──► Task 2.4 (client)
    │                  │
    │                  └──► Task 3.3 (core)
    │                            │
    │                            └──► Task 4.1 (cli)
    │                                      │
    │                                      └──► Task 4.2 (tests)
    │
    └──► Task 4.3 (examples)
```

## File Change Summary

| File | Action | LOC | Phase |
|------|--------|-----|-------|
| `go.mod` | Create | 5 | 1 |
| `crypto/provider.go` | Create | 30 | 1 |
| `crypto/us/provider.go` | Create | 20 | 1 |
| `transport/pipe.go` | Create | 30 | 2 |
| `transport/handler.go` | Create | 50 | 2 |
| `transport/server.go` | Create | 60 | 2 |
| `transport/client.go` | Create | 70 | 2 |
| `infra/conf/config.go` | Create | 80 | 3 |
| `infra/conf/loader.go` | Create | 70 | 3 |
| `core/core.go` | Create | 80 | 3 |
| `cmd/https-vpn/main.go` | Create | 50 | 4 |
| **TOTAL** | | **~545** | |

## Implementation Order

Optimal order for incremental development and testing:

```
1. Task 1.1  → go.mod
2. Task 1.2  → crypto/provider.go
3. Task 1.3  → crypto/us/provider.go
   ✓ Checkpoint: crypto package compiles and tests pass

4. Task 2.1  → transport/pipe.go
5. Task 2.2  → transport/handler.go
6. Task 2.3  → transport/server.go
7. Task 2.4  → transport/client.go
   ✓ Checkpoint: transport package compiles, server accepts connections

8. Task 3.1  → infra/conf/config.go
9. Task 3.2  → infra/conf/loader.go
   ✓ Checkpoint: config loading works with sample JSON

10. Task 3.3 → core/core.go
    ✓ Checkpoint: core.New() + Start() works

11. Task 4.1 → cmd/https-vpn/main.go
    ✓ Checkpoint: binary builds and runs

12. Task 4.2 → test/integration_test.go
13. Task 4.3 → examples/*.json
    ✓ Checkpoint: full integration test passes
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| HTTP/2 CONNECT not working as expected | Low | High | Use Go stdlib net/http which supports this natively |
| TLS fingerprint differs from browser | Medium | Medium | Add uTLS later if needed |
| xray config incompatibility | Low | Medium | Test with real 3x-ui configs early |
| LOC budget exceeded | Low | Low | Core functionality prioritized, edge cases can be deferred |

## Rollback Strategy

Each phase is independently testable. If issues arise:

1. **Phase 1 issues**: Fix crypto interface, no downstream impact
2. **Phase 2 issues**: Transport layer isolated, can swap implementations
3. **Phase 3 issues**: Config/core can be refactored without transport changes
4. **Phase 4 issues**: CLI is thin wrapper, easy to modify

Git strategy: commit after each task, tag after each phase.

```
git tag v0.1.0  # After Phase 1
git tag v0.2.0  # After Phase 2
git tag v0.3.0  # After Phase 3
git tag v0.4.0  # After Phase 4 (release candidate)
```

## Checkpoints

### After Phase 1
- [ ] `go build ./...` succeeds
- [ ] `go test ./crypto/...` passes
- [ ] Provider interface is clean and minimal

### After Phase 2
- [ ] Server starts on specified port
- [ ] Client can establish TLS connection
- [ ] CONNECT request is handled correctly
- [ ] Bidirectional data flow works

### After Phase 3
- [ ] xray-style JSON config parses
- [ ] `core.New(config)` returns valid instance
- [ ] `instance.Start()` begins accepting connections
- [ ] `instance.Close()` shuts down cleanly

### After Phase 4
- [ ] `https-vpn run -c config.json` works
- [ ] Integration test passes
- [ ] Example configs are valid
- [ ] README is accurate

## Open Implementation Questions

*None - all questions resolved in specifications*

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
