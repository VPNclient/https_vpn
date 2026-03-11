# Specifications: HTTPS VPN

> Version: 1.0
> Status: APPROVED
> Last Updated: 2026-03-10
> Requirements: [01-requirements.md](./01-requirements.md)

## Overview

HTTPS VPN is a ~600 LOC VPN implementation using HTTP/2 CONNECT over TLS. The traffic is indistinguishable from browser HTTPS proxy traffic. Crypto providers are pluggable modules for national cryptography standards.

## Affected Systems

| System | Action | Notes |
|--------|--------|-------|
| `core/` | Create | Main entry point, xray-compatible API |
| `transport/` | Create | HTTP/2 CONNECT handler |
| `crypto/` | Create | Crypto provider interface + adapters |
| `infra/conf/` | Create | xray-compatible config parser |
| `cmd/https-vpn/` | Create | CLI binary |

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                              CLI                                    │
│                         cmd/https-vpn/                              │
│                           (~50 LOC)                                 │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              v
┌─────────────────────────────────────────────────────────────────────┐
│                            core/                                    │
│                         (~100 LOC)                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐ │
│  │ Instance        │  │ Server          │  │ Client              │ │
│  │ New()           │  │ Start()         │  │ Dial()              │ │
│  │ Close()         │  │ Close()         │  │ Close()             │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────┘ │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              v               v               v
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────────────────┐
│  infra/conf/    │ │  transport/     │ │  crypto/                    │
│  (~150 LOC)     │ │  (~120 LOC)     │ │  (~60 LOC interface)        │
│                 │ │                 │ │                             │
│  - ConfigLoader │ │  - H2Server     │ │  ┌─────┐ ┌─────┐ ┌─────┐   │
│  - InboundConf  │ │  - H2Client     │ │  │ us/ │ │ ru/ │ │ cn/ │   │
│  - OutboundConf │ │  - ConnectHdlr  │ │  └─────┘ └─────┘ └─────┘   │
│  - TLSConf      │ │  - Pipe         │ │  (adapters ~30 LOC each)   │
└─────────────────┘ └─────────────────┘ └─────────────────────────────┘
```

### Data Flow

```
                        SERVER MODE

Client Request                              Target Server
     │                                           ▲
     │ TLS ClientHello                           │
     v                                           │
┌─────────────────┐                              │
│  TLS Handshake  │ ◄── CryptoProvider           │
│  (national alg) │                              │
└────────┬────────┘                              │
         │                                       │
         │ HTTP/2 CONNECT target:port            │
         v                                       │
┌─────────────────┐                              │
│ ConnectHandler  │──── Dial(target:port) ───────┘
│                 │
│   Pipe(client,  │
│        target)  │
└─────────────────┘


                        CLIENT MODE

Local App                                   VPN Server
    │                                           │
    │ SOCKS5/HTTP                               │
    v                                           │
┌─────────────────┐                             │
│  Local Proxy    │                             │
│  (SOCKS5)       │                             │
└────────┬────────┘                             │
         │                                      │
         │ Extract target                       │
         v                                      │
┌─────────────────┐      TLS + HTTP/2           │
│  H2Client       │────── CONNECT ──────────────┘
│                 │      target:port
└─────────────────┘
```

## Interfaces

### Crypto Provider Interface

```go
// crypto/provider.go (~30 LOC)

package crypto

import (
    "crypto/tls"
)

// Provider configures TLS with specific cryptographic algorithms
type Provider interface {
    // Name returns provider identifier (e.g., "us", "ru", "cn")
    Name() string

    // ConfigureTLS applies crypto settings to tls.Config
    ConfigureTLS(cfg *tls.Config) error

    // SupportedCipherSuites returns cipher suite IDs
    SupportedCipherSuites() []uint16
}

// Registry holds available providers
var Registry = make(map[string]Provider)

// Register adds a provider to registry
func Register(p Provider) {
    Registry[p.Name()] = p
}

// Get returns provider by name
func Get(name string) (Provider, bool) {
    p, ok := Registry[name]
    return p, ok
}
```

### US Provider (stdlib)

```go
// crypto/us/provider.go (~20 LOC)

package us

import (
    "crypto/tls"
    "github.com/.../https-vpn/crypto"
)

func init() {
    crypto.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "us" }

func (p *Provider) ConfigureTLS(cfg *tls.Config) error {
    cfg.MinVersion = tls.VersionTLS13
    cfg.CipherSuites = p.SupportedCipherSuites()
    return nil
}

func (p *Provider) SupportedCipherSuites() []uint16 {
    return []uint16{
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
    }
}
```

### Core Instance

```go
// core/core.go (~80 LOC)

package core

