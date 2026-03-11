// Package core provides the main entry point for HTTPS VPN.
// It exposes an xray-compatible API: New(), Start(), Close().
package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"

	"github.com/nativemind/https-vpn/crypto"
	"github.com/nativemind/https-vpn/infra/conf"
	"github.com/nativemind/https-vpn/transport"
)

// Instance represents an HTTPS VPN server instance.
type Instance struct {
	config   *conf.Config
	server   *transport.H2Server
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new HTTPS VPN instance from config.
// This function has an xray-compatible signature.
func New(config *conf.Config) (*Instance, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if len(config.Inbounds) == 0 {
		return nil, fmt.Errorf("no inbounds configured")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Instance{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start begins accepting connections.
// This function has an xray-compatible signature.
func (i *Instance) Start() error {
	inbound := i.config.Inbounds[0] // Use first inbound for now

	// Get TLS settings
	tlsConfig := &tls.Config{}

	if inbound.StreamSettings != nil && inbound.StreamSettings.TLSSettings != nil {
		tlsSettings := inbound.StreamSettings.TLSSettings

		// Load certificates
		if len(tlsSettings.Certificates) > 0 {
			cert := tlsSettings.Certificates[0]
			certPair, err := tls.LoadX509KeyPair(cert.CertificateFile, cert.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{certPair}
		}

		// Set SNI
		if tlsSettings.ServerName != "" {
			tlsConfig.ServerName = tlsSettings.ServerName
		}

		// Configure crypto provider
		if tlsSettings.CryptoProvider != "" {
			provider, ok := crypto.Get(tlsSettings.CryptoProvider)
			if !ok {
				return fmt.Errorf("crypto provider not found: %s", tlsSettings.CryptoProvider)
			}
			if err := provider.ConfigureTLS(tlsConfig); err != nil {
				return err
			}
		}
	}

	// Create server config
	serverCfg := &transport.ServerConfig{
		Addr:         fmt.Sprintf(":%d", inbound.Port),
		TLSConfig:    tlsConfig,
		CryptoProvider: inbound.StreamSettings.TLSSettings.CryptoProvider,
		Handler:      &transport.ConnectHandler{},
	}

	// Create server
	server, err := transport.NewH2Server(serverCfg)
	if err != nil {
		return err
	}

	i.server = server

	// Start server in background
	go func() {
		if err := server.Start(); err != nil && err != net.ErrClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	return nil
}

// Close shuts down the instance.
// This function has an xray-compatible signature.
func (i *Instance) Close() error {
	if i.cancel != nil {
		i.cancel()
	}

	if i.server != nil {
		return i.server.Shutdown(i.ctx)
	}

	return nil
}
