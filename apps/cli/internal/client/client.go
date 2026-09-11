package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"beep/internal/config"
	"beep/internal/task"
	"beep/internal/version"
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
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				if len(via) > 0 && !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
					req.Header.Del("X-Runner-Token")
				}
				return nil
			},
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
		"version":  version.Version,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"hostname": c.cfg.Hostname,
	}

	var res PingResponse
	if err := c.postJSON(ctx, url, payload, http.StatusOK, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type ServerJob struct {
	ID             string         `json:"id"`
	RunnerID       string         `json:"runner_id"`
	Name           string         `json:"name"`
	Slug           string         `json:"slug"`
	Cron           string         `json:"cron"`
	Timezone       string         `json:"timezone"`
	Status         string         `json:"status"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	Config         map[string]any `json:"config,omitempty"`
	NextRunAt      string         `json:"next_run_at,omitempty"`
	LastRunAt      string         `json:"last_run_at,omitempty"`
}

type CreateJobRequest struct {
	ID             string         `json:"id,omitempty"`
	Slug           string         `json:"slug"`
	Name           string         `json:"name,omitempty"`
	Cron           string         `json:"cron,omitempty"`
	Timezone       string         `json:"timezone,omitempty"`
	TimeoutSeconds int            `json:"timeout_seconds,omitempty"`
	Description    string         `json:"description,omitempty"`
	Config         map[string]any `json:"config,omitempty"`
}

func (c *Client) CreateJob(ctx context.Context, jobReq *CreateJobRequest) (*ServerJob, error) {
	url := fmt.Sprintf("%s/api/v1/runner/jobs", c.cfg.ServerURL)
	var res struct {
		Job *ServerJob `json:"job"`
	}
	if err := c.postJSON(ctx, url, jobReq, http.StatusCreated, &res); err != nil {
		return nil, err
	}
	return res.Job, nil
}

func (c *Client) PushJobs(ctx context.Context, jobs []*CreateJobRequest) ([]*ServerJob, error) {
	url := fmt.Sprintf("%s/api/v1/runner/jobs/push", c.cfg.ServerURL)
	payload := map[string]any{
		"jobs": jobs,
	}
	var res struct {
		Status      string       `json:"status"`
		PushedCount int          `json:"pushed_count"`
		Jobs        []*ServerJob `json:"jobs"`
	}
	if err := c.postJSON(ctx, url, payload, http.StatusOK, &res); err != nil {
		return nil, err
	}
	return res.Jobs, nil
}

func (c *Client) ListJobs(ctx context.Context) ([]*ServerJob, error) {
	url := fmt.Sprintf("%s/api/v1/runner/jobs", c.cfg.ServerURL)
	var res struct {
		Jobs []*ServerJob `json:"jobs"`
	}
	if err := c.getJSON(ctx, url, http.StatusOK, &res); err != nil {
		return nil, err
	}
	return res.Jobs, nil
}

func (c *Client) DeleteJob(ctx context.Context, slug string) error {
	url := fmt.Sprintf("%s/api/v1/runner/jobs/%s", c.cfg.ServerURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("job not found on server (404)")
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type PollResponse struct {
	Task *task.Task `json:"task"`
}

func (c *Client) Poll(ctx context.Context) (*task.Task, error) {
	url := fmt.Sprintf("%s/api/v1/runner/tasks", c.cfg.ServerURL)
	payload := map[string]any{
		"version":  version.Version,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"hostname": c.cfg.Hostname,
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
	if err := c.allowedCallbackURL(logURL); err != nil {
		return err
	}
	return c.postJSON(ctx, logURL, map[string]any{"chunk": chunk}, http.StatusNoContent, nil)
}

func (c *Client) ReportResult(ctx context.Context, resultURL string, result *task.Result) error {
	if err := c.allowedCallbackURL(resultURL); err != nil {
		return err
	}
	payload := map[string]any{
		"status":  result.Status,
		"title":   result.Title,
		"message": result.Message,
		"metrics": result.Metrics,
	}
	return c.postJSON(ctx, resultURL, payload, http.StatusNoContent, nil)
}

func (c *Client) allowedCallbackURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("callback url is empty")
	}
	target, err := url.Parse(raw)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return fmt.Errorf("invalid callback url")
	}
	base, err := url.Parse(c.cfg.ServerURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return fmt.Errorf("invalid server URL")
	}
	if !strings.EqualFold(target.Scheme, base.Scheme) || !strings.EqualFold(target.Host, base.Host) {
		return fmt.Errorf("callback url %s does not match server %s", raw, c.cfg.ServerURL)
	}
	if !strings.HasPrefix(target.Path, "/api/v1/runner/tasks/") {
		return fmt.Errorf("callback url path is not a runner task endpoint")
	}
	return nil
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
		return fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	if dest == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *Client) getJSON(ctx context.Context, url string, want int, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	if c.cfg.RunnerToken != "" {
		req.Header.Set("X-Runner-Token", c.cfg.RunnerToken)
	}
	if c.cfg.CliToken != "" {
		req.Header.Set("X-CLI-Token", c.cfg.CliToken)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("Beep-Runner/%s (%s; %s)", version.Version, runtime.GOOS, runtime.GOARCH))
}

type CliDelivery struct {
	ID        string         `json:"id"`
	BeepRunID *string        `json:"beep_run_id"`
	Status    string         `json:"status"`
	Payload   map[string]any `json:"payload"`
	ExpiresAt *time.Time     `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
}

type DeviceDelivery = CliDelivery

type CliInboxResponse struct {
	Deliveries []CliDelivery `json:"deliveries"`
}

type DeviceInboxResponse = CliInboxResponse

func (c *Client) FetchCliInbox(ctx context.Context) ([]CliDelivery, error) {
	if c.cfg.CliToken == "" {
		return nil, nil
	}
	url := fmt.Sprintf("%s/api/v1/channels/cli/inbox", c.cfg.ServerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("X-CLI-Token", c.cfg.CliToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch cli inbox failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var inboxRes CliInboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&inboxRes); err != nil {
		return nil, err
	}
	return inboxRes.Deliveries, nil
}

func (c *Client) FetchDeviceInbox(ctx context.Context) ([]CliDelivery, error) {
	return c.FetchCliInbox(ctx)
}

func (c *Client) AckCliDelivery(ctx context.Context, deliveryID string, status string, errorMsg string) error {
	url := fmt.Sprintf("%s/api/v1/channels/cli/deliveries/%s/ack", c.cfg.ServerURL, deliveryID)
	payload := map[string]any{
		"status": status,
	}
	if errorMsg != "" {
		payload["error"] = errorMsg
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, mustJSON(payload))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	req.Header.Set("X-CLI-Token", c.cfg.CliToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ack cli delivery failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (c *Client) AckDeviceDelivery(ctx context.Context, deliveryID string, status string, errorMsg string) error {
	return c.AckCliDelivery(ctx, deliveryID, status, errorMsg)
}

func (c *Client) DisconnectChannel(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/channels/cli/disconnect", c.cfg.ServerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Treat already-gone as success so disconnect stays idempotent
	// (e.g. channel was deleted from the web settings page).
	if resp.StatusCode == http.StatusNoContent ||
		resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusNotFound {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("disconnect channel failed (status %d): %s", resp.StatusCode, string(respBody))
}

type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type DeviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

type OAuthErrorResponse struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *OAuthErrorResponse) Error() string {
	if e.ErrorDescription != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorDescription)
	}
	return e.ErrorCode
}

const (
	OAuthErrAuthorizationPending = "authorization_pending"
	OAuthErrSlowDown             = "slow_down"
	OAuthErrExpiredToken         = "expired_token"
	OAuthErrAccessDenied         = "access_denied"
)

func (c *Client) RequestDeviceAuthorization(ctx context.Context, channelName string) (*DeviceAuthorizationResponse, error) {
	url := fmt.Sprintf("%s/api/v1/channels/cli/authorizations", c.cfg.ServerURL)
	payload := map[string]any{
		"channel_name": channelName,
	}

	var res DeviceAuthorizationResponse
	if err := c.postJSON(ctx, url, payload, http.StatusCreated, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) PollDeviceToken(ctx context.Context, deviceCode string) (*DeviceTokenResponse, error) {
	reqURL := fmt.Sprintf("%s/api/v1/channels/cli/authorizations/token", c.cfg.ServerURL)
	payload := map[string]any{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"device_code": deviceCode,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, mustJSON(payload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var tokenRes DeviceTokenResponse
		if err := json.Unmarshal(respBody, &tokenRes); err != nil {
			return nil, err
		}
		return &tokenRes, nil
	}

	var oauthErr OAuthErrorResponse
	if err := json.Unmarshal(respBody, &oauthErr); err == nil && oauthErr.ErrorCode != "" {
		return nil, &oauthErr
	}

	return nil, fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(respBody))
}

func mustJSON(payload any) *bytes.Reader {
	bodyBytes, _ := json.Marshal(payload)
	return bytes.NewReader(bodyBytes)
}
