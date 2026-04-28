# Specifications: HTTPS VPN CI/CD Pipeline

> Version: 1.1
> Status: APPROVED
> Last Updated: 2026-04-28
> Requirements: [01-requirements.md](01-requirements.md)

## Overview

Создание отказоустойчивого CI/CD pipeline на базе GitHub Actions для сборки HTTPS VPN под множество платформ. Pipeline состоит из изолированных workflow файлов, которые могут падать независимо друг от друга.

## Affected Systems

| System | Impact | Notes |
|--------|--------|-------|
| `.github/workflows/release.yml` | Modify | Преобразовать в orchestrator |
| `.github/workflows/build-*.yml` | Create | 6 новых workflow файлов |
| `.github/workflows/_*.yml` | Create | 4 reusable workflow файла |
| `.github/build/friendly-filenames.json` | Modify | Добавить новые платформы |
| `VERSION` | Create | Файл версии для автоинкремента |

## Architecture

### Workflow Structure

```
                              ┌─────────────────────┐
                              │    release.yml      │
                              │   (Orchestrator)    │
                              └──────────┬──────────┘
                                         │
    ┌──────────┬──────────┬──────────────┼──────────────┬──────────┬──────────┐
    ▼          ▼          ▼              ▼              ▼          ▼          ▼
┌────────┐┌────────┐┌──────────┐┌──────────────┐┌────────┐┌────────┐┌──────────┐
│ build- ││ build- ││  build-  ││    build-    ││ build- ││ build- ││  build-  │
│desktop ││ mobile ││ routers  ││ mikrotik-npk ││  gov   ││  libs  ││(elbrus) │
└────────┘└────────┘└──────────┘└──────────────┘└────────┘└────────┘└──────────┘
    │          │          │              │              │          │
    └──────────┴──────────┴──────────────┼──────────────┴──────────┘
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Reusable Workflows                                   │
├─────────────┬──────────────┬──────────────┬─────────────────────────────────┤
│ _build-base │_build-library│_build-package│         _sign-apple             │
└─────────────┴──────────────┴──────────────┴─────────────────────────────────┘
```

### Build Matrix

```
build-desktop.yml
├── Windows: amd64, arm64, 386
├── Linux: amd64, arm64, arm(5,6,7), mips*, riscv64, loong64, ppc64*, s390x
├── macOS: amd64, arm64 (signed + notarized)
├── FreeBSD: amd64, arm64, arm, 386
└── OpenBSD: amd64, arm64, arm, 386

build-mobile.yml
├── Android: arm64, amd64
├── iOS: arm64, simulator (signed framework)
└── Aurora OS: arm64 (RPM package)

build-routers.yml
├── OpenWRT: mips, mipsle, arm, arm64, amd64 (static binary)
├── MikroTik: arm, arm64, mipsbe (static binary only)
└── Cisco: linux-amd64 (static binary)

build-mikrotik-npk.yml (отдельно, требует RouterOS SDK)
└── MikroTik: arm, arm64, mipsbe (.npk packages)

build-gov.yml
├── Astra Linux: amd64 (.deb)
└── Elbrus: e2k (placeholder, self-hosted runner)

build-libraries.yml
├── Linux: .so, .a
├── macOS: .dylib, .a, .framework
├── Windows: .dll, .lib
├── Android: .so, .aar
└── iOS: .framework (signed)
```

## Interfaces

### Reusable Workflow: _build-base.yml

```yaml
# Inputs
inputs:
  goos:
    required: true
    type: string
  goarch:
    required: true
    type: string
  goarm:
    required: false
    type: string
    default: ''
  build_tags:
    required: false
    type: string
    default: ''
  cgo_enabled:
    required: false
    type: string
    default: '0'
  artifact_name:
    required: true
    type: string
  version:
    required: true
    type: string

# Outputs
outputs:
  artifact_path:
    description: "Path to built artifact"
  build_success:
    description: "Whether build succeeded"
```

