package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"beep-runner/internal/config"
	"beep-runner/internal/task"
	"beep-runner/internal/version"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
	}
}

type PingResponse struct {
	Status     string `json:"status"`
	RunnerID   string `json:"runner_id"`
	RunnerName string `json:"runner_name"`
	ServerTime string `json:"server_time"`
}

func (c *Client) Ping(ctx context.Context) (*PingResponse, error) {
	url := fmt.Sprintf("%s/api/v1/runner/ping", c.cfg.ServerURL)
	payload := map[string]any{
		"version":    version.Version,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"hostname":   c.cfg.Hostname,
		"allow_exec": c.cfg.AllowExec,
	}

	var res PingResponse
	if err := c.postJSON(ctx, url, payload, http.StatusOK, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type PollResponse struct {
	Task *task.Task `json:"task"`
}

func (c *Client) Poll(ctx context.Context) (*task.Task, error) {
	url := fmt.Sprintf("%s/api/v1/runner/tasks/poll", c.cfg.ServerURL)
	payload := map[string]any{
		"version":    version.Version,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"hostname":   c.cfg.Hostname,
		"allow_exec": c.cfg.AllowExec,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, mustJSON(payload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("poll failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var res PollResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode poll response: %w", err)
	}
	return res.Task, nil
}

func (c *Client) ReportLog(ctx context.Context, logURL, chunk string) error {
	if chunk == "" {
		return nil
	}
	url := logURL
	if url == "" {
		return fmt.Errorf("log url is empty")
	}
	return c.postJSON(ctx, url, map[string]any{"chunk": chunk}, http.StatusOK, nil)
}

func (c *Client) ReportResult(ctx context.Context, resultURL string, result *task.Result) error {
	if resultURL == "" {
		return fmt.Errorf("result url is empty")
	}
	payload := map[string]any{
		"status":  result.Status,
		"title":   result.Title,
		"message": result.Message,
		"metrics": result.Metrics,
	}
	return c.postJSON(ctx, resultURL, payload, http.StatusOK, nil)
}

func (c *Client) postJSON(ctx context.Context, url string, payload any, want int, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, mustJSON(payload))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		respBody, _ := io.ReadAll(resp.Body)
		if dest == nil && want == http.StatusOK && resp.StatusCode == http.StatusUnprocessableEntity {
			return nil
		}
		return fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	if dest == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Runner-Token", c.cfg.RunnerToken)
	req.Header.Set("User-Agent", fmt.Sprintf("Beep-Runner/%s (%s; %s)", version.Version, runtime.GOOS, runtime.GOARCH))
}

func mustJSON(payload any) *bytes.Reader {
	bodyBytes, _ := json.Marshal(payload)
	return bytes.NewReader(bodyBytes)
}
