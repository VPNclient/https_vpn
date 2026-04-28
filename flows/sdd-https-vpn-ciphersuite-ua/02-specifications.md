# Специфікації: Українська постквантова криптографія (ДСТУ-ПК 2026)

> Версія: 1.0
> Статус: APPROVED
> Останнє оновлення: 2026-04-27
> Вимоги: [01-requirements.md](./01-requirements.md)

## Огляд

Ця специфікація описує реалізацію UA криптографічного провайдера для HTTPS VPN сервера. Провайдер забезпечує підтримку постквантової криптографії на базі українських національних стандартів з гібридним режимом для забезпечення сумісності та безпеки в перехідний період.

## Задіяні системи

| Система | Вплив | Примітки |
|---------|-------|----------|
| `crypto/ua/` | Створити | Новий каталог для UA провайдера |
| `crypto/ua/provider.go` | Створити | Реалізація інтерфейсу Provider |
| `crypto/ua/malva/` | Створити | KEM на базі Module-LWE |
| `crypto/ua/sokil/` | Створити | Цифровий підпис на базі SIS |
| `crypto/ua/kalyna/` | Створити | Блочний шифр 512-біт з GCM |
| `crypto/ua/kupyna/` | Створити | Хеш-функція 512-біт |
| `crypto/ua/tls/` | Створити | TLS cipher suites та константи |
| `crypto/provider.go` | Модифікувати | Додати IsUACryptoSuite() |
| `crypto/certstore.go` | Модифікувати | Додати детекцію UA ключів |

## Архітектура

### Діаграма компонентів

```
┌─────────────────────────────────────────────────────────────┐
│                        TLS Server                            │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    CertificateStore                      ││
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌──────────┐││
│  │  │ US Certs  │ │ CN Certs  │ │ RU Certs  │ │ UA Certs │││
│  │  └───────────┘ └───────────┘ └───────────┘ └──────────┘││
│  └─────────────────────────────────────────────────────────┘│
│                              │                               │
│                              ▼                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Provider Registry                      ││
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐       ││
│  │  │ US  │ │ CN  │ │ RU  │ │ UK  │ │ FR  │ │ UA  │ ← NEW ││
│  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      UA Provider                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   crypto/ua/                            │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │ │
│  │  │  malva/  │  │  sokil/  │  │ kalyna/  │  │ kupyna/ │ │ │
│  │  │  (KEM)   │  │ (Підпис) │  │ (Шифр)   │  │  (Хеш)  │ │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │ │
│  │                      │                                  │ │
│  │                      ▼                                  │ │
│  │              ┌──────────────┐                           │ │
│  │              │    tls/      │                           │ │
│  │              │ cipher_suites│                           │ │
│  │              └──────────────┘                           │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Потік даних (TLS 1.3 Handshake)

```
Клієнт                                              Сервер
   │                                                    │
   │ ClientHello                                        │
   │  - supported_groups: [X25519_Malva, X25519]        │
   │  - cipher_suites: [TLS_UA_KALYNA_512_GCM_KUPYNA]   │
   │  - key_share: X25519 + Malva публічний ключ       │
   │ ─────────────────────────────────────────────────> │
   │                                                    │
   │                                       ServerHello │
   │               - selected: X25519_Malva            │
   │               - selected: TLS_UA_KALYNA_...       │
   │               - key_share: X25519 + Malva відп.   │
   │                                                    │
   │                            EncryptedExtensions    │
   │                              (Kalyna-512-GCM)     │
   │                                                    │
   │                                    Certificate    │
   │                              (Sokil + ДСТУ 4145)  │
   │                                                    │
   │                            CertificateVerify      │
   │                              (Sokil підпис)       │
   │ <───────────────────────────────────────────────── │
   │                                                    │
   │ Finished (Kupyna-512 PRF)                         │
   │ ─────────────────────────────────────────────────> │
   │                                                    │
   │                                 Finished          │
   │ <───────────────────────────────────────────────── │
   │                                                    │
   │ ═══════════════════════════════════════════════════│
   │         Захищений канал (Kalyna-512-GCM)          │
   │ ═══════════════════════════════════════════════════│
```

## Інтерфейси

### Новий інтерфейс провайдера

```go
// crypto/ua/provider.go
package ua

import (
    "crypto/tls"
    "github.com/example/https_vpn/crypto"
)

// Provider реалізує crypto.Provider для української криптографії
type Provider struct{}

func init() {
    crypto.Registry.Register("ua", &Provider{})
}

func (p *Provider) Name() string {
    return "ua"
}

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
    // Налаштування TLS 1.3 з українськими алгоритмами
    cfg.MinVersion = tls.VersionTLS13
    cfg.MaxVersion = tls.VersionTLS13
    cfg.CipherSuites = p.SupportedCipherSuites()
    cfg.CurvePreferences = []tls.CurveID{
        CurveX25519Malva, // Гібридний: X25519 + Malva-1024
        tls.X25519,       // Fallback для сумісності
    }
    return nil
}

