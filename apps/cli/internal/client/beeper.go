package client

import (
	"context"
	"fmt"
	"net/http"
)

type BeeperAppInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

type BeeperApp struct {
	ID                 string           `json:"id"`
	Slug               string           `json:"slug"`
	Version            string           `json:"version"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	DefaultCron        string           `json:"default_cron"`
	FailureThreshold   int              `json:"failure_threshold"`
	MinIntervalSeconds int              `json:"min_interval_seconds"`
	Capabilities       []string         `json:"capabilities"`
	WebhookPing        bool             `json:"webhook_ping"`
	Official           bool             `json:"official"`
	Inputs             []BeeperAppInput `json:"inputs,omitempty"`
	Metrics            []any            `json:"metrics,omitempty"`
	CreatedAt          string           `json:"created_at,omitempty"`
}

type Beeper struct {
	ID                   string          `json:"id"`
	Title                string          `json:"title"`
	Body                 string          `json:"body,omitempty"`
	Cron                 string          `json:"cron"`
	Timezone             string          `json:"timezone"`
	Status               string          `json:"status"`
	AlertState           string          `json:"alert_state"`
	ConsecutiveFailures  int             `json:"consecutive_failures"`
	Config               map[string]any  `json:"config,omitempty"`
	SignalMetadata       map[string]any  `json:"signal_metadata,omitempty"`
	NotificationChannels []string        `json:"notification_channels,omitempty"`
	PingToken            string          `json:"ping_token,omitempty"`
	LastPingAt           string          `json:"last_ping_at,omitempty"`
	NextRunAt            string          `json:"next_run_at,omitempty"`
	LastRunAt            string          `json:"last_run_at,omitempty"`
	CreatedAt            string          `json:"created_at,omitempty"`
	UpdatedAt            string          `json:"updated_at,omitempty"`
	BeeperApp            *BeeperApp      `json:"beeper_app,omitempty"`
	RunStats             *BeeperRunStats `json:"run_stats,omitempty"`
	Runs                 []BeeperRun     `json:"runs,omitempty"`
}

type BeeperRunStats struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
}

type BeeperRun struct {
	ID           string `json:"id"`
	ScheduledFor string `json:"scheduled_for"`
	Status       string `json:"status"`
	SignalStatus string `json:"signal_status,omitempty"`
	SignalResult any    `json:"signal_result,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type CreateBeeperRequest struct {
	BeeperAppSlug        string         `json:"beeper_app_slug"`
	Title                string         `json:"title"`
	Body                 string         `json:"body,omitempty"`
	Cron                 string         `json:"cron,omitempty"`
	Timezone             string         `json:"timezone,omitempty"`
	Config               map[string]any `json:"config,omitempty"`
	NotificationChannels []string       `json:"notification_channels,omitempty"`
}

type listBeeperAppsResponse struct {
	BeeperApps []*BeeperApp `json:"beeper_apps"`
}

type listBeepersResponse struct {
	Beepers []*Beeper `json:"beepers"`
}

type listBeeperRunsResponse struct {
	Runs []*BeeperRun `json:"runs"`
}

func (c *Client) ListBeeperApps(ctx context.Context) ([]*BeeperApp, error) {
	url := fmt.Sprintf("%s/api/v1/beeper_apps", c.cfg.ServerURL)
	var res listBeeperAppsResponse
	if err := c.getAuthJSON(ctx, url, &res); err != nil {
		return nil, err
	}
	return res.BeeperApps, nil
}

func (c *Client) GetBeeperApp(ctx context.Context, slug string) (*BeeperApp, error) {
	url := fmt.Sprintf("%s/api/v1/beeper_apps/%s", c.cfg.ServerURL, slug)
	var app BeeperApp
	if err := c.getAuthJSON(ctx, url, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *Client) ListBeepers(ctx context.Context) ([]*Beeper, error) {
	url := fmt.Sprintf("%s/api/v1/beepers", c.cfg.ServerURL)
	var res listBeepersResponse
	if err := c.getAuthJSON(ctx, url, &res); err != nil {
		return nil, err
	}
	return res.Beepers, nil
}

func (c *Client) GetBeeper(ctx context.Context, id string) (*Beeper, error) {
	url := fmt.Sprintf("%s/api/v1/beepers/%s", c.cfg.ServerURL, id)
	var beeper Beeper
	if err := c.getAuthJSON(ctx, url, &beeper); err != nil {
		return nil, err
	}
	return &beeper, nil
}

func (c *Client) CreateBeeper(ctx context.Context, req *CreateBeeperRequest) (*Beeper, error) {
	url := fmt.Sprintf("%s/api/v1/beepers", c.cfg.ServerURL)
	var beeper Beeper
	if err := c.postAuthJSON(ctx, url, req, http.StatusCreated, &beeper); err != nil {
		return nil, err
	}
	return &beeper, nil
}

func (c *Client) DeleteBeeper(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/v1/beepers/%s", c.cfg.ServerURL, id)
	return c.deleteAuth(ctx, url)
}

func (c *Client) PauseBeeper(ctx context.Context, id string) (*Beeper, error) {
	url := fmt.Sprintf("%s/api/v1/beepers/%s/pause", c.cfg.ServerURL, id)
	var beeper Beeper
	if err := c.postAuthJSON(ctx, url, nil, http.StatusOK, &beeper); err != nil {
		return nil, err
	}
	return &beeper, nil
}

func (c *Client) ResumeBeeper(ctx context.Context, id string) (*Beeper, error) {
	url := fmt.Sprintf("%s/api/v1/beepers/%s/pause", c.cfg.ServerURL, id)
	var beeper Beeper
	if err := c.deleteAuthJSON(ctx, url, &beeper); err != nil {
		return nil, err
	}
	return &beeper, nil
}

func (c *Client) RunBeeper(ctx context.Context, id string) (*BeeperRun, error) {
	url := fmt.Sprintf("%s/api/v1/beepers/%s/runs", c.cfg.ServerURL, id)
	var run BeeperRun
	if err := c.postAuthJSON(ctx, url, nil, http.StatusCreated, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (c *Client) ListBeeperRuns(ctx context.Context, id string) ([]*BeeperRun, error) {
	url := fmt.Sprintf("%s/api/v1/beepers/%s/runs", c.cfg.ServerURL, id)
	var res listBeeperRunsResponse
	if err := c.getAuthJSON(ctx, url, &res); err != nil {
		return nil, err
	}
	return res.Runs, nil
}
