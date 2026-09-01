package config

import (
	"os"
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

func TestConfigValidateMissingToken(t *testing.T) {
	cfg := &Config{
		ServerURL: "https://beep.example.com",
	}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error when token is missing")
	}
}