import (
    "context"
    "github.com/.../https-vpn/infra/conf"
    "github.com/.../https-vpn/transport"
)

// Instance represents HTTPS VPN server/client instance
type Instance struct {
    config  *conf.Config
    server  *transport.H2Server
    client  *transport.H2Client
    ctx     context.Context
    cancel  context.CancelFunc
}

// New creates instance from config (xray-compatible signature)
func New(config *conf.Config) (*Instance, error)

// Start begins accepting connections
func (i *Instance) Start() error

// Close shuts down the instance
func (i *Instance) Close() error
```

### Transport Layer

```go
// transport/server.go (~60 LOC)

package transport

import (
    "crypto/tls"
    "net"
    "net/http"
)

// H2Server handles HTTP/2 CONNECT requests
type H2Server struct {
    listener  net.Listener
    tlsConfig *tls.Config
    handler   http.Handler
}

func NewH2Server(addr string, tlsConfig *tls.Config) (*H2Server, error)
func (s *H2Server) Start() error
func (s *H2Server) Close() error
```

```go
// transport/handler.go (~60 LOC)

package transport

import (
    "io"
    "net"
    "net/http"
)

// ConnectHandler processes HTTP CONNECT requests
type ConnectHandler struct{}

func (h *ConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)

// pipe copies data bidirectionally between connections
func pipe(dst, src net.Conn) error
```

```go
// transport/client.go (~70 LOC)

package transport

import (
    "crypto/tls"
    "net"
    "net/http"
)

// H2Client connects to HTTPS VPN server
type H2Client struct {
    serverAddr string
    tlsConfig  *tls.Config
    httpClient *http.Client
}

func NewH2Client(serverAddr string, tlsConfig *tls.Config) (*H2Client, error)

// Connect establishes tunnel to target via server
func (c *H2Client) Connect(target string) (net.Conn, error)

func (c *H2Client) Close() error
```

### Config Parser (xray-compatible)

```go
// infra/conf/config.go (~100 LOC)

package conf

import (
    "encoding/json"
    "os"
)

// Config is xray-compatible configuration structure
type Config struct {
    Inbounds  []InboundConfig  `json:"inbounds"`
    Outbounds []OutboundConfig `json:"outbounds"`
}

type InboundConfig struct {
    Port           int             `json:"port"`
    Protocol       string          `json:"protocol"`
    Settings       json.RawMessage `json:"settings"`
    StreamSettings *StreamConfig   `json:"streamSettings"`
}

type OutboundConfig struct {
    Protocol string          `json:"protocol"`
    Settings json.RawMessage `json:"settings"`
}

type StreamConfig struct {
    Network     string     `json:"network"`
    Security    string     `json:"security"`
    TLSSettings *TLSConfig `json:"tlsSettings"`
}

type TLSConfig struct {
    ServerName   string        `json:"serverName"`
    Certificates []CertConfig  `json:"certificates"`
    CryptoProvider string      `json:"cryptoProvider"` // extension: "us", "ru", "cn"
}

type CertConfig struct {
    CertificateFile string `json:"certificateFile"`
    KeyFile         string `json:"keyFile"`
}

// LoadConfig reads and parses config file
func LoadConfig(path string) (*Config, error)
```

## Behavior Specifications

### Happy Path: Server

1. Server loads config from JSON file
2. Server initializes crypto provider based on `cryptoProvider` field
3. Server starts TLS listener with HTTP/2 support
4. Client connects with TLS handshake (using national crypto)
5. Client sends `CONNECT target.example.com:443`
6. Server dials target, responds `200 Connection Established`
7. Server pipes data bidirectionally until either side closes

### Happy Path: Client

1. Client loads config from JSON file
2. Client starts local SOCKS5 proxy
3. Local app connects to SOCKS5 proxy, requests `target.example.com:443`
4. Client extracts target from SOCKS5 request
5. Client connects to VPN server via TLS + HTTP/2
6. Client sends `CONNECT target.example.com:443`
7. Client pipes data between local app and VPN server

### Edge Cases

| Case | Trigger | Expected Behavior |
|------|---------|-------------------|
| Target unreachable | Server cannot dial target | Return HTTP 502 Bad Gateway |
| TLS handshake failure | Cert mismatch, crypto mismatch | Return TLS alert, log error |
| Client disconnect | Client closes connection | Close target connection, cleanup |
| Target disconnect | Target closes connection | Close client connection, cleanup |
| Invalid CONNECT | Missing host:port | Return HTTP 400 Bad Request |
| Unsupported method | GET/POST instead of CONNECT | Return HTTP 405 Method Not Allowed |

### Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| `ErrProviderNotFound` | Unknown crypto provider name | Fail startup with clear message |
| `ErrCertLoad` | Cannot read cert/key files | Fail startup with path info |
| `ErrBindFailed` | Port already in use | Fail startup with port info |
| `ErrDialTimeout` | Target not responding | HTTP 504 Gateway Timeout |
| `ErrTLSHandshake` | TLS negotiation failed | Close connection, log details |

## File Structure

```
https-vpn/
├── go.mod
├── go.sum
├── README.md
│
├── core/
│   └── core.go                 # ~80 LOC - Instance, New(), Start(), Close()
│
├── transport/
│   ├── server.go               # ~60 LOC - H2Server
│   ├── client.go               # ~70 LOC - H2Client
│   ├── handler.go              # ~50 LOC - ConnectHandler
│   └── pipe.go                 # ~30 LOC - bidirectional copy
│
├── crypto/
│   ├── provider.go             # ~30 LOC - Provider interface, Registry
│   ├── us/
│   │   └── provider.go         # ~20 LOC - US/NIST provider (stdlib)
│   ├── ru/
│   │   └── provider.go         # ~30 LOC - GOST provider adapter
│   └── cn/
│       └── provider.go         # ~30 LOC - SM provider adapter
│
├── infra/
│   └── conf/
│       ├── config.go           # ~80 LOC - Config structs
│       └── loader.go           # ~70 LOC - LoadConfig()
│
└── cmd/
    └── https-vpn/
        └── main.go             # ~50 LOC - CLI entry point
