# Requirements: CLI Wrapper для интеграции https-vpn в Amnezia

> Version: 2.0
> Status: DRAFT
> Last Updated: 2026-04-21

## Problem Statement

Amnezia VPN — популярное решение с хорошим UX, но оно не поддерживает наш h2 транспорт. Вместо модификации кода Amnezia, мы предоставим **CLI binary wrapper**, который Amnezia сможет запускать как subprocess — аналогично тому, как они работают с OpenVPN и WireGuard.

## Scope Clarification

**ЧТО МЫ ДЕЛАЕМ:**
- CLI executable (wrapper) который Amnezia запускает как внешний процесс
- Поддержка h2 (HTTP/2) транспорта
- Совместимость с моделью "protocol binary" в Amnezia

**ЧТО МЫ НЕ ДЕЛАЕМ:**
- Модификация кода Amnezia Client/Server
- Поддержка других транспортов (HTTP/1.1, WebSocket) — только h2

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Amnezia Client                           │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Protocol Manager                                    │    │
│  │  ┌─────────┐ ┌─────────┐ ┌───────────────────────┐  │    │
│  │  │ OpenVPN │ │WireGuard│ │ https-vpn-cli (ours)  │  │    │
│  │  └────┬────┘ └────┬────┘ └───────────┬───────────┘  │    │
│  └───────┼───────────┼──────────────────┼──────────────┘    │
└──────────┼───────────┼──────────────────┼───────────────────┘
           │           │                  │
           │ subprocess│                  │ subprocess
           ▼           ▼                  ▼
      ┌────────┐  ┌────────┐      ┌──────────────┐
      │openvpn │  │   wg   │      │https-vpn-cli │
      │ binary │  │ binary │      │   (h2)       │
      └────────┘  └────────┘      └──────────────┘
```

## User Stories

### Primary

**As a** разработчик Amnezia интеграции
**I want** предоставить CLI wrapper для https-vpn
**So that** Amnezia может запускать наш h2 транспорт как внешний процесс без модификации их кода.

### Secondary

**As a** пользователь Amnezia
**I want** выбрать https-vpn (h2) как протокол в настройках
**So that** я могу использовать h2 транспорт с привычным UI Amnezia.

**As a** администратор сервера
**I want** развернуть https-vpn-cli на сервере через Amnezia
**So that** клиенты Amnezia могут подключаться по h2.

## Acceptance Criteria

### Must Have

1. **CLI Binary Interface**
   - **Given** https-vpn-cli executable
   - **When** Amnezia вызывает его с конфигурацией
   - **Then** wrapper устанавливает h2 туннель и управляет TUN/TAP интерфейсом

2. **Configuration Format**
   - **Given** конфиг файл или CLI аргументы
   - **When** https-vpn-cli запускается
   - **Then** он принимает конфигурацию в формате совместимом с Amnezia (JSON или CLI flags)

3. **Lifecycle Management**
   - **Given** запущенный https-vpn-cli процесс
   - **When** Amnezia отправляет SIGTERM/SIGINT
   - **Then** wrapper корректно завершает соединение и освобождает ресурсы

4. **Status Reporting**
   - **Given** работающий туннель
   - **When** Amnezia запрашивает статус
   - **Then** wrapper выводит статус в stdout (JSON или plain text)

### Should Have

- Логирование в формате совместимом с Amnezia log viewer
- Поддержка reconnect при обрыве соединения
- Метрики трафика (bytes in/out) доступные через CLI

### Won't Have (This Iteration)

- GUI компоненты
- Поддержка HTTP/1.1, WebSocket или других транспортов
- Модификация Amnezia клиента для нативной поддержки

## Technical Constraints

1. **Executable Format**: Статически слинкованный бинарник (Go) для Linux/macOS/Windows
2. **Transport**: Только HTTP/2 (h2) over TLS
3. **Interface Model**: TUN adapter (аналогично WireGuard)
4. **IPC**: stdout/stderr для логов, exit codes для статуса

## CLI Interface Draft

```bash
# Базовый запуск
https-vpn-cli connect --config /path/to/config.json

# С параметрами
https-vpn-cli connect \
  --server vpn.example.com:443 \
  --transport h2 \
  --tun-name tun-https \
  --log-level info

# Проверка статуса (для Amnezia polling)
https-vpn-cli status --format json

# Корректное завершение
kill -SIGTERM <pid>
```

## Config Format (JSON)

```json
{
  "server": "vpn.example.com:443",
  "transport": "h2",
  "credentials": {
    "type": "certificate",
    "cert_path": "/path/to/client.crt",
    "key_path": "/path/to/client.key"
  },
  "tun": {
    "name": "tun-https",
    "mtu": 1400
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

## Open Questions

- [ ] Какой формат конфига предпочтителен для Amnezia? (JSON vs TOML vs CLI flags)
- [ ] Нужна ли интеграция с Amnezia API для генерации конфигов?
- [ ] Требуется ли поддержка "cloak" режима для маскировки трафика?

## References

- [Amnezia Client Source](https://github.com/amnezia-vpn/amnezia-client) — см. как они запускают OpenVPN/WireGuard
- [Amnezia Protocols](https://github.com/amnezia-vpn/amnezia-client/tree/dev/client/protocols) — интерфейсы протоколов
- https-vpn core: `core/core.go`

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
