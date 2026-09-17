package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/client"
	"beep/internal/cmdutil"
	"beep/internal/config"
)

func TestAuthCommandsRegistration(t *testing.T) {
	for _, sub := range []string{"login", "logout", "status", "whoami"} {
		cmd, _, err := RootCmd.Find([]string{"auth", sub})
		if err != nil {
			t.Fatalf("failed to find 'auth %s': %v", sub, err)
		}
		expectedName := sub
		if sub == "whoami" {
			expectedName = "status"
		}
		if cmd.Name() != expectedName {
			t.Errorf("expected command name %q, got %q", expectedName, cmd.Name())
		}
	}
}

func TestAuthLogoutClearsSessionOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beep-auth-logout-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	initial := &config.FileConfig{
		ServerURL:    "http://example.com",
		AccessToken:  "beep_pat_test123",
		UserEmail:    "test@example.com",
		UserName:     "Test User",
		RunnerToken:  "beep_rt_keep_this",
		ChannelToken: "beep_ct_keep_this",
		AccountSlug:  "my-team",
	}
	if err := config.SaveFile(configPath, initial); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	flagWorkspace = tmpDir
	defer func() { flagWorkspace = "" }()

	if err := authLogoutCmd.RunE(authLogoutCmd, nil); err != nil {
		t.Fatalf("auth logout failed: %v", err)
	}

	updated, err := config.LoadFile(configPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updated.AccessToken != "" {
		t.Errorf("expected AccessToken to be cleared, got %q", updated.AccessToken)
	}
	if updated.UserEmail != "" {
		t.Errorf("expected UserEmail to be cleared, got %q", updated.UserEmail)
	}
	if updated.UserName != "" {
		t.Errorf("expected UserName to be cleared, got %q", updated.UserName)
	}
	// Runner and channel tokens must remain intact!
	if updated.RunnerToken != "beep_rt_keep_this" {
		t.Errorf("expected RunnerToken to be preserved, got %q", updated.RunnerToken)
	}
	if updated.ChannelToken != "beep_ct_keep_this" {
		t.Errorf("expected ChannelToken to be preserved, got %q", updated.ChannelToken)
	}
	if updated.AccountSlug != "my-team" {
		t.Errorf("expected AccountSlug to be preserved, got %q", updated.AccountSlug)
	}
}

func TestRunnerConnectRequiresLogin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beep-runner-login-check-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	initial := &config.FileConfig{
		ServerURL: "http://example.com",
		// No AccessToken!
	}
	if err := config.SaveFile(configPath, initial); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	flagWorkspace = tmpDir
	flagNoInteractive = true
	defer func() {
		flagWorkspace = ""
		flagNoInteractive = false
	}()

	cmd := mustFindCmd(t, "runner", "connect")
	err = cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected runner connect to fail without login, but got nil")
	}
	if !strings.Contains(err.Error(), "beep auth login") {
		t.Errorf("expected error mentioning 'beep auth login', got: %v", err)
	}
}

func TestChannelConnectRequiresLogin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beep-channel-login-check-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	initial := &config.FileConfig{
		ServerURL: "http://example.com",
		// No AccessToken!
	}
	if err := config.SaveFile(configPath, initial); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	flagWorkspace = tmpDir
	flagNoInteractive = true
	defer func() {
		flagWorkspace = ""
		flagNoInteractive = false
	}()

	cmd := mustFindCmd(t, "channel", "connect")
	err = cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected channel connect to fail without login, but got nil")
	}
	if !strings.Contains(err.Error(), "beep auth login") {
		t.Errorf("expected error mentioning 'beep auth login', got: %v", err)
	}
}

func TestRunnerConnectSucceedsPastLoginWhenLoggedIn(t *testing.T) {
	var gotMeAuth string
	var gotRunnerAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/me":
			gotMeAuth = r.Header.Get("Authorization")
			json.NewEncoder(w).Encode(map[string]any{
				"identity": map[string]any{
					"id":    "usr_123",
					"email": "user@example.com",
					"name":  "User",
				},
				"accounts": []map[string]any{
					{"id": "acc_1", "slug": "test-account", "name": "Test"},
				},
				"last_account_slug": "test-account",
			})
		case "/api/v1/runners/authorizations":
			gotRunnerAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"device_code":               "dc_123",
				"user_code":                 "TEST-1234",
				"verification_uri":          "http://example.com/test",
				"verification_uri_complete": "http://example.com/test?code=TEST-1234",
				"expires_in":                1,
				"interval":                  1,
			})
		case "/api/v1/runners/authorizations/token":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "expired_token",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "beep-runner-login-success-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	initial := &config.FileConfig{
		ServerURL:   server.URL,
		AccessToken: "beep_pat_logged_in",
	}
	if err := config.SaveFile(configPath, initial); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	flagWorkspace = tmpDir
	flagNoInteractive = true
	defer func() {
		flagWorkspace = ""
		flagNoInteractive = false
	}()

	cmd := mustFindCmd(t, "runner", "connect")
	_ = cmd.RunE(cmd, nil)

	if gotMeAuth != "Bearer beep_pat_logged_in" {
		t.Errorf("expected /api/v1/me to be called with 'Bearer beep_pat_logged_in', got %q", gotMeAuth)
	}
	if gotRunnerAuth != "Bearer beep_pat_logged_in" {
		t.Errorf("expected /api/v1/runners/authorizations to be called with 'Bearer beep_pat_logged_in', got %q", gotRunnerAuth)
	}
}

func TestResolveAccountSlug(t *testing.T) {
	// 1. Explicit flag/env takes priority
	slug, err := cmdutil.ResolveAccountSlug(nil, "my-explicit-slug", "")
	if err != nil || slug != "my-explicit-slug" {
		t.Errorf("expected my-explicit-slug, got %s (err: %v)", slug, err)
	}

	// 2. Configured account takes precedence when no explicit flag
	slug, err = cmdutil.ResolveAccountSlug(nil, "", "my-cfg-slug")
	if err != nil || slug != "my-cfg-slug" {
		t.Errorf("expected my-cfg-slug, got %s (err: %v)", slug, err)
	}

	// 3. Single account auto-selection
	singleMe := &client.MeResponse{
		Accounts: []client.MeAccount{
			{ID: "1", Name: "Personal", Slug: "personal-slug", Personal: true},
		},
	}
	slug, err = cmdutil.ResolveAccountSlug(singleMe, "", "")
	if err != nil || slug != "personal-slug" {
		t.Errorf("expected auto-selected personal-slug, got %s (err: %v)", slug, err)
	}

	// 4. Non-interactive with multiple accounts returns error requiring explicit account
	multiMe := &client.MeResponse{
		Accounts: []client.MeAccount{
			{ID: "1", Name: "Personal", Slug: "personal-slug", Personal: true},
			{ID: "2", Name: "Team", Slug: "team-slug", Personal: false},
		},
	}
	_, err = cmdutil.ResolveAccountSlug(multiMe, "", "")
	if err == nil {
		t.Fatal("expected error in non-interactive mode with multiple accounts, got nil")
	}
}
