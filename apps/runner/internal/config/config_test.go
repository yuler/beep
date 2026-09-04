package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLoadFromEnv(t *testing.T) {
	os.Setenv("BEEP_SERVER", "https://beep.example.com")
	os.Setenv("BEEP_RUNNER_TOKEN", "beep_rt_test123")
	os.Setenv("BEEP_ALLOW_EXEC", "true")
	os.Setenv("BEEP_CONCURRENCY", "10")
	os.Setenv("BEEP_POLL_INTERVAL", "5s")
	os.Setenv("BEEP_HOSTNAME", "my-nas")
	defer func() {
		os.Unsetenv("BEEP_SERVER")
		os.Unsetenv("BEEP_RUNNER_TOKEN")
		os.Unsetenv("BEEP_ALLOW_EXEC")
		os.Unsetenv("BEEP_CONCURRENCY")
		os.Unsetenv("BEEP_POLL_INTERVAL")
		os.Unsetenv("BEEP_HOSTNAME")
	}()

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerURL != "https://beep.example.com" {
		t.Errorf("expected server URL https://beep.example.com, got %s", cfg.ServerURL)
	}
	if cfg.RunnerToken != "beep_rt_test123" {
		t.Errorf("expected token beep_rt_test123, got %s", cfg.RunnerToken)
	}
	if !cfg.AllowExec {
		t.Errorf("expected allow_exec to be true")
	}
	if cfg.Concurrency != 10 {
		t.Errorf("expected concurrency 10, got %d", cfg.Concurrency)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected poll interval 5s, got %v", cfg.PollInterval)
	}
	if cfg.Hostname != "my-nas" {
		t.Errorf("expected hostname my-nas, got %s", cfg.Hostname)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestConfigFilePersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	allow := true
	fc := &FileConfig{
		ServerURL:    "https://core.internal:8080",
		RunnerToken:  "beep_rt_saved_token",
		AllowExec:    &allow,
		Concurrency:  8,
		PollInterval: "2s",
	}

	if err := SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save config file: %v", err)
	}

	loaded, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load config file: %v", err)
	}
	if loaded.ServerURL != "https://core.internal:8080" {
		t.Errorf("unexpected server URL: %s", loaded.ServerURL)
	}
	if loaded.RunnerToken != "beep_rt_saved_token" {
		t.Errorf("unexpected token: %s", loaded.RunnerToken)
	}

	cfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("failed to load effective config: %v", err)
	}
	if cfg.ServerURL != "https://core.internal:8080" {
		t.Errorf("expected server URL from file, got %s", cfg.ServerURL)
	}
	if cfg.RunnerToken != "beep_rt_saved_token" {
		t.Errorf("expected token from file, got %s", cfg.RunnerToken)
	}
	if !cfg.AllowExec {
		t.Errorf("expected allow_exec from file to be true")
	}
	if cfg.Concurrency != 8 {
		t.Errorf("expected concurrency 8, got %d", cfg.Concurrency)
	}
	if cfg.PollInterval != 2*time.Second {
		t.Errorf("expected poll interval 2s, got %v", cfg.PollInterval)
	}
}

func TestMaskToken(t *testing.T) {
	if MaskToken("") != "(not set)" {
		t.Errorf("expected (not set), got %s", MaskToken(""))
	}
	if MaskToken("beep_rt_12345678abcdefgh") != "beep_rt_••••••••" {
		t.Errorf("unexpected masked token: %s", MaskToken("beep_rt_12345678abcdefgh"))
	}
}
