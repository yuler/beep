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
)

func TestChannelCommandsRegistration(t *testing.T) {
	connectCmd, _, err := RootCmd.Find([]string{"channel", "connect"})
	if err != nil {
		t.Fatalf("failed to find 'channel connect': %v", err)
	}
	if connectCmd.Name() != "connect" {
		t.Errorf("expected command name 'connect', got %s", connectCmd.Name())
	}

	disconnectCmd, _, err := RootCmd.Find([]string{"channel", "disconnect"})
	if err != nil {
		t.Fatalf("failed to find 'channel disconnect': %v", err)
	}
	if disconnectCmd.Name() != "disconnect" {
		t.Errorf("expected command name 'disconnect', got %s", disconnectCmd.Name())
	}
}

func TestDisconnectCommandClearsToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	fc := &config.FileConfig{
		ServerURL:   "https://example.com",
		CliToken:    "beep_ct_test123",
		DeviceToken: "beep_ct_test123",
	}
	if err := config.SaveFile(configPath, fc); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	err := channelDisconnectCmd.RunE(channelDisconnectCmd, nil)
	if err != nil {
		t.Fatalf("channel disconnect failed: %v", err)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if updated.CliToken != "" || updated.DeviceToken != "" {
		t.Errorf("expected tokens to be cleared, got cli_token=%q, device_token=%q", updated.CliToken, updated.DeviceToken)
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

	authRes, err := c.RequestDeviceAuthorization(context.Background(), "Test-Laptop")
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
