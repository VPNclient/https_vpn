// Package core provides the main entry point for HTTPS VPN.
// It exposes an xray-compatible API: New(), Start(), Close().
package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

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

// getProviderName returns crypto provider name based on TLS config.
// It also returns a boolean indicating if the deprecated cryptoProvider field was used.
func getProviderName(tlsSettings *conf.TLSConfig) (name string, usedDeprecated bool) {
	if tlsSettings == nil {
		return "us", false
	}

	// 1. Try CipherSuites first (standard xray-compatible way)
	// Supports comma-separated list like "ru,TLS_AES_128_GCM_SHA256"
	if tlsSettings.CipherSuites != "" {
		parts := strings.Split(tlsSettings.CipherSuites, ",")
		for _, part := range parts {
			n := strings.TrimSpace(strings.ToLower(part))
			if _, ok := crypto.Get(n); ok {
				return n, false
			}
		}
	}

	// 2. Fallback to deprecated CryptoProvider
	if tlsSettings.CryptoProvider != "" {
		n := strings.TrimSpace(strings.ToLower(tlsSettings.CryptoProvider))
		if _, ok := crypto.Get(n); ok {
			return n, true
		}
	}

	// Default to US
	return "us", false
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
	providerName := "us"

	if inbound.StreamSettings != nil && inbound.StreamSettings.TLSSettings != nil {
		tlsSettings := inbound.StreamSettings.TLSSettings
		var usedDeprecated bool
		providerName, usedDeprecated = getProviderName(tlsSettings)

		// Log provider selection
		fmt.Printf("Crypto provider: %s\n", providerName)
		if usedDeprecated {
			fmt.Fprintf(os.Stderr, "Warning: cryptoProvider field is deprecated, use cipherSuites instead\n")
		}

		// Load certificates with automatic provider-based selection
		if len(tlsSettings.Certificates) > 0 {
			priority := crypto.ParseProviderPriority(tlsSettings.CipherSuites)
			certStore, err := crypto.NewCertificateStore(tlsSettings.Certificates, priority)
			if err != nil {
				return fmt.Errorf("failed to load certificates: %w", err)
			}
			// Set callback for dynamic certificate selection
			tlsConfig.GetCertificate = certStore.GetCertificate
			// Also populate Certificates for clients that don't trigger callback
			tlsConfig.Certificates = certStore.AllCertificates()
		}

		// Set SNI
		if tlsSettings.ServerName != "" {
			tlsConfig.ServerName = tlsSettings.ServerName
		}

		// Configure crypto provider
		provider, ok := crypto.Get(providerName)
		if !ok {
			return fmt.Errorf("crypto provider not found: %s", providerName)
		}
		if err := provider.ConfigureTLS(tlsConfig); err != nil {
			return err
		}
	}

	// Create server config
	serverCfg := &transport.ServerConfig{
		Addr:           fmt.Sprintf(":%d", inbound.Port),
		TLSConfig:      tlsConfig,
		CryptoProvider: providerName,
		Handler:        getHandler(inbound, &transport.ConnectHandler{}),
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

func getHandler(inbound conf.InboundConfig, next http.Handler) http.Handler {
	if inbound.OcservBackend != "" {
		return &transport.AnyConnectHandler{
			BackendAddr: inbound.OcservBackend,
			Next:        next,
		}
	}
	return next
}
