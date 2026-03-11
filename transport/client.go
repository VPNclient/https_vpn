package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/nativemind/https-vpn/crypto"
)

// H2Client connects to HTTPS VPN server via HTTP/2 CONNECT.
type H2Client struct {
	serverAddr string
	tlsConfig  *tls.Config
	httpClient *http.Client
}

// ClientConfig holds configuration for H2Client.
type ClientConfig struct {
	ServerAddr   string
	TLSConfig    *tls.Config
	CryptoProvider string
}

// NewH2Client creates a new HTTP/2 client.
func NewH2Client(cfg *ClientConfig) (*H2Client, error) {
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

	// Create HTTP client with HTTP/2 transport
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		ForceAttemptHTTP2: true,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	return &H2Client{
		serverAddr: cfg.ServerAddr,
		tlsConfig:  tlsConfig,
		httpClient: httpClient,
	}, nil
}

// Connect establishes a tunnel to target via the VPN server.
// Returns a net.Conn that can be used to communicate with the target.
func (c *H2Client) Connect(target string) (net.Conn, error) {
	// Create CONNECT request
	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Host: c.serverAddr,
		},
		Host: target,
		Header: make(http.Header),
	}

	// Send CONNECT request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("CONNECT failed: %s", resp.Status)
	}

	// Hijack the connection to get raw access
	hijacker, ok := resp.Body.(http.Hijacker)
	if !ok {
		resp.Body.Close()
		return nil, fmt.Errorf("response body doesn't support hijacking")
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	return conn, nil
}

// Close closes the client and releases resources.
func (c *H2Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// DialContext implements net.Dialer interface for compatibility.
func (c *H2Client) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return c.Connect(addr)
}
