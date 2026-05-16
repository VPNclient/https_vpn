# Specifications: h2-cli-release

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-05-16

## Overview

Cross-platform CLI release build system for h2.core HTTPS VPN.

## Architecture

```
cmd/https-vpn/main.go
        │
        ▼
  build/release.sh
        │
        ├── GOOS=linux GOARCH=amd64 → h2-linux-64.zip
        ├── GOOS=linux GOARCH=arm64 → h2-linux-arm64-v8a.zip
        ├── GOOS=darwin GOARCH=amd64 → h2-macos-64.zip
        ├── GOOS=windows GOARCH=amd64 → h2-windows-64.zip
        └── ... (30+ targets)
        │
        ▼
      dist/*.zip
```

## Build Configuration

### Go Build Flags

```bash
CGO_ENABLED=0           # Static linking, no CGO
-trimpath               # Remove file paths from binary
-ldflags="-s -w"        # Strip debug info
-ldflags="-X main.Version=${VERSION}"  # Inject version
```

### Platform Matrix

| Archive | GOOS | GOARCH | GOARM | Notes |
|---------|------|--------|-------|-------|
| **Linux** |
| h2-linux-64.zip | linux | amd64 | - | x86_64 |
| h2-linux-32.zip | linux | 386 | - | x86 |
| h2-linux-arm64-v8a.zip | linux | arm64 | - | ARM64 |
| h2-linux-arm32-v7a.zip | linux | arm | 7 | ARMv7 |
| h2-linux-arm32-v6.zip | linux | arm | 6 | ARMv6 (RPi1) |
| h2-linux-arm32-v5.zip | linux | arm | 5 | ARMv5 |
| h2-linux-mips32.zip | linux | mips | - | MIPS BE |
| h2-linux-mips32le.zip | linux | mipsle | - | MIPS LE |
| h2-linux-mips64.zip | linux | mips64 | - | MIPS64 BE |
| h2-linux-mips64le.zip | linux | mips64le | - | MIPS64 LE |
| h2-linux-ppc64.zip | linux | ppc64 | - | PPC64 BE |
| h2-linux-ppc64le.zip | linux | ppc64le | - | PPC64 LE |
| h2-linux-riscv64.zip | linux | riscv64 | - | RISC-V 64 |
| h2-linux-s390x.zip | linux | s390x | - | IBM Z |
| h2-linux-loong64.zip | linux | loong64 | - | LoongArch |
| **macOS** |
| h2-macos-64.zip | darwin | amd64 | - | Intel |
| h2-macos-arm64-v8a.zip | darwin | arm64 | - | Apple Silicon |
| **FreeBSD** |
| h2-freebsd-64.zip | freebsd | amd64 | - | |
| h2-freebsd-32.zip | freebsd | 386 | - | |
| h2-freebsd-arm64-v8a.zip | freebsd | arm64 | - | |
| h2-freebsd-arm32-v7a.zip | freebsd | arm | 7 | |
| **OpenBSD** |
| h2-openbsd-64.zip | openbsd | amd64 | - | |
| h2-openbsd-32.zip | openbsd | 386 | - | |
| h2-openbsd-arm64-v8a.zip | openbsd | arm64 | - | |
| h2-openbsd-arm32-v7a.zip | openbsd | arm | 7 | |
| **Windows** |
| h2-windows-64.zip | windows | amd64 | - | Win10+ |
| h2-windows-32.zip | windows | 386 | - | Win10+ |
| h2-windows-arm64-v8a.zip | windows | arm64 | - | WoA |
| h2-win7-64.zip | windows | amd64 | - | Win7 compat |
| h2-win7-32.zip | windows | 386 | - | Win7 compat |
| **Android** |
| h2-android-amd64.zip | android | amd64 | - | x86_64 emu |
| h2-android-arm64-v8a.zip | android | arm64 | - | ARM64 |

## File Structure

### Output

```
dist/
├── h2-linux-64.zip
│   └── https-vpn          # Binary
├── h2-windows-64.zip
│   └── https-vpn.exe      # Binary with .exe
├── ...
└── checksums.txt          # SHA256 checksums
```

### Build Script

**File**: `build/release.sh`

```bash
#!/bin/bash
# Usage: ./build/release.sh [VERSION]
# Output: dist/*.zip

VERSION="${1:-0.1.0}"
OUTPUT_DIR="dist"

# Build each target
for target in "${TARGETS[@]}"; do
    build_target "$target" "$VERSION"
done

# Generate checksums
cd "$OUTPUT_DIR"
sha256sum *.zip > checksums.txt
```

## CLI Interface

**Binary**: `https-vpn` (or `https-vpn.exe` on Windows)

```
https-vpn - HTTPS VPN over HTTP/2

Commands:
  run       Run server/client
  init      Generate config template
  version   Show version

Usage:
  https-vpn run -c config.json
  https-vpn init -crypto us
  https-vpn version
```

## Win7 Compatibility

For `h2-win7-*` targets, build with Go 1.20 (last version supporting Win7):

```bash
# Use Go 1.20 for Win7 builds
GO120=/path/to/go1.20/bin/go
$GO120 build -o dist/h2-win7-64/https-vpn.exe ./cmd/https-vpn
```

## Edge Cases

| Case | Handling |
|------|----------|
| Build fails for platform | Skip, log warning, continue others |
| Missing Go version for Win7 | Skip Win7 builds with warning |
| Binary too large | Consider UPX compression (optional) |

## Testing

### Verification

```bash
# Check binary runs
./dist/h2-linux-64/https-vpn version

# Check static linking (Linux)
ldd ./dist/h2-linux-64/https-vpn
# Should show: "not a dynamic executable"

# Check file size
ls -lh dist/*.zip
```

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-05-16