### Reusable Workflow: _build-library.yml

```yaml
inputs:
  goos:
    required: true
    type: string
  goarch:
    required: true
    type: string
  library_type:
    required: true
    type: string  # 'static' | 'shared' | 'both'
  version:
    required: true
    type: string

outputs:
  static_lib_path:
    description: "Path to .a file"
  shared_lib_path:
    description: "Path to .so/.dylib/.dll file"
```

### Reusable Workflow: _build-package.yml

```yaml
inputs:
  package_type:
    required: true
    type: string  # 'deb' | 'rpm' | 'ipk' | 'npk'
  binary_path:
    required: true
    type: string
  version:
    required: true
    type: string
  arch:
    required: true
    type: string

outputs:
  package_path:
    description: "Path to package file"
```

### Reusable Workflow: _sign-apple.yml

```yaml
inputs:
  artifact_path:
    required: true
    type: string
  artifact_type:
    required: true
    type: string  # 'binary' | 'framework' | 'app'
  notarize:
    required: false
    type: boolean
    default: true

secrets:
  APPLE_CERTIFICATE:
    required: true
  APPLE_CERTIFICATE_PASSWORD:
    required: true
  APPLE_ID:
    required: true
  APPLE_TEAM_ID:
    required: true
  APPLE_APP_SPECIFIC_PASSWORD:
    required: true
```

## Data Models

### Version File (VERSION)

```
1.0.0
```

Семантическое версионирование. Build number добавляется из `github.run_number`.

### Artifact Naming Convention

```
https-vpn-v{VERSION}-build.{BUILD}-{PLATFORM}-{ARCH}[.{EXT}]

Examples:
https-vpn-v1.0.0-build.42-linux-amd64.tar.gz
https-vpn-v1.0.0-build.42-linux-amd64.so
https-vpn-v1.0.0-build.42-windows-64.zip
https-vpn-v1.0.0-build.42-windows-64.dll
https-vpn-v1.0.0-build.42-macos-arm64.tar.gz
https-vpn-v1.0.0-build.42-ios-arm64.framework.zip
https-vpn-v1.0.0-build.42-mikrotik-arm.npk
https-vpn-v1.0.0-build.42-openwrt-mips.ipk
https-vpn-v1.0.0-build.42-astra-amd64.deb
```

### friendly-filenames.json Update

```json
{
  "ios-arm64": { "friendlyName": "ios-arm64" },
  "ios-amd64": { "friendlyName": "ios-simulator" },
  "aurora-arm64": { "friendlyName": "aurora-arm64" },
  "astra-amd64": { "friendlyName": "astra-amd64" },
  "mikrotik-arm": { "friendlyName": "mikrotik-arm" },
  "mikrotik-arm64": { "friendlyName": "mikrotik-arm64" },
  "mikrotik-mipsbe": { "friendlyName": "mikrotik-mipsbe" },
  "openwrt-mips": { "friendlyName": "openwrt-mips" },
  "openwrt-mipsle": { "friendlyName": "openwrt-mipsle" },
  "openwrt-arm": { "friendlyName": "openwrt-arm" },
  "openwrt-arm64": { "friendlyName": "openwrt-arm64" },
  "cisco-amd64": { "friendlyName": "cisco-linux-64" }
}
```

## Behavior Specifications

### Happy Path: Release Build

1. User creates GitHub Release with tag `v1.2.3`
2. `release.yml` triggers and reads version from tag
3. All category workflows start in parallel
4. Each workflow builds its targets with `fail-fast: false`
5. Successful builds upload artifacts
6. `publish` job collects all artifacts
7. Artifacts uploaded to GitHub Release

### Happy Path: Push Build (CI)