func (p *Provider) SupportedCipherSuites() []uint16 {
    return []uint16{
        TLS_UA_KALYNA_512_GCM_KUPYNA_512,
        tls.TLS_AES_256_GCM_SHA384, // Fallback
    }
}
```

### Модифікація існуючих інтерфейсів

```go
// crypto/provider.go - додати функцію детекції
func IsUACryptoSuite(suite uint16) bool {
    return suite >= 0xUA00 && suite <= 0xUAFF
}
```

## Моделі даних

### Нові типи

```go
// crypto/ua/tls/cipher_suites.go
package tls

// Cipher Suites для українського провайдера
const (
    // TLS_UA_KALYNA_512_GCM_KUPYNA_512 використовує:
    // - AEAD: Kalyna-512-GCM (512-біт ключ)
    // - Hash: Kupyna-512
    // - KEM: X25519 + Malva-1024 (гібрид)
    TLS_UA_KALYNA_512_GCM_KUPYNA_512 uint16 = 0xUA01

    // Резервні ID для майбутнього розширення
    TLS_UA_KALYNA_256_GCM_KUPYNA_256 uint16 = 0xUA02
)

// Ідентифікатори кривих
const (
    // CurveX25519Malva - гібридна група для KEM
    CurveX25519Malva tls.CurveID = 0x6D01

    // CurveDSTU4145_512 - українська еліптична крива 512 біт
    CurveDSTU4145_512 tls.CurveID = 0x6D02
)

// Ідентифікатори підписів
const (
    SignatureSokil512 uint16 = 0x0720
    SignatureDSTU4145 uint16 = 0x0721
)
```

```go
// crypto/ua/malva/malva.go
package malva

// Параметри Мальви (аналог Kyber-1024, Category 5)
const (
    // PublicKeySize - розмір відкритого ключа
    PublicKeySize = 1568

    // PrivateKeySize - розмір закритого ключа
    PrivateKeySize = 3168

    // CiphertextSize - розмір шифротексту
    CiphertextSize = 1568

    // SharedSecretSize - розмір спільного секрету
    SharedSecretSize = 32

    // K - параметр модуля
    K = 4

    // N - розмір поліному
    N = 256

    // Q - модуль
    Q = 3329
)

// PublicKey представляє відкритий ключ Мальви
type PublicKey struct {
    pk [PublicKeySize]byte
}

// PrivateKey представляє закритий ключ Мальви
type PrivateKey struct {
    sk [PrivateKeySize]byte
}

// Ciphertext представляє шифротекст KEM
type Ciphertext struct {
    ct [CiphertextSize]byte
}
```

```go
// crypto/ua/sokil/sokil.go
package sokil

// Параметри Сокола (аналог Dilithium-5, Category 5)
const (
    // PublicKeySize - розмір відкритого ключа
    PublicKeySize = 2592

    // PrivateKeySize - розмір закритого ключа
    PrivateKeySize = 4880

    // SignatureSize - розмір підпису
    SignatureSize = 4627

    // K та L - параметри решітки
    K = 8
    L = 7

    // Q - модуль
    Q = 8380417
)

// PublicKey представляє відкритий ключ Сокола
type PublicKey struct {
    pk [PublicKeySize]byte
}

// PrivateKey представляє закритий ключ Сокола
type PrivateKey struct {
    sk [PrivateKeySize]byte
}

// Signature представляє цифровий підпис
type Signature struct {
    sig [SignatureSize]byte
}
```

```go
// crypto/ua/kalyna/kalyna.go
package kalyna

import (
    "crypto/cipher"
)

// Параметри Калини-ПК (512-біт ключ для квантової стійкості)
const (
    // BlockSize - розмір блоку в байтах
    BlockSize = 64 // 512 біт

    // KeySize - розмір ключа в байтах
    KeySize = 64 // 512 біт

    // Rounds - кількість раундів
    Rounds = 18
)

// Cipher реалізує cipher.Block для Калини-512
type Cipher struct {
    roundKeys [Rounds + 1][8]uint64
}

// NewCipher створює новий шифр Калина-512
func NewCipher(key []byte) (cipher.Block, error) {
    // Реалізація
}

// NewGCM створює GCM режим для Калини
func NewGCM(c cipher.Block) (cipher.AEAD, error) {
    // Реалізація
}
```

```go
// crypto/ua/kupyna/kupyna.go
package kupyna

import (
    "hash"
)

// Параметри Купини-ПК (512-біт для квантової стійкості)
const (
    // Size - розмір хешу в байтах
    Size = 64 // 512 біт

    // BlockSize - розмір блоку в байтах
    BlockSize = 64 // 512 біт
)

