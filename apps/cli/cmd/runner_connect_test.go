package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/service"

	"github.com/spf13/cobra"
)

func TestRunnerCommandsRegistration(t *testing.T) {
	for _, sub := range []string{"connect", "disconnect", "start", "up", "stop", "down", "status"} {
		cmd, _, err := RootCmd.Find([]string{"runner", sub})
		if err != nil {
			t.Fatalf("failed to find 'runner %s': %v", sub, err)
		}
		expectedName := sub
		if sub == "up" {
			expectedName = "start"
		}
		if sub == "down" {
			expectedName = "stop"
		}
		if cmd.Name() != expectedName {
			t.Errorf("expected command name %q, got %q", expectedName, cmd.Name())
		}
	}
}

func TestAllCommandsFlagsNoConflict(t *testing.T) {
	var checkCmd func(c *cobra.Command)
	checkCmd = func(c *cobra.Command) {
		c.InitDefaultHelpFlag()
		for _, child := range c.Commands() {
			checkCmd(child)
		}
	}
	checkCmd(RootCmd)

	cmd, _, err := RootCmd.Find([]string{"runner", "connect"})
	if err != nil {
		t.Fatalf("failed to find 'runner connect': %v", err)
	}
	if cmd.Flags().Lookup("tags") == nil {
		t.Fatal("expected 'tags' flag to be registered on runner connect")
	}
}

func TestRunnerDisconnectCommandClearsToken(t *testing.T) {
	var gotMethod, gotPath, gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotToken = r.Header.Get("X-Runner-Token")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BEEP_SERVER", server.URL)
	t.Setenv("BEEP_RUNNER_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:   server.URL,
		RunnerToken: "beep_rt_test123",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	cmd, _, err := RootCmd.Find([]string{"runner", "disconnect"})
	if err != nil {
		t.Fatalf("failed to find 'runner disconnect': %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("runner disconnect failed: %v", err)
	}

	if gotMethod != http.MethodDelete || gotPath != "/api/v1/runner/connection" {
		t.Errorf("expected DELETE /api/v1/runner/connection, got %s %s", gotMethod, gotPath)
	}
	if gotToken != "beep_rt_test123" {
		t.Errorf("expected X-Runner-Token header to be sent, got %q", gotToken)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.RunnerToken != "" {
		t.Errorf("expected runner_token to be cleared, got runner_token=%q", updated.RunnerToken)
	}
}

func TestClientRunnerDeviceAuthorizationFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/runners/authorizations":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":               "dc_runner_123",
				"user_code":                 "WXYZ-9876",
				"verification_uri":          "http://localhost:5173/device/runner",
				"verification_uri_complete": "http://localhost:5173/device/runner?code=WXYZ-9876",
				"expires_in":                900,
				"interval":                  5,
			})
		case "/api/v1/runners/authorizations/token":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["device_code"] == "dc_runner_123" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"access_token": "beep_rt_success456",
					"token_type":   "bearer",
					"runner_id":    "runner-uuid-1",
					"runner_name":  "My Test Runner",
					"tags":         []string{"test", "ci"},
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":             "invalid_grant",
					"error_description": "Invalid device code",
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		ServerURL: server.URL,
	}
	c := client.New(cfg)

	authRes, err := c.RequestRunnerDeviceAuthorization(context.Background(), "My Test Runner", []string{"test"}, map[string]string{"os": "linux"}, "test-account")
	if err != nil {
		t.Fatalf("RequestRunnerDeviceAuthorization failed: %v", err)
	}
	if authRes.DeviceCode != "dc_runner_123" || authRes.UserCode != "WXYZ-9876" {
		t.Fatalf("unexpected auth response: %+v", authRes)
	}

	tokenRes, err := c.PollRunnerDeviceToken(context.Background(), authRes.DeviceCode)
	if err != nil {
		t.Fatalf("PollRunnerDeviceToken failed: %v", err)
	}
	if tokenRes.AccessToken != "beep_rt_success456" || tokenRes.RunnerName != "My Test Runner" {
		t.Fatalf("unexpected token response: %+v", tokenRes)
	}
}

func TestRunnerConnectAutomaticallyStartsDaemon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/me":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"email": "test@example.com",
				"accounts": []map[string]any{
					{"slug": "test-account", "name": "Test Account"},
				},
				"last_account_slug": "test-account",
			})
		case "/api/v1/runners/authorizations":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(client.DeviceAuthorizationResponse{
				DeviceCode:              "dc_runner_auto",
				UserCode:                "RUNNER-AUTO",
				VerificationURI:         "http://example.com/device/runner",
				VerificationURIComplete: "http://example.com/device/runner?code=RUNNER-AUTO",
				ExpiresIn:               900,
				Interval:                1,
			})
		case "/api/v1/runners/authorizations/token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(client.RunnerTokenResponse{
				AccessToken: "beep_rt_auto_token",
				TokenType:   "bearer",
				RunnerID:    "run_auto",
				RunnerName:  "Auto-Runner",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:   server.URL,
		AccessToken: "test-login-token",
		AccountSlug: "test-account",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	origStart := service.StartServiceDaemonFn
	defer func() { service.StartServiceDaemonFn = origStart }()

	var startedService string
	var startedRawArgs []string
	service.StartServiceDaemonFn = func(service string, rawArgs []string, c *config.Config) error {
		startedService = service
		startedRawArgs = rawArgs
		return nil
	}

	cmd, _, err := RootCmd.Find([]string{"runner", "connect"})
	if err != nil {
		t.Fatalf("failed to find 'runner connect': %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("runner connect failed: %v", err)
	}

	if startedService != "runner" {
		t.Errorf("expected started service 'runner', got %q", startedService)
	}
	if len(startedRawArgs) < 2 || startedRawArgs[0] != "--workspace" || startedRawArgs[1] != tmpDir {
		t.Errorf("expected rawArgs to contain --workspace %s, got %v", tmpDir, startedRawArgs)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.RunnerToken != "beep_rt_auto_token" {
		t.Errorf("expected runner_token 'beep_rt_auto_token', got %q", updated.RunnerToken)
	}
}
