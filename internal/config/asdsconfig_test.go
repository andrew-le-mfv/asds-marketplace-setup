package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestASDSConfig_Defaults(t *testing.T) {
	_ = config.DefaultASDSConfig()
}

func TestASDSConfig_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.ASDSConfig{}

	if err := config.WriteASDSConfig(path, &cfg); err != nil {
		t.Fatalf("WriteASDSConfig error: %v", err)
	}

	_, err := config.ReadASDSConfig(path)
	if err != nil {
		t.Fatalf("ReadASDSConfig error: %v", err)
	}
}

func TestReadASDSConfig_NotFound_ReturnsDefaults(t *testing.T) {
	_, err := config.ReadASDSConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveASDSConfigPath(t *testing.T) {
	path := config.ResolveASDSConfigPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "asds", "config.yaml")
	if path != expected {
		t.Errorf("config path = %q, want %q", path, expected)
	}
}
