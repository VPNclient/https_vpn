# Requirements: HTTPS VPN

> Version: 1.1
> Status: DRAFT
> Last Updated: 2026-03-10

## Problem Statement

Существующие VPN-решения имеют несколько критических проблем:

1. **Несоответствие национальным криптостандартам** - большинство VPN используют только западные криптоалгоритмы (AES, ChaCha20), что делает невозможным их сертификацию и легальное использование в странах с собственными криптостандартами (Россия - ГОСТ, Китай - SM2/SM3/SM4, и др.)

2. **Легкость обнаружения** - специализированные VPN-протоколы имеют уникальные сигнатуры, что позволяет DPI-системам идентифицировать и блокировать VPN-трафик

3. **Сложность сертификации** - объемные кодовые базы требуют дорогостоящей и длительной сертификации

4. **Сложность интеграции** - новые решения требуют переработки существующей инфраструктуры

## User Stories

### Primary

**As a** организация, требующая соответствия национальным криптостандартам
**I want** VPN-решение с поддержкой сертифицированных криптоалгоритмов
**So that** я могу легально использовать VPN в соответствии с требованиями регуляторов

**As a** оператор VPN-инфраструктуры
**I want** drop-in замену для существующего xray-сервера
**So that** я могу использовать существующие панели управления (3x-ui, marzban) без изменений

**As a** пользователь VPN
**I want** чтобы мой VPN-трафик выглядел как обычный HTTPS-трафик
**So that** он не блокировался системами DPI и не привлекал внимания

### Secondary

**As a** разработчик/интегратор
**I want** минимальный объем нового кода
**So that** облегчить процесс сертификации и аудита

**As a** security engineer
**I want** минимальную поверхность атаки
**So that** снизить риски уязвимостей в VPN-решении

## National Cryptography Standards

Каждый криптостандарт реализуется в отдельном TDD flow для модульности и независимой сертификации.

| Страна | ISO | Crypto Org | PKI / Signature | Hash | Symmetric | Browser-compatible transport | TDD Flow |
|--------|-----|------------|-----------------|------|-----------|------------------------------|----------|
| 🇺🇸 США | US | National Institute of Standards and Technology | ECDSA / EdDSA | SHA-2 / SHA-3 | AES | TLS 1.3 + HTTP/2/HTTP/3 с параметрами как у Chrome/Firefox (ALPN, cipher order, extensions) | [tdd-crypto-us](../tdd-crypto-us/) |
| 🇨🇳 Китай | CN | State Cryptography Administration | SM2 | SM3 | SM4 | GMSSL (вариант TLS) с браузероподобным handshake. Используется в китайских браузерах | [tdd-crypto-cn](../tdd-crypto-cn/) |
| 🇷🇺 Россия | RU | Federal Security Service | GOST R 34.10 | Streebog | Kuznyechik | TLS с GOST cipher suites внутри стандартного HTTPS транспорта | [tdd-crypto-ru](../tdd-crypto-ru/) |
| 🇰🇷 Южная Корея | KR | Korea Internet & Security Agency | KCDSA | HAS-160 | SEED | TLS стек с SEED cipher suites, совместимый с HTTPS | [tdd-crypto-kr](../tdd-crypto-kr/) |
| 🇯🇵 Япония | JP | CRYPTREC | ECDSA профили | SHA-2 | Camellia | TLS cipher suites Camellia + стандартный HTTPS стек | [tdd-crypto-jp](../tdd-crypto-jp/) |
| 🇮🇳 Индия | IN | Standardisation Testing and Quality Certification Directorate | ECSDSA | SHA-2 | AES | TLS-транспорт с ECC профилями, совпадающий с браузерными handshake | [tdd-crypto-in](../tdd-crypto-in/) |
| 🇪🇺 ЕС | EU | European Telecommunications Standards Institute | Brainpool ECC | SHA-2 | AES | TLS с Brainpool curves, стандартный HTTPS transport | [tdd-crypto-eu](../tdd-crypto-eu/) |
| 🇫🇷 Франция | FR | Agence nationale de la sécurité des systèmes d'information | ECDSA | SHA-256 | AES | Рекомендуется стандартный TLS без нестандартных протоколов | [tdd-crypto-fr](../tdd-crypto-fr/) |
| 🇬🇧 Великобритания | GB | National Cyber Security Centre | ECDSA | SHA-2 | AES | TLS 1.3 handshake, идентичный браузерам | [tdd-crypto-gb](../tdd-crypto-gb/) |
| 🇮🇱 Израиль | IL | Israel National Cyber Directorate | ECC профили | SHA-2 | AES | HTTPS-транспорт с обычным TLS стеком | [tdd-crypto-il](../tdd-crypto-il/) |
| 🇧🇷 Бразилия | BR | Instituto Nacional de Tecnologia da Informação | ECDSA | SHA-2 | AES | PKI-Brasil поверх стандартного TLS | [tdd-crypto-br](../tdd-crypto-br/) |
| 🇮🇷 Иран | IR | Iranian National Center for Cyberspace | ECC / RSA | SHA-2 | AES | HTTPS-совместимый TLS стек | [tdd-crypto-ir](../tdd-crypto-ir/) |

