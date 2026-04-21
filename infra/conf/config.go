package conf

import "encoding/json"

// Config is xray-compatible configuration structure.
type Config struct {
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
}

// InboundConfig configures an inbound listener.
type InboundConfig struct {
	Port           int             `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings"`
	StreamSettings *StreamConfig   `json:"streamSettings"`
	Tag            string          `json:"tag,omitempty"`
	Sniffing       *SniffingConfig `json:"sniffing,omitempty"`
}

// OutboundConfig configures an outbound connection.
type OutboundConfig struct {
	Protocol string          `json:"protocol"`
	Settings json.RawMessage `json:"settings"`
	Tag      string          `json:"tag,omitempty"`
}

// StreamConfig configures transport and security settings.
type StreamConfig struct {
	Network     string     `json:"network"`
	Security    string     `json:"security"`
	TLSSettings *TLSConfig `json:"tlsSettings"`
}

// TLSConfig configures TLS settings.
type TLSConfig struct {
	ServerName     string        `json:"serverName"`
	Certificates   []CertConfig  `json:"certificates"`
	CipherSuites   string        `json:"cipherSuites,omitempty"`
	CryptoProvider string        `json:"cryptoProvider,omitempty"` // deprecated: use CipherSuites
	MinVersion     string        `json:"minVersion,omitempty"`
	MaxVersion     string        `json:"maxVersion,omitempty"`
}

// CertConfig configures a TLS certificate.
type CertConfig struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

// SniffingConfig configures traffic sniffing.
type SniffingConfig struct {
	Enabled     bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Inbounds: []InboundConfig{
			{
				Port:     443,
				Protocol: "https-vpn",
				StreamSettings: &StreamConfig{
					Network:  "h2",
					Security: "tls",
					TLSSettings: &TLSConfig{
						CryptoProvider: "us",
						MinVersion:     "TLS1.3",
					},
				},
			},
		},
		Outbounds: []OutboundConfig{
			{
				Protocol: "freedom",
			},
		},
	}
}
