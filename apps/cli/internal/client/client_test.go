package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"beep/internal/config"
	"beep/internal/task"
)

func TestClientPing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runner/ping" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"runner_id":   "run_123",
			"runner_name": "My-Runner",
			"server_time": "2026-09-01T12:00:00Z",
		})
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_rt_test"})
	res, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected ping error: %v", err)
	}
	if res.RunnerID != "run_123" {
		t.Errorf("expected runner_id run_123, got %s", res.RunnerID)
	}
}

func TestClientPollLogAndResult(t *testing.T) {
	var gotLog, gotResult bool
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/runner/tasks":
			json.NewEncoder(w).Encode(map[string]any{
				"task": map[string]any{
					"id":              "task-1",
					"job_slug":        "intranet-http",
					"name":            "Intranet HTTP",
					"timeout_seconds": 15,
					"log_url":         tsURL + "/api/v1/runner/tasks/task-1/logs",
					"result_url":      tsURL + "/api/v1/runner/tasks/task-1/result",
					"config":          map[string]any{"target_url": "http://10.0.0.5"},
				},
			})
		case "/api/v1/runner/tasks/task-1/logs":
			gotLog = true
			w.WriteHeader(http.StatusNoContent)
		case "/api/v1/runner/tasks/task-1/result":
			gotResult = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()
	tsURL = ts.URL

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_rt_test"})
	job, err := c.Poll(context.Background())
	if err != nil || job == nil || job.JobSlug != "intranet-http" {
		t.Fatalf("poll: %v %#v", err, job)
	}

	if err := c.ReportLog(context.Background(), job.LogURL, "hello\n"); err != nil {
		t.Fatal(err)
	}
	if err := c.ReportResult(context.Background(), job.ResultURL, task.Ok("ok", "", nil)); err != nil {
		t.Fatal(err)
	}
	if !gotLog || !gotResult {
		t.Fatalf("expected log and result posts, got log=%v result=%v", gotLog, gotResult)
	}
}

func TestReportRejectsForeignCallbackURL(t *testing.T) {
	c := New(&config.Config{ServerURL: "https://core.example.com", RunnerToken: "beep_rt_test"})
	if err := c.ReportLog(context.Background(), "https://evil.example/api/v1/runner/tasks/x/logs", "leak\n"); err == nil {
		t.Fatal("expected foreign log URL to be rejected")
	}
	if err := c.ReportResult(context.Background(), "https://core.example.com/not-a-task", task.Ok("ok", "", nil)); err == nil {
		t.Fatal("expected non-task path to be rejected")
	}
}

func TestReportLogRejectsUnprocessableEntity(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Run is no longer accepting logs"}`))
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_rt_test"})
	err := c.ReportLog(context.Background(), ts.URL+"/api/v1/runner/tasks/task-1/logs", "hello\n")
	if err == nil {
		t.Fatal("expected 422 log report to fail")
	}
}

func TestCrossHostRedirectStripsRunnerToken(t *testing.T) {
	var receivedTokenOnTarget string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTokenOnTarget = r.Header.Get("X-Runner-Token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobs": []any{}})
	}))
	defer targetServer.Close()

	originServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL+"/api/v1/runner/jobs", http.StatusFound)
	}))
	defer originServer.Close()

	c := New(&config.Config{ServerURL: originServer.URL, RunnerToken: "secret_runner_token"})
	_, err := c.ListJobs(context.Background())
	if err != nil {
		t.Fatalf("unexpected list jobs error: %v", err)
	}
	if receivedTokenOnTarget != "" {
		t.Fatalf("expected X-Runner-Token to be stripped on cross-host redirect, but got: %q", receivedTokenOnTarget)
	}
}

func TestClientDeleteJobNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Job not found"}`))
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL, RunnerToken: "beep_rt_test"})
	err := c.DeleteJob(context.Background(), "non-existent-job")
	if err == nil {
		t.Fatal("expected 404 to return error, got nil")
	}
}

