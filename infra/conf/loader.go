package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig reads and parses a config file.
// Supports xray-compatible JSON config format.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate config
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig writes config to a file.
func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateConfig validates the configuration.
func validateConfig(cfg *Config) error {
	if len(cfg.Inbounds) == 0 {
		return fmt.Errorf("no inbounds configured")
	}

	if len(cfg.Outbounds) == 0 {
		return fmt.Errorf("no outbounds configured")
	}

	for i, inbound := range cfg.Inbounds {
		if inbound.Port <= 0 || inbound.Port > 65535 {
			return fmt.Errorf("inbound %d: invalid port %d", i, inbound.Port)
		}

		if inbound.StreamSettings != nil {
			if err := validateStreamConfig(inbound.StreamSettings); err != nil {
				return fmt.Errorf("inbound %d: %w", i, err)
			}
		}
	}

	return nil
}

// validateStreamConfig validates stream settings.
func validateStreamConfig(sc *StreamConfig) error {
	if sc.Security == "tls" && sc.TLSSettings != nil {
		if len(sc.TLSSettings.Certificates) == 0 {
			return fmt.Errorf("TLS enabled but no certificates configured")
		}

		for i, cert := range sc.TLSSettings.Certificates {
			if cert.CertificateFile == "" {
				return fmt.Errorf("certificate %d: certificateFile is required", i)
			}
			if cert.KeyFile == "" {
				return fmt.Errorf("certificate %d: keyFile is required", i)
			}
		}
	}

	return nil
}
