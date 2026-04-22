# Implementation Plan: Chinese National Cryptography (SM Series)

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-04-22
> Specifications: [02-specifications.md](02-specifications.md)

## Summary

Реализация китайской криптографии в 4 фазах:
1. **Foundation** — SM3 (hash) и SM4 (cipher) без зависимостей
2. **Core** — SM2 (elliptic curve) использует SM3
3. **Advanced** — SM9 (identity-based) с BN256 pairings
4. **Integration** — TLS cipher suites и провайдер

## Task Breakdown

### Phase 1: Foundation (SM3 + SM4)

#### Task 1.1: Create package structure
- **Description**: Создать директории и базовые файлы
- **Files**:
  - `crypto/cn/sm3/sm3.go` - Create
  - `crypto/cn/sm4/sm4.go` - Create
  - `crypto/cn/sm2/sm2.go` - Create (stub)
  - `crypto/cn/sm9/sm9.go` - Create (stub)
  - `crypto/cn/tls/cipher_suites.go` - Create (stub)
  - `crypto/cn/provider.go` - Create (stub)
- **Dependencies**: None
- **Verification**: `go build ./crypto/cn/...` passes
- **Complexity**: Low

#### Task 1.2: Implement SM3 hash
- **Description**: SM3 хэш-функция по GB/T 32905
- **Files**:
  - `crypto/cn/sm3/sm3.go` - Implement
  - `crypto/cn/sm3/sm3_test.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Unit tests pass with official test vectors
- **Complexity**: Medium

**SM3 Implementation Details:**
```
- Block size: 64 bytes (512 bits)
- Digest size: 32 bytes (256 bits)
- Rounds: 64
- Operations: FF, GG, P0, P1 transformations
- Test vector: SM3("abc") = 66c7f0f4...
```

#### Task 1.3: Implement SM4 block cipher
- **Description**: SM4 блочный шифр по GB/T 32907
- **Files**:
  - `crypto/cn/sm4/sm4.go` - Implement
  - `crypto/cn/sm4/sm4_test.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Unit tests pass with official test vectors
- **Complexity**: Medium

**SM4 Implementation Details:**
```
- Block size: 16 bytes (128 bits)
- Key size: 16 bytes (128 bits)
- Rounds: 32
- S-box: 256-byte substitution table
- Key schedule: 32 round keys from master key
```

### Phase 2: Core (SM2)

