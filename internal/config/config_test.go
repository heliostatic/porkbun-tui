package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromEnvVars(t *testing.T) {
	// Set env vars
	os.Setenv("PORKBUN_API_KEY", "pk1_test")
	os.Setenv("PORKBUN_SECRET_KEY", "sk1_test")
	defer os.Unsetenv("PORKBUN_API_KEY")
	defer os.Unsetenv("PORKBUN_SECRET_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "pk1_test" {
		t.Errorf("expected APIKey 'pk1_test', got '%s'", cfg.APIKey)
	}
	if cfg.SecretKey != "sk1_test" {
		t.Errorf("expected SecretKey 'sk1_test', got '%s'", cfg.SecretKey)
	}
}

func TestLoad_MissingCredentials(t *testing.T) {
	// Clear env vars
	os.Unsetenv("PORKBUN_API_KEY")
	os.Unsetenv("PORKBUN_SECRET_KEY")

	// Also clear XDG_CONFIG_HOME to avoid loading from file
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
}

func TestLoad_PartialEnvVars(t *testing.T) {
	// Only set API key, not secret
	os.Setenv("PORKBUN_API_KEY", "pk1_test")
	os.Unsetenv("PORKBUN_SECRET_KEY")
	defer os.Unsetenv("PORKBUN_API_KEY")

	// Clear XDG_CONFIG_HOME to avoid loading from file
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for partial credentials, got nil")
	}
}

func TestLoad_FromFile(t *testing.T) {
	// Clear env vars
	os.Unsetenv("PORKBUN_API_KEY")
	os.Unsetenv("PORKBUN_SECRET_KEY")

	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "porkbun-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `api_key: pk1_from_file
secret_key: sk1_from_file
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to temp dir
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "pk1_from_file" {
		t.Errorf("expected APIKey 'pk1_from_file', got '%s'", cfg.APIKey)
	}
	if cfg.SecretKey != "sk1_from_file" {
		t.Errorf("expected SecretKey 'sk1_from_file', got '%s'", cfg.SecretKey)
	}
}

func TestLoad_EnvVarsPrecedence(t *testing.T) {
	// Create temp config dir with file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "porkbun-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `api_key: pk1_from_file
secret_key: sk1_from_file
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Set env vars - should take precedence
	os.Setenv("PORKBUN_API_KEY", "pk1_from_env")
	os.Setenv("PORKBUN_SECRET_KEY", "sk1_from_env")
	defer os.Unsetenv("PORKBUN_API_KEY")
	defer os.Unsetenv("PORKBUN_SECRET_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env vars should win
	if cfg.APIKey != "pk1_from_env" {
		t.Errorf("expected APIKey 'pk1_from_env', got '%s'", cfg.APIKey)
	}
	if cfg.SecretKey != "sk1_from_env" {
		t.Errorf("expected SecretKey 'sk1_from_env', got '%s'", cfg.SecretKey)
	}
}