func TestClientGetMe(t *testing.T) {
	var gotAuth, gotAccount string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		gotAccount = r.Header.Get("X-Account-Slug")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"identity": map[string]any{
				"id":    "usr_123",
				"email": "user@example.com",
				"name":  "Test User",
				"staff": false,
			},
			"accounts": []map[string]any{
				{
					"id":       "acc_123",
					"name":     "Personal",
					"slug":     "test-user",
					"personal": true,
				},
			},
			"last_account_slug": "test-user",
		})
	}))
	defer ts.Close()

	c := New(&config.Config{
		ServerURL:   ts.URL,
		AccessToken: "beep_pat_secret",
		AccountSlug: "test-user",
	})
	me, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected GetMe error: %v", err)
	}
	if gotAuth != "Bearer beep_pat_secret" {
		t.Errorf("expected Authorization header 'Bearer beep_pat_secret', got %q", gotAuth)
	}
	if gotAccount != "test-user" {
		t.Errorf("expected X-Account-Slug header 'test-user', got %q", gotAccount)
	}
	if me.Identity.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", me.Identity.Email)
	}
	if me.LastAccountSlug != "test-user" {
		t.Errorf("expected last_account_slug test-user, got %s", me.LastAccountSlug)
	}
}

func TestClientCliDeviceFlow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/cli/authorizations":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"device_code":               "dc_123",
				"user_code":                 "ABCD-EFGH",
				"verification_uri":          "http://example.com/device/cli",
				"verification_uri_complete": "http://example.com/device/cli?code=ABCD-EFGH",
				"expires_in":                900,
				"interval":                  5,
			})
		case "/api/v1/cli/authorizations/token":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "beep_pat_new_token",
				"token_type":   "bearer",
				"user": map[string]any{
					"id":    "usr_123",
					"email": "user@example.com",
					"name":  "Test User",
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL})
	authRes, err := c.RequestCliDeviceAuthorization(context.Background(), "My Machine")
	if err != nil {
		t.Fatalf("unexpected RequestCliDeviceAuthorization error: %v", err)
	}
	if authRes.UserCode != "ABCD-EFGH" {
		t.Errorf("expected user code ABCD-EFGH, got %s", authRes.UserCode)
	}

	tokenRes, err := c.PollCliDeviceToken(context.Background(), authRes.DeviceCode)
	if err != nil {
		t.Fatalf("unexpected PollCliDeviceToken error: %v", err)
	}
	if tokenRes.AccessToken != "beep_pat_new_token" {
		t.Errorf("expected access token beep_pat_new_token, got %s", tokenRes.AccessToken)
	}
	if tokenRes.User.Email != "user@example.com" {
		t.Errorf("expected user email user@example.com, got %s", tokenRes.User.Email)
	}
}

func TestParseAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/errors-array":
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    "VALIDATION_ERROR",
				"message": "Title can't be blank and Cron is invalid",
				"errors":  []string{"Title can't be blank", "Cron is invalid"},
			})
		case "/api/errors-map":
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    "VALIDATION_ERROR",
				"message": "Validation failed",
				"errors": map[string][]string{
					"title": {"can't be blank"},
				},
			})
		case "/api/message-only":
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Bad request parameter",
			})
		}
	}))
	defer ts.Close()

	c := New(&config.Config{ServerURL: ts.URL})

	// 1. Errors array
	resp1, err := c.httpClient.Get(ts.URL + "/api/errors-array")
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()
	err1 := parseAPIError(resp1)
	errList1 := ExtractErrorList(err1)
	if len(errList1) != 2 || errList1[0] != "Title can't be blank" || errList1[1] != "Cron is invalid" {
		t.Errorf("unexpected error list from array: %v", errList1)
	}

	// 2. Errors map
	resp2, err := c.httpClient.Get(ts.URL + "/api/errors-map")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	err2 := parseAPIError(resp2)
	errList2 := ExtractErrorList(err2)
	if len(errList2) != 1 || errList2[0] != "title can't be blank" {
		t.Errorf("unexpected error list from map: %v", errList2)
	}

	// 3. Message only
	resp3, err := c.httpClient.Get(ts.URL + "/api/message-only")
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	err3 := parseAPIError(resp3)
	errList3 := ExtractErrorList(err3)
	if len(errList3) != 1 || errList3[0] != "Bad request parameter" {
		t.Errorf("unexpected error list from message only: %v", errList3)
	}
}
