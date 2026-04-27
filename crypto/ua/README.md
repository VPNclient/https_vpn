# Українська постквантова криптографія (ДСТУ-ПК 2026)

Цей пакет реалізує криптографічний провайдер для української постквантової криптографії на базі концепції стандарту ДСТУ-ПК 2026.

## Алгоритми

### Хеш-функція: Купина-512 (Kupyna)
- **Стандарт**: ДСТУ 7564:2014 (модифікація для постквантової стійкості)
- **Розмір хешу**: 512 біт
- **Стійкість**: 256 біт (Category 5)

```go
import "github.com/nativemind/https-vpn/crypto/ua/kupyna"

hash := kupyna.Sum512([]byte("повідомлення"))
```

### Блочний шифр: Калина-512 (Kalyna)
- **Стандарт**: ДСТУ 7624:2014 (модифікація з 512-бітним ключем)
- **Розмір блоку**: 512 біт
- **Розмір ключа**: 512 біт
- **Режим**: GCM для аутентифікованого шифрування

```go
import "github.com/nativemind/https-vpn/crypto/ua/kalyna"

cipher, err := kalyna.NewCipher512(key)
```

### KEM: Мальва-1024 (Malva)
- **Основа**: Module-LWE (аналог ML-KEM/Kyber-1024)
- **Стійкість**: 256 біт (Category 5)
- **Гібридний режим**: X25519 + Malva для TLS

### Цифровий підпис: Сокіл-512 (Sokil)
- **Основа**: SIS/решітки (аналог ML-DSA/Dilithium-5)
- **Стійкість**: 256 біт (Category 5)
- **Гібридний режим**: ДСТУ 4145 + Сокіл

## TLS Cipher Suites

| ID | Назва | Опис |
|----|-------|------|
| 0xD001 | TLS_UA_KALYNA_512_GCM_KUPYNA_512 | Основний постквантовий suite |
| 0xD002 | TLS_UA_KALYNA_256_GCM_KUPYNA_256 | Полегшений варіант |

## Використання

### Конфігурація сервера

```json
{
  "tlsSettings": {
    "cipherSuites": "ua"
  }
}
```

### Програмне використання

```go
import (
    "github.com/nativemind/https-vpn/crypto"
    _ "github.com/nativemind/https-vpn/crypto/ua" // Реєстрація провайдера
)

provider, ok := crypto.Get("ua")
if ok {
    err := provider.ConfigureTLS(tlsConfig)
}
```

## Статус реалізації

| Компонент | Статус |
|-----------|--------|
| Купина-512 | Базова реалізація |
| Калина-512 | Базова реалізація |
| Мальва-1024 | В розробці |
| Сокіл-512 | В розробці |
| TLS провайдер | Реалізовано (fallback на AES) |
| Гібридний KEM | В розробці |

## Ліцензія

MIT License

## Посилання

- [ДСТУ 7564:2014](https://uk.wikipedia.org/wiki/Купина_(хеш-функція)) - Хеш-функція Купина
- [ДСТУ 7624:2014](https://uk.wikipedia.org/wiki/Калина_(шифр)) - Блочний шифр Калина
- [NIST FIPS 203](https://csrc.nist.gov/pubs/fips/203/final) - ML-KEM (референс для Мальви)
- [NIST FIPS 204](https://csrc.nist.gov/pubs/fips/204/final) - ML-DSA (референс для Сокола)
