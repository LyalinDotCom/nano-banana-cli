package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigFilePath(t *testing.T) {
	t.Setenv("NANOBANANA_CONFIG_DIR", t.TempDir())

	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v", err)
	}

	if filepath.Base(path) != "config.yaml" {
		t.Fatalf("unexpected config filename: %s", path)
	}
	if filepath.Dir(path) != filepath.Clean(os.Getenv("NANOBANANA_CONFIG_DIR")) {
		t.Fatalf("unexpected config directory: %s", path)
	}
}

func TestSetAndLoadAPIKey(t *testing.T) {
	t.Setenv("NANOBANANA_CONFIG_DIR", t.TempDir())

	if err := SetAPIKey("test-key"); err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIKey != "test-key" {
		t.Fatalf("cfg.APIKey = %q, want %q", cfg.APIKey, "test-key")
	}
	if cfg.Model != DefaultModel {
		t.Fatalf("cfg.Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Fatalf("cfg.Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
}

func TestClearAPIKey(t *testing.T) {
	t.Setenv("NANOBANANA_CONFIG_DIR", t.TempDir())

	cfg := &Config{
		APIKey:    "test-key",
		Model:     DefaultModel,
		Timeout:   3 * time.Minute,
		OutputDir: ".",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := ClearAPIKey(); err != nil {
		t.Fatalf("ClearAPIKey() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.APIKey != "" {
		t.Fatalf("loaded.APIKey = %q, want empty", loaded.APIKey)
	}
}
