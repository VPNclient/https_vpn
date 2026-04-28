# Implementation Plan: Korean Ciphersuite (ARIA/SEED)

> Version: 1.0
> Status: DRAFT
> Last Updated: 2026-04-28
> Specifications: [02-specifications.md](./02-specifications.md)

## Summary

Реализация корейского криптографического провайдера по аналогии с `crypto/ua/`. Основные компоненты:
1. ARIA блочный шифр (128/192/256)
2. SEED блочный шифр (128)
3. TLS cipher suites
4. Provider для интеграции

## Task Breakdown

### Phase 1: Foundation

#### Task 1.1: Create directory structure
- **Description**: Создать структуру каталогов для KO провайдера
- **Files**:
  - `crypto/ko/` - Create directory
  - `crypto/ko/aria/` - Create directory
  - `crypto/ko/seed/` - Create directory
  - `crypto/ko/tls/` - Create directory
- **Dependencies**: None
- **Verification**: Directories exist
- **Complexity**: Low

#### Task 1.2: TLS constants
- **Description**: Определить TLS cipher suite константы
- **Files**:
  - `crypto/ko/tls/cipher_suites.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Constants compile
- **Complexity**: Low

### Phase 2: ARIA Implementation

#### Task 2.1: ARIA S-boxes and constants
- **Description**: Реализовать S-box таблицы и константы ARIA
- **Files**:
  - `crypto/ko/aria/consts.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Constants match RFC 5794
- **Complexity**: Medium

#### Task 2.2: ARIA core implementation
- **Description**: Реализовать ARIA encrypt/decrypt
- **Files**:
  - `crypto/ko/aria/aria.go` - Create
- **Dependencies**: Task 2.1
- **Verification**: RFC 5794 test vectors pass
- **Complexity**: High

#### Task 2.3: ARIA tests
- **Description**: Написать тесты с RFC векторами
- **Files**:
  - `crypto/ko/aria/aria_test.go` - Create
- **Dependencies**: Task 2.2
- **Verification**: `go test ./crypto/ko/aria/...` passes
- **Complexity**: Medium

### Phase 3: SEED Implementation

#### Task 3.1: SEED S-boxes and constants
- **Description**: Реализовать S-box таблицы и константы SEED
- **Files**:
  - `crypto/ko/seed/consts.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Constants match RFC 4269
- **Complexity**: Medium

#### Task 3.2: SEED core implementation
- **Description**: Реализовать SEED encrypt/decrypt (Feistel network)
- **Files**:
  - `crypto/ko/seed/seed.go` - Create
- **Dependencies**: Task 3.1
- **Verification**: RFC 4269 test vectors pass
- **Complexity**: High

#### Task 3.3: SEED tests
- **Description**: Написать тесты с RFC векторами
- **Files**:
  - `crypto/ko/seed/seed_test.go` - Create
- **Dependencies**: Task 3.2
- **Verification**: `go test ./crypto/ko/seed/...` passes
- **Complexity**: Medium

### Phase 4: Provider Integration

#### Task 4.1: KO Provider
- **Description**: Реализовать crypto.Provider для Кореи
- **Files**:
  - `crypto/ko/provider.go` - Create
- **Dependencies**: Task 2.2, Task 3.2
- **Verification**: Provider registers correctly
- **Complexity**: Medium

#### Task 4.2: Provider tests
- **Description**: Написать тесты провайдера
- **Files**:
  - `crypto/ko/provider_test.go` - Create
- **Dependencies**: Task 4.1
- **Verification**: `go test ./crypto/ko/...` passes
- **Complexity**: Medium

#### Task 4.3: Add IsKOCryptoSuite
- **Description**: Добавить функцию детекции KO cipher suites
- **Files**:
  - `crypto/provider.go` - Modify
- **Dependencies**: Task 1.2
- **Verification**: Detection works correctly
- **Complexity**: Low

### Phase 5: Documentation

#### Task 5.1: README
- **Description**: Написать документацию
- **Files**:
  - `crypto/ko/README.md` - Create
- **Dependencies**: Task 4.1
- **Verification**: Documentation is clear
- **Complexity**: Low

#### Task 5.2: Configuration example
- **Description**: Создать пример конфигурации
- **Files**:
  - `config.ko.json` - Create
- **Dependencies**: Task 4.1
- **Verification**: Config is valid JSON
- **Complexity**: Low

#### Task 5.3: Update main README
- **Description**: Обновить README.md проекта
- **Files**:
  - `README.md` - Modify
- **Dependencies**: Task 5.1
- **Verification**: KO provider documented
- **Complexity**: Low

## Dependency Graph

```
Task 1.1 ─┬─→ Task 1.2 ────────────────────────→ Task 4.3
          │
          ├─→ Task 2.1 ─→ Task 2.2 ─→ Task 2.3 ─┬─→ Task 4.1 ─→ Task 4.2
          │                                      │
          └─→ Task 3.1 ─→ Task 3.2 ─→ Task 3.3 ─┘      │
                                                        ↓
                                                   Task 5.1 ─→ Task 5.2 ─→ Task 5.3
```

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `crypto/ko/tls/cipher_suites.go` | Create | TLS constants |
| `crypto/ko/aria/consts.go` | Create | ARIA S-boxes |
| `crypto/ko/aria/aria.go` | Create | ARIA cipher |
| `crypto/ko/aria/aria_test.go` | Create | ARIA tests |
| `crypto/ko/seed/consts.go` | Create | SEED S-boxes |
| `crypto/ko/seed/seed.go` | Create | SEED cipher |
| `crypto/ko/seed/seed_test.go` | Create | SEED tests |
| `crypto/ko/provider.go` | Create | KO provider |
| `crypto/ko/provider_test.go` | Create | Provider tests |
| `crypto/ko/README.md` | Create | Documentation |
| `crypto/provider.go` | Modify | Add IsKOCryptoSuite |
| `config.ko.json` | Create | Example config |
| `README.md` | Modify | Add KO to docs |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| S-box implementation errors | Medium | High | Use RFC test vectors |
| Performance issues | Low | Medium | Profile and optimize |
| Go TLS incompatibility | Medium | Medium | Fallback to AES |

## Rollback Strategy

If implementation fails:
1. Delete `crypto/ko/` directory
2. Revert changes to `crypto/provider.go`
3. Revert changes to `README.md`

## Checkpoints

After each phase, verify:

- [ ] All tests pass
- [ ] No new warnings/errors
- [ ] Behavior matches specifications

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
