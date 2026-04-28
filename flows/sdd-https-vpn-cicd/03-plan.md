# Implementation Plan: HTTPS VPN CI/CD Pipeline

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-04-28
> Specifications: [02-specifications.md](02-specifications.md)

## Summary

Реорганизация CI/CD pipeline из монолитного `release.yml` в модульную структуру с изолированными workflow файлами. Реализация в 5 фаз: reusable workflows, миграция desktop, новые платформы (mobile/routers/gov), библиотеки, Apple signing.

## Task Breakdown

### Phase 1: Foundation (Reusable Workflows)

#### Task 1.1: Create VERSION file
- **Description**: Создать файл VERSION с семантической версией
- **Files**:
  - `VERSION` - Create
- **Dependencies**: None
- **Verification**: `cat VERSION` показывает `1.0.0`
- **Complexity**: Low

#### Task 1.2: Create _build-base.yml
- **Description**: Reusable workflow для базовой Go сборки
- **Files**:
  - `.github/workflows/_build-base.yml` - Create
- **Dependencies**: Task 1.1
- **Verification**: Можно вызвать из другого workflow
- **Complexity**: Medium

#### Task 1.3: Create _build-library.yml
- **Description**: Reusable workflow для сборки библиотек (.so/.dll/.dylib/.a)
- **Files**:
  - `.github/workflows/_build-library.yml` - Create
- **Dependencies**: Task 1.2
- **Verification**: Создаёт .so и .a файлы
- **Complexity**: Medium

#### Task 1.4: Create _build-package.yml
- **Description**: Reusable workflow для пакетирования (deb/rpm)
- **Files**:
  - `.github/workflows/_build-package.yml` - Create
- **Dependencies**: Task 1.2
- **Verification**: Создаёт .deb пакет
- **Complexity**: Medium

#### Task 1.5: Create _sign-apple.yml
- **Description**: Reusable workflow для Apple code signing
- **Files**:
  - `.github/workflows/_sign-apple.yml` - Create
- **Dependencies**: None
- **Verification**: Подписывает тестовый бинарник (требует secrets)
- **Complexity**: High

### Phase 2: Desktop Migration

#### Task 2.1: Create build-desktop.yml
- **Description**: Workflow для Windows/Linux/macOS/BSD сборок
- **Files**:
  - `.github/workflows/build-desktop.yml` - Create
- **Dependencies**: Task 1.2, Task 1.5
- **Verification**: Все существующие desktop платформы собираются
- **Complexity**: Medium

#### Task 2.2: Update friendly-filenames.json
- **Description**: Добавить новые платформы в маппинг имён
- **Files**:
  - `.github/build/friendly-filenames.json` - Modify
- **Dependencies**: None
- **Verification**: JSON валиден, все платформы присутствуют
- **Complexity**: Low

#### Task 2.3: Convert release.yml to orchestrator
- **Description**: Преобразовать release.yml в orchestrator, вызывающий category workflows
- **Files**:
  - `.github/workflows/release.yml` - Modify
- **Dependencies**: Task 2.1
- **Verification**: Release workflow вызывает build-desktop
- **Complexity**: Medium

### Phase 3: New Platforms

#### Task 3.1: Create build-mobile.yml
- **Description**: Workflow для Android/iOS/Aurora OS
- **Files**:
  - `.github/workflows/build-mobile.yml` - Create
- **Dependencies**: Task 1.2, Task 1.5
- **Verification**: Android и iOS сборки успешны
- **Complexity**: High

#### Task 3.2: Create build-routers.yml
- **Description**: Workflow для OpenWRT/MikroTik(binary)/Cisco
- **Files**:
  - `.github/workflows/build-routers.yml` - Create
- **Dependencies**: Task 1.2
- **Verification**: Статические бинарники для роутеров создаются
- **Complexity**: Medium

#### Task 3.3: Create build-mikrotik-npk.yml
- **Description**: Отдельный workflow для MikroTik .npk пакетов
- **Files**:
  - `.github/workflows/build-mikrotik-npk.yml` - Create
- **Dependencies**: Task 3.2
- **Verification**: .npk пакеты создаются (требует RouterOS SDK)
- **Complexity**: High

#### Task 3.4: Create build-gov.yml
- **Description**: Workflow для Astra Linux + placeholder для Elbrus
- **Files**:
  - `.github/workflows/build-gov.yml` - Create
- **Dependencies**: Task 1.2, Task 1.4
- **Verification**: Astra .deb пакет создаётся
- **Complexity**: Medium

### Phase 4: Libraries

#### Task 4.1: Create build-libraries.yml
- **Description**: Workflow для сборки библиотек всех платформ
- **Files**:
  - `.github/workflows/build-libraries.yml` - Create
- **Dependencies**: Task 1.3
- **Verification**: .so/.dll/.dylib/.a создаются для всех платформ
- **Complexity**: Medium

### Phase 5: Integration & Testing

#### Task 5.1: Update release.yml orchestrator
- **Description**: Добавить все category workflows в orchestrator
- **Files**:
  - `.github/workflows/release.yml` - Modify
