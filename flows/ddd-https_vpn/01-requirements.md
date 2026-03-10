# Requirements: HTTPS VPN

> Version: 1.2
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

## Technical Decisions

### DECISION-001: Transport Protocol

**Выбрано: HTTP/2 CONNECT Proxy over TLS**

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser (reference)                      │
├─────────────────────────────────────────────────────────────┤
│  Client ──TLS 1.3──> HTTP/2 ──CONNECT──> [tunnel data]     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    HTTPS VPN (implementation)               │
├─────────────────────────────────────────────────────────────┤
│  Client ──TLS 1.3──> HTTP/2 ──CONNECT──> [tunnel data]     │
│          (+ nat'l crypto)                                   │
└─────────────────────────────────────────────────────────────┘
```

**Обоснование:**
- Идентичен браузерному HTTPS proxy (RFC 7540 + RFC 7231)
- AI-based DPI не может отличить - это тот же самый протокол
- Go stdlib поддерживает из коробки (`net/http` с HTTP/2)
- Минимум кода: ~600 LOC (без учета crypto providers)

**HTTP/3 (QUIC):** Отложено. HTTP/2 достаточно для MVP.

### DECISION-002: xray-core API Compatibility

**Требование: Drop-in library replacement**

xray-core и его клиенты **вне скоупа** данного проекта. Однако:

1. **Функции и методы** - те же имена что в xray-core
2. **JSON конфиг** - тот же формат и структура
3. **Go package API** - совместимые сигнатуры

При замене `import "github.com/xtls/xray-core/..."` на `import "github.com/.../https-vpn/..."` код должен компилироваться без изменений.

**Пример совместимости:**

```go
// xray-core style config (должен работать as-is)
{
  "inbounds": [{
    "port": 443,
    "protocol": "vless",
    "settings": { ... },
    "streamSettings": {
      "network": "h2",
      "security": "tls",
      "tlsSettings": { ... }
    }
  }],
  "outbounds": [{
    "protocol": "freedom"
  }]
}
```

### DECISION-003: Code Size Target

**Цель: ~600 LOC нового кода** (без crypto providers)

```
┌─────────────────────────────────┬────────┬─────────────────┐
│ Компонент                       │ LOC    │ Сертификация    │
├─────────────────────────────────┼────────┼─────────────────┤
│ HTTP/2 CONNECT handler          │ ~80    │ Требуется       │
│ Bidirectional pipe              │ ~40    │ Требуется       │
│ TLS config + provider interface │ ~60    │ Требуется       │
│ xray-compat config parser       │ ~150   │ Требуется       │
│ Main + CLI                      │ ~50    │ Требуется       │
│ Client library                  │ ~220   │ Требуется       │
├─────────────────────────────────┼────────┼─────────────────┤
│ ИТОГО новый код                 │ ~600   │ ~600 LOC        │
├─────────────────────────────────┼────────┼─────────────────┤
│ Go stdlib (net/http, crypto)    │ -      │ Не требуется    │
│ Crypto providers (GOST, SM)     │ -      │ Уже сертифиц.   │
│ uTLS (fingerprinting)           │ ~0     │ Опционально     │
└─────────────────────────────────┴────────┴─────────────────┘

Сравнение: xray-core ~100,000 LOC vs HTTPS VPN ~600 LOC
Упрощение сертификации: ~166x меньше кода
```

## National Cryptography Standards

Каждый криптостандарт реализуется в отдельном TDD flow для модульности и независимой сертификации.

| Страна | ISO | Crypto Org | PKI / Signature | Hash | Symmetric | Browser-compatible transport | TDD Flow |
|--------|-----|------------|-----------------|------|-----------|------------------------------|----------|
| 🇺🇸 США | US | National Institute of Standards and Technology | ECDSA / EdDSA | SHA-2 / SHA-3 | AES | TLS 1.3 + HTTP/2 с параметрами Chrome/Firefox | [tdd-crypto-us](../tdd-crypto-us/) |
| 🇨🇳 Китай | CN | State Cryptography Administration | SM2 | SM3 | SM4 | GMSSL (вариант TLS) с браузероподобным handshake | [tdd-crypto-cn](../tdd-crypto-cn/) |
| 🇷🇺 Россия | RU | Federal Security Service | GOST R 34.10 | Streebog | Kuznyechik | TLS с GOST cipher suites внутри стандартного HTTPS | [tdd-crypto-ru](../tdd-crypto-ru/) |
| 🇰🇷 Южная Корея | KR | Korea Internet & Security Agency | KCDSA | HAS-160 | SEED | TLS с SEED cipher suites | [tdd-crypto-kr](../tdd-crypto-kr/) |
| 🇯🇵 Япония | JP | CRYPTREC | ECDSA профили | SHA-2 | Camellia | TLS с Camellia cipher suites | [tdd-crypto-jp](../tdd-crypto-jp/) |
| 🇮🇳 Индия | IN | Standardisation Testing and Quality Certification Directorate | ECSDSA | SHA-2 | AES | TLS с ECC профилями | [tdd-crypto-in](../tdd-crypto-in/) |
| 🇪🇺 ЕС | EU | European Telecommunications Standards Institute | Brainpool ECC | SHA-2 | AES | TLS с Brainpool curves | [tdd-crypto-eu](../tdd-crypto-eu/) |
| 🇫🇷 Франция | FR | Agence nationale de la sécurité des systèmes d'information | ECDSA | SHA-256 | AES | Стандартный TLS | [tdd-crypto-fr](../tdd-crypto-fr/) |
| 🇬🇧 Великобритания | GB | National Cyber Security Centre | ECDSA | SHA-2 | AES | TLS 1.3 идентичный браузерам | [tdd-crypto-gb](../tdd-crypto-gb/) |
| 🇮🇱 Израиль | IL | Israel National Cyber Directorate | ECC профили | SHA-2 | AES | HTTPS с обычным TLS | [tdd-crypto-il](../tdd-crypto-il/) |
| 🇧🇷 Бразилия | BR | Instituto Nacional de Tecnologia da Informação | ECDSA | SHA-2 | AES | PKI-Brasil поверх TLS | [tdd-crypto-br](../tdd-crypto-br/) |
| 🇮🇷 Иран | IR | Iranian National Center for Cyberspace | ECC / RSA | SHA-2 | AES | HTTPS-совместимый TLS | [tdd-crypto-ir](../tdd-crypto-ir/) |

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

1. **Given** сервер HTTPS VPN настроен с crypto provider
   **When** клиент устанавливает соединение
   **Then** TLS handshake использует алгоритмы выбранного провайдера

2. **Given** DPI-система (включая AI-based) анализирует трафик
   **When** трафик проходит через анализатор
   **Then** трафик неотличим от браузерного HTTP/2 over TLS

3. **Given** существующий xray JSON конфиг
   **When** используется с HTTPS VPN библиотекой
   **Then** конфиг парсится и работает без модификаций

4. **Given** код использующий xray-core API
   **When** import заменяется на https-vpn
   **Then** код компилируется без изменений

5. **Given** HTTP/2 CONNECT запрос от клиента
   **When** сервер обрабатывает запрос
   **Then** устанавливается bidirectional tunnel к целевому хосту

### Should Have

1. Модульная архитектура криптографических провайдеров
2. uTLS для browser fingerprint имитации
3. Полная совместимость с панелями 3x-ui, marzban

### Won't Have (This Iteration)

1. Собственные панели управления
2. Собственные клиентские приложения
3. HTTP/3 (QUIC) поддержка
4. Протоколы VMess/VLESS (только HTTP/2 CONNECT)
5. Обфускация сверх TLS

## Architecture

### Core Architecture (~600 LOC)

```
┌──────────────────────────────────────────────────────────────┐
│                      HTTPS VPN Core                          │
│                        (~600 LOC)                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐  │
│  │ Config Parser  │  │ HTTP/2 Server  │  │ CONNECT       │  │
│  │ (xray-compat)  │  │ (stdlib)       │  │ Handler       │  │
│  │ ~150 LOC       │  │ 0 LOC          │  │ ~120 LOC      │  │
│  └───────┬────────┘  └───────┬────────┘  └───────┬───────┘  │
│          │                   │                   │          │
│          v                   v                   v          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Crypto Provider Interface                │  │
│  │                      (~60 LOC)                        │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                              │
└──────────────────────────────┼──────────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
         v                     v                     v
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  tdd-crypto-us  │  │  tdd-crypto-ru  │  │  tdd-crypto-cn  │
│  (Go stdlib)    │  │  (GOST libs)    │  │  (SM libs)      │
│  0 LOC          │  │  adapter only   │  │  adapter only   │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### Traffic Flow

```
Client App
    │
    │ SOCKS5/HTTP Proxy (local)
    v
┌─────────────────┐
│  HTTPS VPN      │
│  Client         │
└────────┬────────┘
         │
         │ TLS 1.3 + HTTP/2 (browser-identical)
         │ CONNECT target.com:443
         v
┌─────────────────┐
│  HTTPS VPN      │
│  Server         │
└────────┬────────┘
         │
         │ TCP connection
         v
    target.com:443
```

### xray-core API Compatibility

```go
// Целевая совместимость:

// До (xray-core)
import "github.com/xtls/xray-core/core"
server, _ := core.New(config)
server.Start()

// После (https-vpn) - тот же код работает
import "github.com/.../https-vpn/core"
server, _ := core.New(config)
server.Start()
```

**Совместимые пакеты:**
- `core` - основной entry point
- `common/net` - network utilities
- `transport/internet` - transport layer
- `infra/conf` - config parsing

## Constraints

### Technical

- **xray API совместимость**: функции, методы, JSON конфиг - те же имена и структура
- **~600 LOC лимит**: весь новый код должен укладываться в этот бюджет
- **Go stdlib**: максимальное использование стандартной библиотеки
- **HTTP/2 only**: никаких кастомных протоколов

### Performance

- Latency overhead: <5ms на соединение
- Throughput: не менее 90% от raw TLS

### Regulatory

- Код готов к сертификации (~600 LOC vs ~100,000 LOC)
- Криптомодули изолированы и заменяемы
- Четкое разделение: core (~600 LOC) + crypto providers (сертифицированные библиотеки)

## Open Questions

*Все вопросы закрыты в версии 1.2*

## References

### Standards
- RFC 7540: HTTP/2
- RFC 7231: HTTP/1.1 Semantics (CONNECT method)
- RFC 8446: TLS 1.3
- ГОСТ 34.10-2018, 34.11-2018, 34.12-2015
- GB/T 32918: SM2/SM3/SM4

### Ecosystem
- xray-core: https://github.com/XTLS/Xray-core
- 3x-ui: https://github.com/MHSanaei/3x-ui
- marzban: https://github.com/Gozargah/Marzban

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
