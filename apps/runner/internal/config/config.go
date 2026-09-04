package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileConfig struct {
	ServerURL    string `json:"server_url,omitempty"`
	RunnerToken  string `json:"runner_token,omitempty"`
	Workspace    string `json:"workspace,omitempty"`
	Concurrency  int    `json:"concurrency,omitempty"`
	PollInterval string `json:"poll_interval,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
}

type Config struct {
	ServerURL    string
	RunnerToken  string
	Concurrency  int
	PollInterval time.Duration
	Hostname     string
	Workspace    string
	ConfigFile   string
}

func DefaultWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".beep-runner"
	}
	return filepath.Join(home, ".beep-runner")
}

func GetConfigPath(ws string) string {
	if ws == "" {
		ws = getEnv("BEEP_WORKSPACE", "")
	}
	if ws == "" {
		ws = DefaultWorkspace()
	}
	if strings.HasPrefix(ws, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			ws = filepath.Join(home, strings.TrimPrefix(ws, "~"))
		}
	}
	abs, err := filepath.Abs(ws)
	if err == nil {
		ws = abs
	}
	return filepath.Join(ws, "config.json")
}

func LoadFile(configPath string) (*FileConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		return nil, err
	}
	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("invalid config file %s: %w", configPath, err)
	}
	return &fc, nil
}

func SaveFile(configPath string, fc *FileConfig) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(data, '\n'), 0o600)
}

func Load(wsHint string) (*Config, error) {
	configPath := GetConfigPath(wsHint)
	fc, err := LoadFile(configPath)
	if err != nil {
		// Log or ignore unparseable non-fatal config file error and fallback
		fc = &FileConfig{}
	}

	ws := wsHint
	if ws == "" {
		ws = getEnv("BEEP_WORKSPACE", fc.Workspace)
	}
	if ws == "" {
		ws = DefaultWorkspace()
	}

	serverURL := getEnv("BEEP_SERVER", fc.ServerURL)
	if serverURL == "" {
		serverURL = "http://core.localhost:3000"
	}
	serverURL = strings.TrimRight(serverURL, "/")

	runnerToken := getEnv("BEEP_RUNNER_TOKEN", fc.RunnerToken)

	concurrency := 5
	if fc.Concurrency > 0 {
		concurrency = fc.Concurrency
	}
	concurrency = getEnvInt("BEEP_CONCURRENCY", concurrency)

	pollInterval := 3 * time.Second
	if fc.PollInterval != "" {
		if d, err := time.ParseDuration(fc.PollInterval); err == nil && d > 0 {
			pollInterval = d
		}
	}
	pollInterval = getEnvDuration("BEEP_POLL_INTERVAL", pollInterval)

	hostname := getEnv("BEEP_HOSTNAME", fc.Hostname)
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "unknown-host"
		}
	}

	cfg := &Config{
		ServerURL:    serverURL,
		RunnerToken:  runnerToken,
		Concurrency:  concurrency,
		PollInterval: pollInterval,
		Hostname:     hostname,
		Workspace:    ws,
		ConfigFile:   configPath,
	}

	return cfg, nil
}

func LoadFromEnv() (*Config, error) {
	return Load("")
}

func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL is required (set via 'beep-runner config set --server <url>' or --server or BEEP_SERVER)")
	}
	if c.RunnerToken == "" {
		return fmt.Errorf("runner token is required (set via 'beep-runner config set --token <token>' or --token or BEEP_RUNNER_TOKEN)")
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 5
	}
	if c.PollInterval < 500*time.Millisecond {
		c.PollInterval = 500 * time.Millisecond
	}
	return nil
}

func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 12 {
		return "••••••••"
	}
	prefix := token[:8]
	return prefix + "••••••••"
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
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
