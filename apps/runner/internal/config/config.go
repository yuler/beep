package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerURL    string
	RunnerToken  string
	AllowExec    bool
	Concurrency  int
	PollInterval time.Duration
	Hostname     string
	Workspace    string
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		ServerURL:    getEnv("BEEP_SERVER", "http://core.localhost:3000"),
		RunnerToken:  getEnv("BEEP_RUNNER_TOKEN", ""),
		AllowExec:    getEnvBool("BEEP_ALLOW_EXEC", false),
		Concurrency:  getEnvInt("BEEP_CONCURRENCY", 5),
		PollInterval: getEnvDuration("BEEP_POLL_INTERVAL", 3*time.Second),
		Hostname:     getEnv("BEEP_HOSTNAME", ""),
		Workspace:    getEnv("BEEP_WORKSPACE", ""),
	}

	if cfg.Hostname == "" {
		if h, err := os.Hostname(); err == nil {
			cfg.Hostname = h
		} else {
			cfg.Hostname = "unknown-host"
		}
	}

	cfg.ServerURL = strings.TrimRight(cfg.ServerURL, "/")

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL is required (set --server or BEEP_SERVER)")
	}
	if c.RunnerToken == "" {
		return fmt.Errorf("runner token is required (set --token or BEEP_RUNNER_TOKEN)")
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 5
	}
	if c.PollInterval < 500*time.Millisecond {
		c.PollInterval = 500 * time.Millisecond
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		v := strings.ToLower(strings.TrimSpace(val))
		return v == "1" || v == "true" || v == "yes" || v == "on"
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(val)); err == nil && d > 0 {
			return d
		}
		if sec, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return defaultVal
}
