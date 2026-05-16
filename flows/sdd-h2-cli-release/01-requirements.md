# Requirements: h2-cli-release

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-05-16

## Problem Statement

h2.core needs cross-platform CLI binary releases for direct deployment on servers and clients. Users should be able to download a single zip for their platform and run the HTTPS VPN immediately.

## User Stories

### Primary

**As a** system administrator
**I want** to download h2.core CLI for my platform
**So that** I can run an HTTPS VPN server/client without building from source

**As a** developer
**I want** automated release builds for all platforms
**So that** releases are consistent and reproducible

## Acceptance Criteria

### Must Have

1. **Given** the build script `build/release.sh`
   **When** running without arguments
   **Then** all 30+ platform binaries are built and zipped

2. **Given** a platform target (e.g., linux/amd64)
   **When** the binary is built
   **Then** it should be statically linked (CGO_ENABLED=0) and stripped (-s -w)

3. **Given** the built archives
   **When** checking file names
   **Then** they match the naming convention: `h2-{os}-{arch}.zip`

## Target Platforms

| Archive Name | GOOS | GOARCH | GOARM |
|-------------|------|--------|-------|
| h2-linux-64.zip | linux | amd64 | - |
| h2-linux-32.zip | linux | 386 | - |
| h2-linux-arm64-v8a.zip | linux | arm64 | - |
| h2-linux-arm32-v7a.zip | linux | arm | 7 |
| h2-linux-arm32-v6.zip | linux | arm | 6 |
| h2-linux-arm32-v5.zip | linux | arm | 5 |
| h2-linux-mips32.zip | linux | mips | - |
| h2-linux-mips32le.zip | linux | mipsle | - |
| h2-linux-mips64.zip | linux | mips64 | - |
| h2-linux-mips64le.zip | linux | mips64le | - |
| h2-linux-ppc64.zip | linux | ppc64 | - |
| h2-linux-ppc64le.zip | linux | ppc64le | - |
| h2-linux-riscv64.zip | linux | riscv64 | - |
| h2-linux-s390x.zip | linux | s390x | - |
| h2-linux-loong64.zip | linux | loong64 | - |
| h2-macos-64.zip | darwin | amd64 | - |
| h2-macos-arm64-v8a.zip | darwin | arm64 | - |
| h2-freebsd-64.zip | freebsd | amd64 | - |
| h2-freebsd-32.zip | freebsd | 386 | - |
| h2-freebsd-arm64-v8a.zip | freebsd | arm64 | - |
| h2-freebsd-arm32-v7a.zip | freebsd | arm | 7 |
| h2-openbsd-64.zip | openbsd | amd64 | - |
| h2-openbsd-32.zip | openbsd | 386 | - |
| h2-openbsd-arm64-v8a.zip | openbsd | arm64 | - |
| h2-openbsd-arm32-v7a.zip | openbsd | arm | 7 |
| h2-windows-64.zip | windows | amd64 | - |
| h2-windows-32.zip | windows | 386 | - |
| h2-windows-arm64-v8a.zip | windows | arm64 | - |
| h2-win7-64.zip | windows | amd64 | - |
| h2-win7-32.zip | windows | 386 | - |
| h2-android-amd64.zip | android | amd64 | - |
| h2-android-arm64-v8a.zip | android | arm64 | - |

## Constraints

- **No CGO**: All builds with `CGO_ENABLED=0` for static linking
- **No .so/.dylib**: This flow is CLI only, no shared libraries
- **No gomobile**: Mobile wrappers are in libh2, not h2.core

## Non-Goals

- iOS builds (CLI not applicable)
- Shared library builds (.so, .dylib, .dll)
- Gomobile frameworks (xcframework, aar)

## References

- Source: `cmd/https-vpn/main.go`
- Build script: `build/release.sh`

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-05-16
