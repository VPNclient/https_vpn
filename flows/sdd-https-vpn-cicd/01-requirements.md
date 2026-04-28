# Requirements: HTTPS VPN CI/CD Pipeline

> Version: 1.2
> Status: APPROVED
> Last Updated: 2026-04-28

## Problem Statement

Проект HTTPS VPN требует автоматизированного процесса сборки и релиза для поддержки широкого спектра платформ, архитектур процессоров и сетевого оборудования. Текущий CI/CD pipeline (унаследованный от xray-core) покрывает основные платформы, но не включает:

1. **Экзотические платформы**: iOS, Aurora OS, Astra Linux
2. **Специфические процессоры**: Elbrus (Эльбрус), Shakti (Индия), полная поддержка Loongson
3. **Сетевое оборудование**: Cisco, MikroTik, OpenWRT
4. **Формат библиотеки**: линкуемая библиотека (.a/.so/.dylib) в дополнение к бинарникам
5. **Автоинкремент версий**: автоматическая нумерация билдов

## User Stories

### Primary

**As a** DevOps инженер
**I want** автоматизированный CI/CD pipeline
**So that** каждый push/release автоматически собирает артефакты для всех целевых платформ

**As a** разработчик встраиваемых систем
**I want** линкуемую библиотеку (.a / .so)
**So that** я могу интегрировать HTTPS VPN в свои приложения

**As a** администратор сети
**I want** сборки для роутеров (Cisco, MikroTik, OpenWRT)
**So that** я могу развернуть VPN на уровне сетевого оборудования

### Secondary

**As a** разработчик мобильных приложений
**I want** сборки для iOS и Aurora OS
**So that** я могу использовать HTTPS VPN на мобильных устройствах

**As a** администратор государственных систем
**I want** сборки для Astra Linux и Эльбрус
**So that** я могу использовать сертифицированное ПО

## Acceptance Criteria

### Must Have

1. **Given** push в main или создание release
   **When** запускается GitHub Actions workflow
   **Then** собираются бинарники и библиотеки для всех поддерживаемых платформ

2. **Given** успешная сборка
   **When** создан release
   **Then** все артефакты загружаются в GitHub Releases с правильными именами

3. **Given** новый билд
   **When** отсутствует явный тег версии
   **Then** номер билда автоматически инкрементируется (формат: vX.Y.Z-build.N)

4. **Given** запрос на сборку библиотеки
   **When** указан build mode = library
   **Then** создаются статические (.a) и динамические (.so/.dylib/.dll) библиотеки

### Should Have

5. **Given** целевая платформа = iOS
   **When** запускается сборка
   **Then** создается signed framework для iOS (arm64, simulator)

6. **Given** целевая платформа = macOS
   **When** запускается release сборка
   **Then** бинарник подписан Developer ID и прошел notarization

7. **Given** целевая платформа = OpenWRT
   **When** указана архитектура роутера (mips, arm, etc.)
   **Then** создается .ipk пакет для embedded Linux

8. **Given** целевая платформа = MikroTik
   **When** указана архитектура (arm, arm64, mipsbe)
   **Then** создаются raw binary и .npk пакет

9. **Given** целевая платформа = Astra Linux
   **When** архитектура = amd64
   **Then** создается .deb пакет, совместимый с Astra Linux

### Won't Have (This Iteration)

- Сборка для Windows CE / Windows Mobile
- Сборка для QNX
- Поддержка архитектуры SPARC
- Поддержка процессора Elbrus (e2k) - отложено, нет поддержки в Go
- Автоматическая публикация в App Store / Google Play

## Constraints

### Technical

- **GitHub Actions**: использование стандартных runners (ubuntu-latest, macos-latest, windows-latest)
- **Cross-compilation**: использование Go cross-compilation где возможно
- **CGO**: минимизация CGO-зависимостей для упрощения cross-compilation
- **Build time**: общее время сборки не должно превышать 60 минут

### Platform-Specific Constraints

| Platform | Constraint | Solution |
|----------|------------|----------|
| **iOS** | Требует macOS runner + Xcode + signing | macos-latest + Apple Developer certs |
| **macOS** | Code signing + notarization | Apple Developer ID + notarytool |
| **Shakti** | Экспериментальная поддержка RISC-V | GOARCH=riscv64 |
| **Loongson** | GOARCH=loong64 в Go 1.19+ | Уже поддерживается |
| **Aurora OS** | Основана на Sailfish (Linux ARM) | Cross-compile linux-arm64 + RPM |
| **Astra Linux** | Debian-based | linux-amd64 + .deb packaging |
| **Cisco** | IOS-XE/NX-OS Linux-based | Статический linux-amd64 бинарник |
| **MikroTik** | RouterOS Linux-based | Cross-compile + .npk (RouterOS SDK) |
| **OpenWRT** | Использует musl libc | Статическая линковка CGO_ENABLED=0 |

### Dependencies

- Существующий workflow `release.yml` как база
- Go 1.25+ (из go.mod)
- NDK для Android
- Xcode для iOS
- Apple Developer Account (для code signing iOS/macOS)
- MikroTik RouterOS SDK (для .npk пакетов)

## Open Questions

- [x] ~~Доступен ли self-hosted runner для Elbrus?~~ → **Отложено** (нет поддержки в Go)
- [x] ~~Какой формат пакета для MikroTik?~~ → **Оба: raw binary + .npk пакет**
- [x] ~~Нужен ли code signing для iOS/macOS?~~ → **Да**, есть Apple Developer Account
- [x] ~~Какие модели Cisco?~~ → **linux-amd64** статический бинарник (ISR/Catalyst/Nexus на IOS-XE)
- [x] ~~Формат версионирования?~~ → **`v{MAJOR}.{MINOR}.{PATCH}-build.{N}`**
- [ ] Нужна ли поддержка CGO для криптопровайдеров (ГОСТ, SM)?