// Digest реалізує hash.Hash для Купина-512
type Digest struct {
    state [8]uint64
    buf   [BlockSize]byte
    len   uint64
}

// New створює новий хеш Купина-512
func New() hash.Hash {
    // Реалізація
}

// Sum512 обчислює хеш даних
func Sum512(data []byte) [Size]byte {
    // Реалізація
}
```

### Зміни схеми

Немає змін у постійних даних. Конфігурація зберігається у JSON.

## Поведінкові специфікації

### Щасливий шлях

1. Адміністратор налаштовує `config.json` з `cipherSuites: "ua"`
2. Сервер завантажує конфігурацію та реєструє UA провайдер
3. Клієнт підключається з підтримкою UA cipher suites
4. Handshake використовує гібридний KEM (X25519 + Malva)
5. Встановлюється захищений канал з Kalyna-512-GCM
6. Дані передаються з аутентифікованим шифруванням

### Граничні випадки

| Випадок | Тригер | Очікувана поведінка |
|---------|--------|---------------------|
| Клієнт без підтримки UA | Клієнт пропонує лише стандартні suites | Fallback на TLS_AES_256_GCM_SHA384 |
| Недійсний ключ Malva | Спроба KEM з пошкодженим ключем | Handshake завершується з помилкою |
| Занадто короткий ключ Sokil | Підпис з недійсним ключем | Верифікація повертає false |
| GCM nonce reuse | Повторне використання nonce | AEAD відмовляє в операції |

### Обробка помилок

| Помилка | Причина | Відповідь |
|---------|---------|-----------|
| `ErrInvalidKeySize` | Ключ не відповідає очікуваному розміру | Повернути помилку, не продовжувати |
| `ErrDecapsulationFailed` | KEM розшифрування не вдалося | Закрити з'єднання з помилкою handshake |
| `ErrVerificationFailed` | Підпис не пройшов перевірку | Закрити з'єднання, не довіряти серверу |
| `ErrInvalidCiphertext` | AEAD розшифрування не вдалося | Закрити з'єднання з bad_record_mac |

## Залежності

### Потребує

- Існуюча інфраструктура провайдерів (`crypto/provider.go`)
- CertificateStore для автовибору сертифікатів
- Go 1.21+ з підтримкою crypto/tls

### Блокує

- Інтеграція з АЦСК для постквантових сертифікатів (майбутнє)
- Клієнтські бібліотеки для інших платформ

## Точки інтеграції

### Зовнішні системи

- Немає зовнішніх залежностей. Всі алгоритми реалізуються локально.

### Внутрішні системи

| Система | Тип інтеграції |
|---------|----------------|
| `crypto/provider.go` | Реєстрація провайдера |
| `crypto/certstore.go` | Детекція UA ключів |
| `config/` | Парсинг конфігурації |
| `server/tls.go` | Налаштування TLS |

## Стратегія тестування

### Юніт-тести

- [ ] `crypto/ua/malva/malva_test.go` - KEM операції (keygen, encaps, decaps)
- [ ] `crypto/ua/sokil/sokil_test.go` - Підпис операції (keygen, sign, verify)
- [ ] `crypto/ua/kalyna/kalyna_test.go` - Блочний шифр та GCM режим
- [ ] `crypto/ua/kupyna/kupyna_test.go` - Хеш-функція з тестовими векторами
- [ ] `crypto/ua/provider_test.go` - Реєстрація та конфігурація TLS

### Інтеграційні тести

- [ ] Повний TLS handshake з UA cipher suite
- [ ] Fallback на стандартні suites при несумісності
- [ ] Автовибір сертифіката за типом ключа

### Ручна верифікація

- [ ] Підключення клієнта з підтримкою UA криптографії
- [ ] Перегляд параметрів з'єднання в логах
- [ ] Перевірка продуктивності handshake

## Міграція / Впровадження

1. **Фаза 1**: Реалізація базових криптографічних примітивів
2. **Фаза 2**: Інтеграція з TLS провайдером
3. **Фаза 3**: Тестування з реальними клієнтами
4. **Фаза 4**: Документація та прикладні конфігурації

Зворотна сумісність забезпечується fallback на стандартні cipher suites.

## Відкриті питання дизайну

- [ ] Чи використовувати бібліотеку pqcrypto для ML-KEM/ML-DSA чи власну реалізацію?
- [ ] Формат гібридного ключа X25519 + Malva (конкатенація чи ASN.1?)
- [ ] OID для нових алгоритмів (узгодити з ДСТУ)

---

## Затвердження

- [x] Переглянуто: 2026-04-27
- [x] Затверджено: 2026-04-27
- [x] Примітки: Затверджено користувачем