1. Developer pushes to any branch
2. `release.yml` triggers in CI mode
3. Builds only for testing (no publishing)
4. Artifacts available as workflow artifacts (not release)

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Partial failure | One workflow fails | Other workflows continue; partial release with successful builds |
| Apple signing fails | Invalid certificate | macOS/iOS builds fail; other platforms succeed |
| MikroTik SDK unavailable | npk build fails | Raw binary still created |
| Self-hosted runner offline | Elbrus build | Skipped gracefully via `if` condition |
| Version tag invalid | Non-semver tag | Use `0.0.0-build.N` as fallback |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| Build failure | Compilation error | Mark job failed, continue others |
| Signing failure | Certificate expired | Alert, upload unsigned artifact with `-unsigned` suffix |
| Upload failure | GitHub API error | Retry 3 times with backoff |
| Timeout | Build > 60 min | Cancel and mark failed |

## Dependencies

### Requires

- Go 1.25+ (from go.mod)
- GitHub Actions runners (ubuntu-latest, macos-latest, windows-latest)
- Android NDK r28b
- Xcode (latest on macos-latest)
- Apple Developer Account secrets in GitHub

### External Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| `goreleaser` | Optional: unified release | `go install github.com/goreleaser/goreleaser` |
| `fpm` | Package creation (deb/rpm) | Ruby gem |
| `dpkg-deb` | DEB package | apt install |
| `rpmbuild` | RPM package | apt install rpm |
| `codesign` | macOS signing | Xcode CLI |
| `notarytool` | macOS notarization | Xcode CLI |

## Integration Points

### GitHub Secrets Required

```
APPLE_CERTIFICATE          # Base64-encoded .p12 certificate
APPLE_CERTIFICATE_PASSWORD # Certificate password
APPLE_ID                   # Apple ID email
APPLE_TEAM_ID              # Team ID (10-char)
APPLE_APP_SPECIFIC_PASSWORD # App-specific password for notarization
```

### GitHub Variables

```
ELBRUS_RUNNER_AVAILABLE    # 'true' to enable Elbrus builds
VERSION_OVERRIDE           # Override version (optional)
```

### Internal Systems

| System | Integration |
|--------|-------------|
| `crypto/` | Build tags for crypto providers |
| `cmd/https-vpn/` | Main binary entry point |
| `core/` | Library entry point for .so/.dll builds |

## Testing Strategy

### Unit Tests

- [ ] Version parsing from tag/file
- [ ] Artifact naming generation
- [ ] Build matrix generation

### Integration Tests

- [ ] Full build on ubuntu-latest (Linux amd64)
- [ ] Cross-compilation to ARM
- [ ] Library build produces valid .so

### Manual Verification

- [ ] Create test release, verify all artifacts present
- [ ] Download and run binary on target platform
- [ ] Verify iOS framework imports in Xcode
- [ ] Verify .npk installs on MikroTik RouterOS
- [ ] Verify .ipk installs on OpenWRT
- [ ] Verify .deb installs on Astra Linux

## Migration / Rollout

### Phase 1: Reusable Workflows

1. Create `_build-base.yml` with existing logic
2. Test with single platform

### Phase 2: Category Workflows

1. Create `build-desktop.yml` first (most targets)
2. Migrate from monolithic `release.yml`
3. Verify all existing builds work

### Phase 3: New Platforms

1. Add `build-mobile.yml` (iOS, Aurora)
2. Add `build-routers.yml` (OpenWRT, MikroTik, Cisco)
3. Add `build-gov.yml` (Astra)

### Phase 4: Libraries

1. Add `build-libraries.yml`
2. Implement CGO builds where needed

### Phase 5: Apple Signing

1. Configure secrets
2. Add `_sign-apple.yml`
3. Enable signing for macOS/iOS

## Open Design Questions

- [x] ~~CGO for crypto providers~~ → Start with CGO_ENABLED=0, add per-provider if needed
- [x] ~~MikroTik SDK~~ → Отдельный workflow `build-mikrotik-npk.yml`
- [x] ~~OpenWRT SDK~~ → Статические бинарники (без SDK)

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-28
- [x] Notes: MikroTik .npk в отдельном workflow; OpenWRT - статические бинарники