- **Dependencies**: Task 3.1, Task 3.2, Task 3.3, Task 3.4, Task 4.1
- **Verification**: Все workflows запускаются параллельно
- **Complexity**: Low

#### Task 5.2: Add publish job
- **Description**: Финальный job для сбора артефактов и публикации
- **Files**:
  - `.github/workflows/release.yml` - Modify
- **Dependencies**: Task 5.1
- **Verification**: Артефакты загружаются в GitHub Release
- **Complexity**: Medium

#### Task 5.3: Test full release cycle
- **Description**: Создать тестовый release и проверить все артефакты
- **Files**: None (manual testing)
- **Dependencies**: Task 5.2
- **Verification**: Все ожидаемые артефакты присутствуют в release
- **Complexity**: Low

## Dependency Graph

```
Task 1.1 (VERSION)
    │
    ▼
Task 1.2 (_build-base) ──────────────────────────────────┐
    │                                                     │
    ├──────────────┬──────────────┐                      │
    ▼              ▼              ▼                      │
Task 1.3       Task 1.4       Task 1.5                   │
(_build-lib)   (_build-pkg)   (_sign-apple)              │
    │              │              │                      │
    │              │              ├──────────────────────┤
    │              │              │                      │
    ▼              ▼              ▼                      │
Task 4.1       Task 3.4       Task 2.1 ◄─────────────────┘
(build-libs)   (build-gov)    (build-desktop)
                                  │
                                  ▼
                              Task 2.2 (filenames.json)
                                  │
                                  ▼
                              Task 2.3 (release.yml v1)
                                  │
    ┌─────────────────────────────┼─────────────────────────────┐
    │                             │                             │
    ▼                             ▼                             ▼
Task 3.1                      Task 3.2                      Task 3.4
(build-mobile)                (build-routers)               (build-gov)
    │                             │
    │                             ▼
    │                         Task 3.3
    │                         (build-mikrotik-npk)
    │                             │
    └─────────────────────────────┼─────────────────────────────┘
                                  │
                                  ▼
                              Task 5.1 (release.yml v2)
                                  │
                                  ▼
                              Task 5.2 (publish job)
                                  │
                                  ▼
                              Task 5.3 (test release)
```

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `VERSION` | Create | Хранение семантической версии |
| `.github/workflows/_build-base.yml` | Create | Reusable Go build |
| `.github/workflows/_build-library.yml` | Create | Reusable library build |
| `.github/workflows/_build-package.yml` | Create | Reusable packaging |
| `.github/workflows/_sign-apple.yml` | Create | Apple signing |
| `.github/workflows/build-desktop.yml` | Create | Desktop platforms |
| `.github/workflows/build-mobile.yml` | Create | Mobile platforms |
| `.github/workflows/build-routers.yml` | Create | Router platforms |
| `.github/workflows/build-mikrotik-npk.yml` | Create | MikroTik .npk |
| `.github/workflows/build-gov.yml` | Create | Government platforms |
| `.github/workflows/build-libraries.yml` | Create | Library builds |
| `.github/workflows/release.yml` | Modify | Convert to orchestrator |
| `.github/build/friendly-filenames.json` | Modify | Add new platforms |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Apple signing secrets invalid | Medium | High | Test signing early in Phase 1 |
| MikroTik SDK unavailable | High | Low | .npk в отдельном workflow, binary всегда работает |
| iOS build requires paid account | Low | Medium | Уже есть Apple Developer |
| Cross-compilation fails for exotic arch | Medium | Low | CGO_ENABLED=0 для большинства |
| Workflow timeout (>60 min) | Low | Medium | Параллельные jobs, кэширование |

## Rollback Strategy

Если имплементация провалится:

1. Revert commits to `.github/workflows/`
2. Восстановить оригинальный `release.yml` из git history
3. Удалить `VERSION` файл если не нужен
4. Все существующие сборки продолжат работать

## Checkpoints

### После Phase 1:
- [ ] Все reusable workflows синтаксически корректны
- [ ] `_build-base.yml` успешно собирает linux-amd64

### После Phase 2:
- [ ] `build-desktop.yml` собирает все существующие платформы
- [ ] `release.yml` корректно вызывает build-desktop
- [ ] Артефакты имеют правильные имена

### После Phase 3:
- [ ] iOS framework создаётся (unsigned OK для теста)
- [ ] Router бинарники создаются
- [ ] Astra .deb пакет создаётся

### После Phase 4:
- [ ] Библиотеки создаются для всех платформ
- [ ] .so/.dll/.dylib линкуются корректно

### После Phase 5:
- [ ] Полный release cycle работает
- [ ] Partial failure не блокирует другие сборки
- [ ] Все артефакты в GitHub Release

## Open Implementation Questions

- [ ] Нужен ли caching для Go modules между workflows?
- [ ] Как обрабатывать version bump (manual или automatic)?
- [ ] Нужен ли отдельный workflow для nightly builds?

---

## Approval

- [x] Reviewed by: User
- [x] Approved on: 2026-04-28
- [x] Notes: Proceed with implementation
