package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Try environment variables first (highest priority)
	cfg.APIKey = os.Getenv("PORKBUN_API_KEY")
	cfg.SecretKey = os.Getenv("PORKBUN_SECRET_KEY")

	// If env vars are set, return early
	if cfg.APIKey != "" && cfg.SecretKey != "" {
		return cfg, nil
	}

	// Try config file as fallback
	configPath := getConfigPath()
	if configPath != "" {
		fileCfg, err := loadFromFile(configPath)
		if err == nil {
			// Only override if not set by env vars
			if cfg.APIKey == "" {
				cfg.APIKey = fileCfg.APIKey
			}
			if cfg.SecretKey == "" {
				cfg.SecretKey = fileCfg.SecretKey
			}
		}
	}

	// Validate
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing API credentials. Set PORKBUN_API_KEY and PORKBUN_SECRET_KEY environment variables, or create ~/.config/porkbun-tui/config.yaml")
	}

	return cfg, nil
}

func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "porkbun-tui", "config.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fall back to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(home, ".config", "porkbun-tui", "config.yaml")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
