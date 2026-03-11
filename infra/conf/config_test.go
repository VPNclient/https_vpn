package conf_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nativemind/https-vpn/infra/conf"
)

// TestLoadConfig_Valid tests loading a valid config
func TestLoadConfig_Valid(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"inbounds": [{
			"port": 443,
			"protocol": "https-vpn",
			"streamSettings": {
				"network": "h2",
				"security": "tls",
				"tlsSettings": {
					"serverName": "example.com",
					"certificates": [{
						"certificateFile": "/path/to/cert.pem",
						"keyFile": "/path/to/key.pem"
					}],
					"cryptoProvider": "us"
				}
			}
		}],
		"outbounds": [{
			"protocol": "freedom"
		}]
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := conf.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Inbounds) != 1 {
		t.Errorf("Expected 1 inbound, got %d", len(cfg.Inbounds))
	}

	if cfg.Inbounds[0].Port != 443 {
		t.Errorf("Expected port 443, got %d", cfg.Inbounds[0].Port)
	}
}

// TestLoadConfig_InvalidJSON tests loading invalid JSON
func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, err := conf.LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestLoadConfig_NoInbounds tests config with no inbounds
func TestLoadConfig_NoInbounds(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"outbounds": [{"protocol": "freedom"}]
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, err := conf.LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for no inbounds")
	}
}

// TestLoadConfig_InvalidPort tests config with invalid port
func TestLoadConfig_InvalidPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"inbounds": [{
			"port": 99999,
			"protocol": "https-vpn"
		}],
		"outbounds": [{"protocol": "freedom"}]
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, err := conf.LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

// TestSaveConfig tests saving config to file
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := conf.DefaultConfig()
	cfg.Inbounds[0].Port = 8443

	if err := conf.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var loaded map[string]interface{}
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Saved config is not valid JSON: %v", err)
	}
}

// TestDefaultConfig tests default config generation
func TestDefaultConfig(t *testing.T) {
	cfg := conf.DefaultConfig()

	if len(cfg.Inbounds) == 0 {
		t.Error("Expected default inbounds")
	}

	if len(cfg.Outbounds) == 0 {
		t.Error("Expected default outbounds")
	}

	if cfg.Inbounds[0].StreamSettings == nil {
		t.Error("Expected default stream settings")
	}
}
