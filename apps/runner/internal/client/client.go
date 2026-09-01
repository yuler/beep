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
	"beep-runner/internal/probe"
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
			Timeout: 35 * time.Second, // Long-polling hold can be up to 30s
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

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ping request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ping failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var res PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode ping response: %w", err)
	}

	return &res, nil
}

type PollResponse struct {
	Task *probe.RunnerTask `json:"task"`
}

func (c *Client) Poll(ctx context.Context) (*probe.RunnerTask, error) {
	url := fmt.Sprintf("%s/api/v1/runner/tasks/poll", c.cfg.ServerURL)
	payload := map[string]any{
		"version":    version.Version,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"hostname":   c.cfg.Hostname,
		"allow_exec": c.cfg.AllowExec,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
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
		return nil, nil // No task due
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

func (c *Client) ReportResult(ctx context.Context, taskID string, signal *probe.Signal) error {
	url := fmt.Sprintf("%s/api/v1/runner/tasks/%s/result", c.cfg.ServerURL, taskID)

	payload := map[string]any{
		"status":  signal.Status,
		"title":   signal.Title,
		"message": signal.Message,
		"metrics": signal.Metrics,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("result upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("result upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Runner-Token", c.cfg.RunnerToken)
	req.Header.Set("User-Agent", fmt.Sprintf("Beep-Runner/%s (%s; %s)", version.Version, runtime.GOOS, runtime.GOARCH))
}