### TDD Flows по приоритету

**Phase 1 (MVP):**
- `tdd-crypto-us` - базовый стандарт, максимальная совместимость
- `tdd-crypto-ru` - ГОСТ (приоритет для РФ рынка)
- `tdd-crypto-cn` - SM алгоритмы (приоритет для CN рынка)

**Phase 2:**
- `tdd-crypto-eu` - Brainpool curves
- `tdd-crypto-jp` - Camellia
- `tdd-crypto-kr` - SEED

**Phase 3:**
- Остальные страны по запросу

## Acceptance Criteria

### Must Have

1. **Given** сервер HTTPS VPN и клиент с поддержкой ГОСТ
   **When** клиент устанавливает соединение
   **Then** используются ГОСТ-алгоритмы для шифрования (ГОСТ 34.12-2015, ГОСТ 34.13-2015)

2. **Given** сервер HTTPS VPN и клиент с поддержкой SM-алгоритмов
   **When** клиент устанавливает соединение
   **Then** используются китайские алгоритмы (SM2, SM3, SM4)

3. **Given** существующая инфраструктура с 3x-ui или marzban
   **When** xray заменяется на HTTPS VPN
   **Then** панели управления продолжают работать без модификаций

4. **Given** DPI-система анализирует трафик HTTPS VPN
   **When** трафик проходит через DPI
   **Then** трафик неотличим от стандартного браузерного HTTPS-трафика

5. **Given** клиент HTTPS VPN
   **When** требуется туннелирование произвольного трафика
   **Then** туннелирование выполняется через SOCKS5 протокол

### Should Have

1. Модульная архитектура криптографических провайдеров (см. TDD flows выше)
2. Использование существующих сертифицированных криптографических библиотек
3. Полная совместимость с существующими xray-клиентами (v2rayN, v2rayNG, NekoBox, и др.)

### Won't Have (This Iteration)

1. Собственные панели управления (используем существующие)
2. Собственные клиентские приложения (интеграция в существующие)
3. Поддержка устаревших протоколов (VMess без TLS)
4. Новые протоколы маскировки (используем стандартный TLS/HTTPS)
5. Обфускация трафика сверх TLS

## Constraints

### Technical

- **API-совместимость с xray-core**: конфигурационные файлы должны быть совместимы или требовать минимальных изменений
- **Архитектура трафика**: должна точно соответствовать паттернам браузерного HTTPS/TLS трафика
- **Криптобиблиотеки**: использовать только сертифицированные/проверенные реализации
- **Модульность**: криптопровайдеры должны быть независимыми модулями

### Performance

- Производительность не должна быть значительно ниже xray-core
- Overhead от дополнительного шифрования должен быть минимальным

### Regulatory

- Код должен быть готов к сертификации
- Минимизация объема кода для ускорения сертификации
- Четкое разделение криптографических модулей (каждый сертифицируется отдельно)

### Dependencies

- Совместимость с Go runtime (как xray-core)
- Зависимости от криптобиблиотек определяются в соответствующих TDD flows

## Architecture Principles

### 1. Минимальный код - Максимальное переиспользование

```
+-------------------------+
|      HTTPS VPN Core     |  <- Минимальный "клей" код
+-------------------------+
            |
            v
+-------------------------+
|   Crypto Provider API   |  <- Единый интерфейс
+-------------------------+
     |      |      |
     v      v      v
+------+ +------+ +------+
|  US  | |  RU  | |  CN  |   <- TDD flows (модули)
| AES  | | GOST | |  SM  |
+------+ +------+ +------+
            |
            v
+-------------------------+
|   Standard TLS Stack    |  <- Стандартная TLS-реализация
+-------------------------+
            |
            v
+-------------------------+
|    SOCKS5 Tunnel        |  <- Стандартный SOCKS5
+-------------------------+
```

### 2. Браузероподобный трафик

- TLS handshake идентичен браузерному
- HTTP/2 или HTTP/1.1 как в реальных браузерах
- Паттерны передачи данных соответствуют веб-трафику
- Валидные сертификаты (не самоподписанные в production)

### 3. Drop-in замена xray

- Совместимость с xray config.json
- Поддержка VLESS/Trojan over TLS протоколов
- API для панелей управления

## Open Questions

- [ ] Нужна ли поддержка QUIC (HTTP/3) в первой версии?
- [ ] Поддержка каких xray-клиентов критична в первую очередь?
- [ ] Какие страны добавить в Phase 2/3?

## References

### Standards
- ГОСТ 34.10-2018: Цифровая подпись
- ГОСТ 34.11-2018: Хеш-функции (Streebog)
- ГОСТ 34.12-2015: Блочные шифры (Kuznyechik)
- GB/T 32918: SM2 Elliptic Curve Cryptography
- NIST FIPS 197: AES
- RFC 8446: TLS 1.3

### Ecosystem
- xray-core: https://github.com/XTLS/Xray-core
- 3x-ui: https://github.com/MHSanaei/3x-ui
- marzban: https://github.com/Gozargah/Marzban

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