## Decisions Made

### Версионирование и имена файлов

Формат версии: `v{MAJOR}.{MINOR}.{PATCH}-build.{N}`

Где `N` = `github.run_number` (автоинкремент)

Примеры имен артефактов:
```
https-vpn-v1.2.3-build.147-linux-amd64.tar.gz
https-vpn-v1.2.3-build.147-linux-amd64.so
https-vpn-v1.2.3-build.147-windows-64.zip
https-vpn-v1.2.3-build.147-windows-64.dll
https-vpn-v1.2.3-build.147-macos-arm64.tar.gz
https-vpn-v1.2.3-build.147-macos-arm64.framework.zip
https-vpn-v1.2.3-build.147-ios-arm64.framework.zip
https-vpn-v1.2.3-build.147-openwrt-mips.ipk
https-vpn-v1.2.3-build.147-mikrotik-arm.npk
https-vpn-v1.2.3-build.147-mikrotik-arm.tar.gz
https-vpn-v1.2.3-build.147-astra-amd64.deb
https-vpn-v1.2.3-build.147-cisco-linux-64.tar.gz
```

### Apple Code Signing

- **iOS**: Обязательная подпись + notarization
- **macOS**: Developer ID signing + notarization
- Секреты в GitHub: `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_ID`, `APPLE_TEAM_ID`
- Использовать `fastlane` или нативный `codesign`/`notarytool`

### MikroTik

Поддержка двух форматов:
1. **Raw binary** - для ручной установки через SSH
2. **.npk package** - для установки через WinBox/WebFig

Архитектуры:
- `arm` (hAP, RB4011, etc.)
- `arm64` (CCR2004, RB5009)
- `mipsbe` (RB750, RB2011)

### Cisco

Статический linux-amd64 бинарник для:
- ISR 1000/4000 (IOS-XE)
- Catalyst 9000 (IOS-XE)
- Nexus 9000 (NX-OS)

### Workflow Architecture (Fault Tolerance)

Сборки разнесены по отдельным workflow файлам для отказоустойчивости:

```
.github/workflows/
├── _build-base.yml        # Reusable: базовая Go сборка
├── _build-library.yml     # Reusable: сборка библиотек (.so/.dll/.dylib/.a)
├── _build-package.yml     # Reusable: пакетирование (deb/rpm/ipk/npk)
├── _sign-apple.yml        # Reusable: Apple code signing + notarization
│
├── release.yml            # Orchestrator - вызывает все остальные
├── build-desktop.yml      # Windows, Linux, macOS, FreeBSD, OpenBSD
├── build-mobile.yml       # Android, iOS, Aurora OS
├── build-routers.yml      # OpenWRT, MikroTik, Cisco
├── build-gov.yml          # Astra Linux + placeholder для Elbrus
└── build-libraries.yml    # Библиотеки для всех платформ
```

**Принципы:**
1. Падение одного workflow не влияет на другие
2. Все категории собираются параллельно
3. Частичный релиз возможен (публикуется то, что собралось)
4. Можно перезапустить только упавший workflow
5. Elbrus активируется через `vars.ELBRUS_RUNNER_AVAILABLE`

## Supported Targets Summary

### Operating Systems

| OS | Architecture | Build Type | Status |
|----|--------------|------------|--------|
| Windows | amd64, arm64, 386 | Binary + DLL | ✅ Existing |
| Linux | amd64, arm64, arm, mips, riscv64, loong64, ppc64, s390x | Binary + .so | ✅ Existing |
| macOS | amd64, arm64 | Binary + .dylib + Framework | ✅ Existing + Framework |
| FreeBSD | amd64, arm64, arm, 386 | Binary | ✅ Existing |
| OpenBSD | amd64, arm64, arm, 386 | Binary | ✅ Existing |
| Android | arm64, amd64 | Binary + .so | ✅ Existing |
| iOS | arm64 | Framework | 🆕 New |
| Aurora OS | arm64 | Binary + RPM | 🆕 New |
| Astra Linux | amd64 | Binary + DEB | 🆕 New |

### Processors

| Processor | GOARCH | Status |
|-----------|--------|--------|
| Intel/AMD x86-64 | amd64 | ✅ Existing |
| Intel/AMD x86 | 386 | ✅ Existing |
| Apple Silicon | arm64 | ✅ Existing |
| ARM (v5-v8) | arm, arm64 | ✅ Existing |
| Elbrus | e2k | ⏸️ Deferred (no Go support) |
| Shakti (India RISC-V) | riscv64 | ✅ Existing (as riscv64) |
| Loongson | loong64 | ✅ Existing |
| MIPS | mips, mips64, mipsle, mips64le | ✅ Existing |

### Network Equipment

| Device | OS/Platform | Architecture | Package Format | Status |
|--------|-------------|--------------|----------------|--------|
| Cisco ISR/Catalyst/Nexus | IOS-XE/NX-OS | amd64 | Static binary | 🆕 New |
| MikroTik | RouterOS | arm, arm64, mipsbe | Binary + .npk | 🆕 New |
| OpenWRT | Linux (musl) | mips, arm, amd64 | .ipk | 🆕 New |

## References

- Existing workflow: `.github/workflows/release.yml`
- Go supported platforms: https://go.dev/doc/install/source#environment
- OpenWRT SDK: https://openwrt.org/docs/guide-developer/toolchain/using_the_sdk
- MikroTik development: https://help.mikrotik.com/docs/

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-28
- [x] Notes: Workflow architecture approved with fault-tolerant design (separate files per category)