#### Task 2.1: Implement SM2 curve
- **Description**: SM2-P256 эллиптическая кривая
- **Files**:
  - `crypto/cn/sm2/curve.go` - Create
  - `crypto/cn/sm2/curve_test.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Point operations work correctly
- **Complexity**: Medium

**SM2 Curve Parameters:**
```
P  = FFFFFFFE FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFF 00000000 FFFFFFFF FFFFFFFF
A  = FFFFFFFE FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFF 00000000 FFFFFFFF FFFFFFFC
B  = 28E9FA9E 9D9F5E34 4D5A9E4B CF6509A7 F39789F5 15AB8F92 DDBCBD41 4D940E93
N  = FFFFFFFE FFFFFFFF FFFFFFFF FFFFFFFF 7203DF6B 21C6052B 53BBF409 39D54123
Gx = 32C4AE2C 1F198119 5F990446 6A39C994 8FE30BBF F2660BE1 715A4589 334C74C7
Gy = BC3736A2 F4F6779C 59BDCEE3 6B692153 D0A9877C C62A4740 02DF32E5 2139F0A0
```

#### Task 2.2: Implement SM2 signatures
- **Description**: SM2 цифровые подписи
- **Files**:
  - `crypto/cn/sm2/sm2.go` - Implement Sign/Verify
  - `crypto/cn/sm2/sm2_test.go` - Create
- **Dependencies**: Task 2.1, Task 1.2 (SM3)
- **Verification**: Sign/Verify with test vectors
- **Complexity**: Medium

**SM2 Signature Algorithm:**
```
1. e = SM3(Z_A || M) where Z_A = SM3(ENTL || ID || a || b || Gx || Gy || Px || Py)
2. Generate random k
3. (x1, y1) = [k]G
4. r = (e + x1) mod n
5. s = ((1 + d)^-1 * (k - r*d)) mod n
6. Signature = (r, s)
```

#### Task 2.3: Implement SM2 encryption
- **Description**: SM2 асимметричное шифрование
- **Files**:
  - `crypto/cn/sm2/sm2.go` - Implement Encrypt/Decrypt
  - `crypto/cn/sm2/sm2_test.go` - Extend
- **Dependencies**: Task 2.1, Task 1.2 (SM3)
- **Verification**: Encrypt/Decrypt roundtrip
- **Complexity**: Medium

**SM2 Encryption:**
```
Encrypt:
1. k = random
2. C1 = [k]G
3. (x2, y2) = [k]P_B
4. t = KDF(x2 || y2, klen)
5. C2 = M XOR t
6. C3 = SM3(x2 || M || y2)
7. Output: C1 || C3 || C2 (or C1 || C2 || C3)
```

### Phase 3: Advanced (SM9)

#### Task 3.1: Implement BN256 pairing curve
- **Description**: BN256 кривая для SM9 pairings
- **Files**:
  - `crypto/cn/sm9/bn256.go` - Create
  - `crypto/cn/sm9/bn256_test.go` - Create
- **Dependencies**: Task 1.1
- **Verification**: Pairing properties hold
- **Complexity**: High

**BN256 Parameters (SM9):**
```
t = 0x600000000058F98A
p = 36*t^4 + 36*t^3 + 24*t^2 + 6*t + 1
n = 36*t^4 + 36*t^3 + 18*t^2 + 6*t + 1
```

#### Task 3.2: Implement SM9 signatures
- **Description**: SM9 identity-based подписи
- **Files**:
  - `crypto/cn/sm9/sm9.go` - Implement
  - `crypto/cn/sm9/sm9_test.go` - Create
- **Dependencies**: Task 3.1, Task 1.2 (SM3)
- **Verification**: Sign/Verify with test vectors
- **Complexity**: High

#### Task 3.3: Implement SM9 encryption
- **Description**: SM9 identity-based шифрование
- **Files**:
  - `crypto/cn/sm9/sm9.go` - Extend
  - `crypto/cn/sm9/sm9_test.go` - Extend
- **Dependencies**: Task 3.1, Task 1.2, Task 1.3
- **Verification**: Encrypt/Decrypt roundtrip
- **Complexity**: High

### Phase 4: Integration

#### Task 4.1: Define TLS cipher suites
- **Description**: TLS_SM4_GCM_SM3, TLS_SM4_CCM_SM3
- **Files**:
  - `crypto/cn/tls/cipher_suites.go` - Implement
- **Dependencies**: Task 1.2, Task 1.3
- **Verification**: Constants match RFC 8998
- **Complexity**: Low

#### Task 4.2: Implement SM4-GCM mode
- **Description**: GCM mode для SM4
- **Files**:
  - `crypto/cn/sm4/gcm.go` - Create
  - `crypto/cn/sm4/gcm_test.go` - Create
- **Dependencies**: Task 1.3
- **Verification**: GCM encrypt/decrypt works
- **Complexity**: Medium

#### Task 4.3: Implement SM4-CCM mode
- **Description**: CCM mode для SM4
- **Files**:
  - `crypto/cn/sm4/ccm.go` - Create
  - `crypto/cn/sm4/ccm_test.go` - Create
- **Dependencies**: Task 1.3
- **Verification**: CCM encrypt/decrypt works
- **Complexity**: Medium

#### Task 4.4: Implement CN provider
- **Description**: Регистрация провайдера "cn"
- **Files**:
  - `crypto/cn/provider.go` - Implement
  - `crypto/cn/provider_test.go` - Create
- **Dependencies**: Task 4.1, Task 4.2, Task 4.3
- **Verification**: `crypto.Get("cn")` returns valid provider
- **Complexity**: Low

#### Task 4.5: Integration testing
- **Description**: End-to-end тесты
- **Files**:
  - `crypto/cn/integration_test.go` - Create
- **Dependencies**: All previous tasks
- **Verification**: TLS handshake with SM cipher suite works
- **Complexity**: Medium

## Dependency Graph

```
Phase 1 (Foundation):
  Task 1.1 ─┬─→ Task 1.2 (SM3)
            └─→ Task 1.3 (SM4)

