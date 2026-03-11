package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/nativemind/https-vpn/crypto"
)

// H2Server handles HTTP/2 CONNECT requests over TLS.
type H2Server struct {
	server   *http.Server
	listener net.Listener
}

// ServerConfig holds configuration for H2Server.
type ServerConfig struct {
	Addr         string
	TLSConfig    *tls.Config
	CryptoProvider string
	Handler      http.Handler
}

// NewH2Server creates a new HTTP/2 server.
func NewH2Server(cfg *ServerConfig) (*H2Server, error) {
	// Create TLS config
	tlsConfig := cfg.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	// Apply crypto provider settings
	if cfg.CryptoProvider != "" {
		provider, ok := crypto.Get(cfg.CryptoProvider)
		if !ok {
			return nil, fmt.Errorf("crypto provider not found: %s", cfg.CryptoProvider)
		}
		if err := provider.ConfigureTLS(tlsConfig); err != nil {
			return nil, err
		}
	}

	// Force HTTP/2
	tlsConfig.NextProtos = []string{"h2"}

	// Create listener
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	// Wrap with TLS
	tlsListener := tls.NewListener(listener, tlsConfig)

	// Create HTTP server
	handler := cfg.Handler
	if handler == nil {
		handler = &ConnectHandler{}
	}

	server := &http.Server{
		Handler: handler,
	}

	return &H2Server{
		server:   server,
		listener: tlsListener,
	}, nil
}

// Start begins accepting connections.
func (s *H2Server) Start() error {
	return s.server.Serve(s.listener)
}

// Shutdown gracefully stops the server.
func (s *H2Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Close immediately closes the server.
func (s *H2Server) Close() error {
	return s.server.Close()
}
