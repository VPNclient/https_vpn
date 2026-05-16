# Plan: h2-cli-release

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-05-16

## Overview

Build script and release process for 32 cross-platform CLI binaries.

## Tasks

### Phase 1: Build Script

| # | Task | Files | Complexity |
|---|------|-------|------------|
| 1.1 | Update release.sh with full platform matrix | `build/release.sh` | Low |
| 1.2 | Add Win7 compatibility build (Go 1.20) | `build/release.sh` | Medium |
| 1.3 | Add checksum generation | `build/release.sh` | Low |

### Phase 2: Verification

| # | Task | Files | Complexity |
|---|------|-------|------------|
| 2.1 | Test build on Linux (native) | - | Low |
| 2.2 | Verify static linking (ldd) | - | Low |
| 2.3 | Test version command | - | Low |

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `build/release.sh` | Modify | Complete platform matrix, checksums |
| `dist/` | Create | Output directory (gitignored) |
| `.gitignore` | Modify | Add dist/ |

## Dependencies

```
Phase 1 ─┬─► Phase 2 ─► Phase 3 (optional)
         │
         └─► Can run independently
```

## Execution Order

1. **1.1** Update platform matrix in release.sh
2. **1.2** Add Win7 support section
3. **1.3** Add SHA256 checksum generation
4. **2.1-2.3** Run test build and verify

## Estimated Scope

- **Files changed**: 2
- **New files**: 0 (dist/ is output)
- **Lines of code**: ~100

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-05-16
- [x] Notes: CI/CD deferred to separate SDD flow