Phase 2 (Core):
  Task 1.1 ──→ Task 2.1 (curve) ─┬─→ Task 2.2 (sign)
  Task 1.2 ─────────────────────┘└─→ Task 2.3 (encrypt)

Phase 3 (Advanced):
  Task 1.1 ──→ Task 3.1 (BN256) ─┬─→ Task 3.2 (SM9 sign)
  Task 1.2 ─────────────────────┘└─→ Task 3.3 (SM9 encrypt)
  Task 1.3 ────────────────────────┘

Phase 4 (Integration):
  Task 1.2 ─┬─→ Task 4.1 (cipher suites)
  Task 1.3 ─┼─→ Task 4.2 (GCM)
            └─→ Task 4.3 (CCM)

  Task 4.1 ─┬
  Task 4.2 ─┼─→ Task 4.4 (provider) ──→ Task 4.5 (integration)
  Task 4.3 ─┘
```

## File Change Summary

| File | Action | Reason |
|------|--------|--------|
| `crypto/cn/sm3/sm3.go` | Create | SM3 hash implementation |
| `crypto/cn/sm3/sm3_test.go` | Create | SM3 unit tests |
| `crypto/cn/sm4/sm4.go` | Create | SM4 cipher implementation |
| `crypto/cn/sm4/sm4_test.go` | Create | SM4 unit tests |
| `crypto/cn/sm4/gcm.go` | Create | GCM mode for SM4 |
| `crypto/cn/sm4/ccm.go` | Create | CCM mode for SM4 |
| `crypto/cn/sm2/curve.go` | Create | SM2-P256 curve |
| `crypto/cn/sm2/sm2.go` | Create | SM2 sign/verify/encrypt/decrypt |
| `crypto/cn/sm2/sm2_test.go` | Create | SM2 unit tests |
| `crypto/cn/sm9/bn256.go` | Create | BN256 pairing curve |
| `crypto/cn/sm9/sm9.go` | Create | SM9 identity-based crypto |
| `crypto/cn/sm9/sm9_test.go` | Create | SM9 unit tests |
| `crypto/cn/tls/cipher_suites.go` | Create | TLS constants |
| `crypto/cn/provider.go` | Create | Provider registration |
| `crypto/cn/provider_test.go` | Create | Provider tests |
| `crypto/cn/integration_test.go` | Create | E2E tests |

**Total: 16 new files**

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| BN256 pairing bugs | Medium | High | Use reference implementation, extensive testing |
| SM9 complexity | High | Medium | Implement last, can be deferred |
| Performance issues | Low | Medium | Optimize hot paths after correctness |
| Standard compliance | Medium | High | Use official test vectors |

## Rollback Strategy

1. Each phase is independent — can roll back to previous phase
2. Provider registration is last — easy to disable
3. All new files — no existing code modified

## Checkpoints

### After Phase 1
- [ ] SM3 tests pass with official vectors
- [ ] SM4 tests pass with official vectors
- [ ] `go build ./crypto/cn/...` succeeds

### After Phase 2
- [ ] SM2 sign/verify tests pass
- [ ] SM2 encrypt/decrypt tests pass
- [ ] Curve operations verified

### After Phase 3
- [ ] BN256 pairing tests pass
- [ ] SM9 sign/verify tests pass
- [ ] SM9 encrypt/decrypt tests pass

### After Phase 4
- [ ] Provider registers successfully
- [ ] TLS handshake with SM works
- [ ] All tests pass

## Open Implementation Questions

- [x] Все решены

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
