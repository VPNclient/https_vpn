# HTTPS VPN

A lightweight, certification-ready VPN that uses standard HTTP/2 CONNECT over TLS — indistinguishable from regular browser traffic.

## Why HTTPS VPN?

| Problem | HTTPS VPN Solution |
|---------|-------------------|
| VPN protocols have unique signatures detectable by DPI | Uses standard HTTP/2 CONNECT — identical to browser HTTPS proxy |
| No support for national cryptography standards | Pluggable crypto providers (GOST, SM2/SM3/SM4, etc.) |
| Large codebases (~100K LOC) are expensive to certify | ~600 LOC core — 166x less code to audit |
| Complex integration with existing infrastructure | Drop-in xray-core API compatible library |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTPS VPN (~600 LOC)                     │
├─────────────────────────────────────────────────────────────┤
│  Config Parser    │  HTTP/2 Server   │  CONNECT Handler    │
│  (xray-compat)    │  (Go stdlib)     │                     │
├─────────────────────────────────────────────────────────────┤
│                  Crypto Provider Interface                  │
├──────────────┬──────────────┬──────────────┬───────────────┤
│  US (AES)    │  RU (GOST)   │  CN (SM)     │  ...          │
│  stdlib      │  certified   │  certified   │               │
└──────────────┴──────────────┴──────────────┴───────────────┘
```

## Traffic Pattern

```
Browser HTTPS Proxy:    Client ──TLS 1.3──> HTTP/2 ──CONNECT──> [data]
HTTPS VPN:              Client ──TLS 1.3──> HTTP/2 ──CONNECT──> [data]
                                 └─ with national crypto ─┘
```

AI-based DPI cannot distinguish HTTPS VPN traffic from regular browser traffic because it **is** the same protocol (RFC 7540 + RFC 7231).

## Supported Cryptography Standards

| Country | Standard | Signature | Hash | Cipher | Status |
|---------|----------|-----------|------|--------|--------|
| 🇺🇸 USA | NIST | ECDSA | SHA-2 | AES | Phase 1 |
| 🇷🇺 Russia | GOST | GOST R 34.10 | Streebog | Kuznyechik | Phase 1 |
| 🇨🇳 China | GM/T | SM2 | SM3 | SM4 | Phase 1 |
| 🇪🇺 EU | ETSI | Brainpool | SHA-2 | AES | Phase 2 |
| 🇯🇵 Japan | CRYPTREC | ECDSA | SHA-2 | Camellia | Phase 2 |
| 🇰🇷 Korea | KISA | KCDSA | HAS-160 | SEED | Phase 2 |

## xray-core Compatibility

HTTPS VPN is designed as a drop-in replacement for xray-core library:

```go
// Before (xray-core)
import "github.com/xtls/xray-core/core"
server, _ := core.New(config)
server.Start()

// After (https-vpn) — same code works
import "github.com/example/https-vpn/core"
server, _ := core.New(config)
server.Start()
```

Existing xray JSON configs work without modification:

```json
{
  "inbounds": [{
    "port": 443,
    "protocol": "https-vpn",
    "settings": {},
    "streamSettings": {
      "network": "h2",
      "security": "tls",
      "tlsSettings": {
        "certificates": [{"certificateFile": "...", "keyFile": "..."}]
      }
    }
  }],
  "outbounds": [{"protocol": "freedom"}]
}
```

Compatible with management panels: **3x-ui**, **Marzban**, and xray-based applications.

## Code Size Comparison

```
┌─────────────────────┬─────────────┬───────────────────┐
│ Component           │ xray-core   │ HTTPS VPN         │
├─────────────────────┼─────────────┼───────────────────┤
│ Core code           │ ~100,000    │ ~600 LOC          │
│ Certification scope │ ~100,000    │ ~600 LOC          │
│ Audit effort        │ Months      │ Days              │
│ Attack surface      │ Large       │ Minimal           │
└─────────────────────┴─────────────┴───────────────────┘
```

## Quick Start

### Server

```bash
# Generate config
https-vpn init --crypto us

# Run server
https-vpn run -c config.json
```

### Client

```bash
# Connect to server
https-vpn client -s server.example.com:443 -l 127.0.0.1:1080
```

Local SOCKS5 proxy available at `127.0.0.1:1080`.

## Building

```bash
# Default (US crypto - Go stdlib)
go build -o https-vpn ./cmd/https-vpn

# With GOST support
go build -tags gost -o https-vpn ./cmd/https-vpn

# With SM support
go build -tags sm -o https-vpn ./cmd/https-vpn
```

## Project Structure

```
https-vpn/
├── core/                 # Main entry point (xray-compatible)
├── transport/            # HTTP/2 CONNECT implementation
├── crypto/               # Crypto provider interface
│   ├── us/               # NIST (Go stdlib)
│   ├── ru/               # GOST provider
│   └── cn/               # SM provider
├── infra/conf/           # Config parsing (xray-compatible)
└── cmd/https-vpn/        # CLI
```

## Design Principles

1. **Minimal code** — ~600 LOC core, everything else is stdlib or certified libraries
2. **Browser-identical traffic** — HTTP/2 CONNECT over TLS, same as browser HTTPS proxy
3. **Pluggable crypto** — swap crypto providers without changing core code
4. **Certification-ready** — small attack surface, isolated crypto modules
5. **xray-compatible** — same API, same config format, drop-in replacement

## Documentation

- [Requirements](flows/ddd-https_vpn/01-requirements.md) — detailed requirements and decisions
- [Specifications](flows/ddd-https_vpn/02-specifications.md) — technical specifications
- [Implementation Plan](flows/ddd-https_vpn/03-plan.md) — development roadmap

## License

[TBD]

## Contributing

[TBD]