```

## LOC Budget

| Component | File | LOC | Cumulative |
|-----------|------|-----|------------|
| Core | core/core.go | 80 | 80 |
| Server | transport/server.go | 60 | 140 |
| Client | transport/client.go | 70 | 210 |
| Handler | transport/handler.go | 50 | 260 |
| Pipe | transport/pipe.go | 30 | 290 |
| Crypto Interface | crypto/provider.go | 30 | 320 |
| US Provider | crypto/us/provider.go | 20 | 340 |
| Config Structs | infra/conf/config.go | 80 | 420 |
| Config Loader | infra/conf/loader.go | 70 | 490 |
| CLI | cmd/https-vpn/main.go | 50 | 540 |
| **TOTAL** | | | **~540 LOC** |

Remaining budget: ~60 LOC for edge cases, logging, minor utilities.

## Dependencies

### Required (Go stdlib)

- `crypto/tls` - TLS 1.3 implementation
- `net/http` - HTTP/2 server/client
- `encoding/json` - config parsing
- `context` - cancellation
- `io` - pipe operations

### Optional (Crypto Providers)

| Provider | Library | Build Tag |
|----------|---------|-----------|
| US | Go stdlib | (default) |
| RU | `github.com/AlfredLoworworthy/gostcrypto` | `gost` |
| CN | `github.com/emmansun/gmsm` | `sm` |

## Testing Strategy

### Unit Tests

- [ ] `crypto/provider_test.go` - Registry, Get()
- [ ] `crypto/us/provider_test.go` - ConfigureTLS()
- [ ] `infra/conf/loader_test.go` - LoadConfig() with valid/invalid JSON
- [ ] `transport/handler_test.go` - CONNECT parsing, error responses
- [ ] `transport/pipe_test.go` - Bidirectional copy

### Integration Tests

- [ ] Server starts and accepts TLS connection
- [ ] Client connects and CONNECT succeeds
- [ ] Full tunnel: client -> server -> target -> response -> client
- [ ] Crypto provider switching (us/ru/cn)

### Manual Verification

- [ ] Traffic capture shows standard HTTP/2 CONNECT (Wireshark)
- [ ] TLS fingerprint matches browser (JA3 comparison)
- [ ] Works with 3x-ui panel (config compatibility)
- [ ] Works with marzban panel (config compatibility)

## xray Compatibility Matrix

| xray Feature | HTTPS VPN Support | Notes |
|--------------|-------------------|-------|
| `inbounds[].port` | Yes | Direct mapping |
| `inbounds[].protocol` | Partial | Only "https-vpn" |
| `streamSettings.network: "h2"` | Yes | HTTP/2 |
| `streamSettings.security: "tls"` | Yes | TLS 1.3 |
| `tlsSettings.certificates` | Yes | Cert/key paths |
| `tlsSettings.serverName` | Yes | SNI |
| `outbounds[].protocol: "freedom"` | Yes | Direct connection |
| VMess/VLESS protocols | No | Out of scope |
| WebSocket transport | No | Out of scope |
| gRPC transport | No | Out of scope |

---

## Approval

- [ ] Reviewed by:
- [ ] Approved on:
- [ ] Notes:
