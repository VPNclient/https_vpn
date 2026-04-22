# Status: sdd-https-vpn-amnezia-server-backward-compatibility

## Current Phase

REQUIREMENTS

## Phase Status

DRAFTING

## Last Updated

2026-04-21 by Claude

## Blockers

- [None]

## Progress

- [x] Requirements drafted (v2.0 - скорректирован scope)
- [ ] Requirements approved
- [ ] Specifications drafted
- [ ] Specifications approved
- [ ] Plan drafted
- [ ] Plan approved
- [ ] Implementation started
- [ ] Implementation complete

## Context Notes

Key decisions and context for resuming:

- **Scope pivot (v2.0)**: Вместо интеграции напрямую в Amnezia, мы предоставляем **CLI binary wrapper** который Amnezia запускает как subprocess
- **Wrapper type**: CLI executable (аналогично тому как Amnezia работает с openvpn/wg binary)
- **Transport**: Только HTTP/2 (h2)
- **Model**: Amnezia → subprocess → https-vpn-cli → h2 tunnel → TUN adapter

## Next Actions

1. Получить approval на requirements v2.0
2. Исследовать Amnezia client/protocols для точного понимания интерфейса subprocess
3. Перейти к SPECIFICATIONS фазе после approval
