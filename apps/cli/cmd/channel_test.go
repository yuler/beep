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
)

func TestChannelCommandsRegistration(t *testing.T) {
	for _, sub := range []string{"connect", "disconnect", "start", "up", "stop", "down", "status"} {
		cmd, _, err := RootCmd.Find([]string{"channel", sub})
		if err != nil {
			t.Fatalf("failed to find 'channel %s': %v", sub, err)
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

func TestDisconnectCommandClearsToken(t *testing.T) {
	var gotMethod, gotPath, gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotToken = r.Header.Get("X-CLI-Token")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("BEEP_SERVER", server.URL)
	t.Setenv("BEEP_SERVER", server.URL)
	t.Setenv("BEEP_CHANNEL_TOKEN", "")
	t.Setenv("BEEP_CLI_TOKEN", "")
	t.Setenv("BEEP_DEVICE_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:    server.URL,
		ChannelToken: "beep_ct_test123",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	cmd := mustFindCmd(t, "channel", "disconnect")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("channel disconnect failed: %v", err)
	}

	if gotMethod != http.MethodDelete || gotPath != "/api/v1/channels/cli/connection" {
		t.Errorf("expected DELETE /api/v1/channels/cli/connection, got %s %s", gotMethod, gotPath)
	}
	if gotToken != "beep_ct_test123" {
		t.Errorf("expected X-CLI-Token header to be sent, got %q", gotToken)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.ChannelToken != "" || updated.CliToken != "" || updated.DeviceToken != "" {
		t.Errorf("expected tokens to be cleared, got channel_token=%q", updated.ChannelToken)
	}
}

func TestDisconnectCommandClearsTokenWhenServerGone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	t.Setenv("BEEP_SERVER", server.URL)
	t.Setenv("BEEP_CHANNEL_TOKEN", "")
	t.Setenv("BEEP_CLI_TOKEN", "")
	t.Setenv("BEEP_DEVICE_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:    server.URL,
		ChannelToken: "beep_ct_test123",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	cmd := mustFindCmd(t, "channel", "disconnect")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("channel disconnect failed: %v", err)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.ChannelToken != "" || updated.CliToken != "" || updated.DeviceToken != "" {
		t.Errorf("expected tokens to be cleared, got channel_token=%q", updated.ChannelToken)
	}
}

func TestDeviceAuthorizationAndPollMockServer(t *testing.T) {
	pollCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/channels/cli/authorizations":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(client.DeviceAuthorizationResponse{
				DeviceCode:              "dev_code_123",
				UserCode:                "TEST-CODE",
				VerificationURI:         "http://example.com/device",
				VerificationURIComplete: "http://example.com/device?code=TEST-CODE",
				ExpiresIn:               900,
				Interval:                1,
			})
		case "/api/v1/channels/cli/authorizations/token":
			pollCount++
			w.Header().Set("Content-Type", "application/json")
			if pollCount == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(client.OAuthErrorResponse{
					ErrorCode:        client.OAuthErrAuthorizationPending,
					ErrorDescription: "pending",
				})
			} else {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(client.DeviceTokenResponse{
					AccessToken: "beep_ct_success_token",
					TokenType:   "bearer",
					ChannelID:   "chan_123",
					ChannelName: "Test-Laptop",
				})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		ServerURL: server.URL,
		Hostname:  "Test-Laptop",
	}
	c := client.New(cfg)

	authRes, err := c.RequestDeviceAuthorization(context.Background(), "Test-Laptop", "test-account")
	if err != nil {
		t.Fatalf("RequestDeviceAuthorization failed: %v", err)
	}
	if authRes.UserCode != "TEST-CODE" {
		t.Errorf("expected user code 'TEST-CODE', got %q", authRes.UserCode)
	}

	_, err = c.PollDeviceToken(context.Background(), authRes.DeviceCode)
	if err == nil {
		t.Fatal("expected pending error on first poll, got nil")
	}
	oauthErr, ok := err.(*client.OAuthErrorResponse)
	if !ok || oauthErr.ErrorCode != client.OAuthErrAuthorizationPending {
		t.Fatalf("expected authorization_pending error, got %v", err)
	}

	tokenRes, err := c.PollDeviceToken(context.Background(), authRes.DeviceCode)
	if err != nil {
		t.Fatalf("expected successful poll, got %v", err)
	}
	if tokenRes.AccessToken != "beep_ct_success_token" {
		t.Errorf("expected access token 'beep_ct_success_token', got %q", tokenRes.AccessToken)
	}
}

func TestChannelConnectAutomaticallyStartsDaemon(t *testing.T) {
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
		case "/api/v1/channels/cli/authorizations":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(client.DeviceAuthorizationResponse{
				DeviceCode:              "dev_code_auto",
				UserCode:                "AUTO-CODE",
				VerificationURI:         "http://example.com/device",
				VerificationURIComplete: "http://example.com/device?code=AUTO-CODE",
				ExpiresIn:               900,
				Interval:                1,
			})
		case "/api/v1/channels/cli/authorizations/token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(client.DeviceTokenResponse{
				AccessToken: "beep_ct_auto_token",
				TokenType:   "bearer",
				ChannelID:   "chan_auto",
				ChannelName: "Auto-Channel",
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

	cmd := mustFindCmd(t, "channel", "connect")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("channel connect failed: %v", err)
	}

	if startedService != "channel" {
		t.Errorf("expected started service 'channel', got %q", startedService)
	}
	if len(startedRawArgs) < 2 || startedRawArgs[0] != "--workspace" || startedRawArgs[1] != tmpDir {
		t.Errorf("expected rawArgs to contain --workspace %s, got %v", tmpDir, startedRawArgs)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.ChannelToken != "beep_ct_auto_token" {
		t.Errorf("expected channel_token 'beep_ct_auto_token', got %q", updated.ChannelToken)
	}
}
