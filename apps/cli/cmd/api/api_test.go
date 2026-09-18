package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beep/internal/cmdutil"
)

func setupTestWorkspace(t *testing.T, serverURL, token, accountSlug string) string {
	t.Helper()
	dir := t.TempDir()
	cfg := map[string]any{
		"server_url":   serverURL,
		"access_token": token,
		"account_slug": accountSlug,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	cmdutil.SetOverrideWorkspace(dir)
	t.Cleanup(func() {
		cmdutil.SetOverrideWorkspace("")
	})
	return dir
}

func TestAPIGetRequest(t *testing.T) {
	var receivedPath, receivedAuth, receivedAccount, receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedAuth = r.Header.Get("Authorization")
		receivedAccount = r.Header.Get("X-Account-Slug")
		receivedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"beep-1","title":"Hello"}]`))
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "test-token-123", "my-account")

	cmd := NewCmdAPI()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"/api/v1/beeps",
		"-X", "GET",
		"-f", "status=active",
		"-f", "limit=10",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedPath != "/api/v1/beeps" {
		t.Errorf("got path %q, want %q", receivedPath, "/api/v1/beeps")
	}
	if receivedAuth != "Bearer test-token-123" {
		t.Errorf("got auth %q, want 'Bearer test-token-123'", receivedAuth)
	}
	if receivedAccount != "my-account" {
		t.Errorf("got account %q, want 'my-account'", receivedAccount)
	}
	if !strings.Contains(receivedQuery, "status=active") || !strings.Contains(receivedQuery, "limit=10") {
		t.Errorf("unexpected query: %s", receivedQuery)
	}

	if !strings.Contains(out.String(), `"id":"beep-1"`) {
		t.Errorf("unexpected output: %s", out.String())
	}
}

func TestAPIPostInferenceAndBody(t *testing.T) {
	var receivedMethod string
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		bodyBytes, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(bodyBytes, &receivedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"beep-created","title":"New Beep"}`))
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "token-xyz", "personal")

	cmd := NewCmdAPI()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"api/v1/beeps", // relative without leading slash
		"-f", "title=New Beep",
		"-f", "body=Meeting at 4pm",
		"--raw-field", "count=42",
		"--raw-field", "active=true",
		"--raw-field", `metadata={"team":"core"}`,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedMethod != "POST" {
		t.Errorf("got method %q, want POST (inferred)", receivedMethod)
	}
	if receivedBody["title"] != "New Beep" {
		t.Errorf("got title %v, want 'New Beep'", receivedBody["title"])
	}
	if receivedBody["body"] != "Meeting at 4pm" {
		t.Errorf("got body %v, want 'Meeting at 4pm'", receivedBody["body"])
	}
	if count, ok := receivedBody["count"].(float64); !ok || count != 42 {
		t.Errorf("got count %v, want 42", receivedBody["count"])
	}
	if active, ok := receivedBody["active"].(bool); !ok || !active {
		t.Errorf("got active %v, want true", receivedBody["active"])
	}
	if meta, ok := receivedBody["metadata"].(map[string]any); !ok || meta["team"] != "core" {
		t.Errorf("got metadata %v, want map with team=core", receivedBody["metadata"])
	}
}

func TestAPICustomHeaderAndInclude(t *testing.T) {
	var receivedCustomHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCustomHeader = r.Header.Get("X-Custom-Beep")
		w.Header().Set("X-Server-Time", "2026-09-18T00:00:00Z")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "token", "slug")

	cmd := NewCmdAPI()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"/api/v1/beeps",
		"-H", "X-Custom-Beep: hello-world",
		"-i",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedCustomHeader != "hello-world" {
		t.Errorf("got custom header %q, want 'hello-world'", receivedCustomHeader)
	}

	output := out.String()
	if !strings.Contains(output, "HTTP/1.1 200 OK") {
		t.Errorf("expected HTTP status line with -i, got: %s", output)
	}
	if !strings.Contains(output, "X-Server-Time: 2026-09-18T00:00:00Z") {
		t.Errorf("expected header in output with -i, got: %s", output)
	}
	if !strings.Contains(output, `{"ok":true}`) {
		t.Errorf("expected body in output, got: %s", output)
	}
}

func TestAPIInputFileAndStdin(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		receivedBody = string(data)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "token", "slug")

	// 1. From input file
	tmpFile := filepath.Join(t.TempDir(), "payload.json")
	payloadContent := `{"custom":"payload","num":123}`
	if err := os.WriteFile(tmpFile, []byte(payloadContent), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmdAPI()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"/api/v1/beeps",
		"--input", tmpFile,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedBody != payloadContent {
		t.Errorf("got body %q, want %q", receivedBody, payloadContent)
	}

	// 2. From stdin
	cmdStdin := NewCmdAPI()
	out.Reset()
	cmdStdin.SetOut(&out)
	cmdStdin.SetIn(strings.NewReader(`{"stdin":true}`))
	cmdStdin.SetArgs([]string{
		"/api/v1/beeps",
		"--input", "-",
	})

	if err := cmdStdin.Execute(); err != nil {
		t.Fatalf("Execute with stdin failed: %v", err)
	}

	if receivedBody != `{"stdin":true}` {
		t.Errorf("got body from stdin %q, want '{\"stdin\":true}'", receivedBody)
	}
}

func TestAPIErrorStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"resource not found"}`))
	}))
	defer server.Close()

	setupTestWorkspace(t, server.URL, "token", "slug")

	cmd := NewCmdAPI()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"/api/v1/nonexistent",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for 404 response, got nil")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
	if !strings.Contains(out.String(), "resource not found") {
		t.Errorf("expected error body in stdout, got: %s", out.String())
	}
}

func TestAPIAbsoluteURLRejected(t *testing.T) {
	setupTestWorkspace(t, "http://127.0.0.1:3000", "token", "slug")

	for _, url := range []string{"https://evil.com/api", "http://evil.com/api"} {
		cmd := NewCmdAPI()
		cmd.SetArgs([]string{url})
		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected error for absolute URL %q, got nil", url)
		}
		if !strings.Contains(err.Error(), "absolute URLs are not allowed") {
			t.Errorf("expected 'absolute URLs are not allowed' in error, got: %v", err)
		}
	}
}

func TestAPINoRunnerTokenReuse(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Configure workspace with runner_token ONLY (no access_token)
	dir := t.TempDir()
	cfg := map[string]any{
		"server_url":   server.URL,
		"runner_token": "runner-secret-token",
		"account_slug": "slug",
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600)
	cmdutil.SetOverrideWorkspace(dir)
	defer cmdutil.SetOverrideWorkspace("")

	cmd := NewCmdAPI()
	cmd.SetArgs([]string{"/api/v1/test"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedAuth != "" {
		t.Errorf("expected Authorization header to be empty when only runner_token exists, got %q", receivedAuth)
	}
}
