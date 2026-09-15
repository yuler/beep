package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Beep struct {
	ID                   string         `json:"id"`
	Title                string         `json:"title"`
	Body                 string         `json:"body,omitempty"`
	Kind                 string         `json:"kind"`
	Status               string         `json:"status"`
	Cron                 string         `json:"cron,omitempty"`
	RunAt                string         `json:"run_at,omitempty"`
	NextRunAt            string         `json:"next_run_at,omitempty"`
	LastRunAt            string         `json:"last_run_at,omitempty"`
	Timezone             string         `json:"timezone,omitempty"`
	NotificationChannels []string       `json:"notification_channels,omitempty"`
	BeeperID             string         `json:"beeper_id,omitempty"`
	SourceType           string         `json:"source_type,omitempty"`
	SourceID             string         `json:"source_id,omitempty"`
	Intent               string         `json:"intent,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
	Source               string         `json:"source,omitempty"`
	Beeper               *BeepBeeperRef `json:"beeper,omitempty"`
	Runs                 []BeepRun      `json:"runs,omitempty"`
	CreatedAt            string         `json:"created_at,omitempty"`
}

type BeepBeeperRef struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type BeepRun struct {
	ID           string `json:"id"`
	ScheduledFor string `json:"scheduled_for"`
	Status       string `json:"status"`
	Result       any    `json:"result,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type CreateBeepRequest struct {
	Title                string         `json:"title"`
	Body                 string         `json:"body,omitempty"`
	Kind                 string         `json:"kind,omitempty"`
	RunAt                string         `json:"run_at,omitempty"`
	Cron                 string         `json:"cron,omitempty"`
	Timezone             string         `json:"timezone,omitempty"`
	NotificationChannels []string       `json:"notification_channels,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

type BeepProposal struct {
	Intent      string   `json:"intent"`
	Kind        string   `json:"kind"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	RunAt       string   `json:"run_at"`
	Cron        string   `json:"cron"`
	Timezone    string   `json:"timezone"`
	Errors      []string `json:"errors"`
	Confirmable bool     `json:"confirmable"`
	Message     string   `json:"message"`
}

type listBeepsResponse struct {
	Beeps []*Beep `json:"beeps"`
}

func (c *Client) ListBeeps(ctx context.Context) ([]*Beep, error) {
	url := fmt.Sprintf("%s/api/v1/beeps", c.cfg.ServerURL)
	var res listBeepsResponse
	if err := c.getAuthJSON(ctx, url, &res); err != nil {
		return nil, err
	}
	return res.Beeps, nil
}

func (c *Client) GetBeep(ctx context.Context, id string) (*Beep, error) {
	url := fmt.Sprintf("%s/api/v1/beeps/%s", c.cfg.ServerURL, url.PathEscape(id))
	var beep Beep
	if err := c.getAuthJSON(ctx, url, &beep); err != nil {
		return nil, err
	}
	return &beep, nil
}

func (c *Client) CreateBeep(ctx context.Context, req *CreateBeepRequest) (*Beep, error) {
	url := fmt.Sprintf("%s/api/v1/beeps", c.cfg.ServerURL)
	var beep Beep
	if err := c.postAuthJSON(ctx, url, req, http.StatusCreated, &beep); err != nil {
		return nil, err
	}
	return &beep, nil
}

func (c *Client) DeleteBeep(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/v1/beeps/%s", c.cfg.ServerURL, url.PathEscape(id))
	return c.deleteAuth(ctx, url)
}

func (c *Client) PauseBeep(ctx context.Context, id string) (*Beep, error) {
	url := fmt.Sprintf("%s/api/v1/beeps/%s/pause", c.cfg.ServerURL, url.PathEscape(id))
	var beep Beep
	if err := c.postAuthJSON(ctx, url, nil, http.StatusOK, &beep); err != nil {
		return nil, err
	}
	return &beep, nil
}

func (c *Client) ResumeBeep(ctx context.Context, id string) (*Beep, error) {
	url := fmt.Sprintf("%s/api/v1/beeps/%s/pause", c.cfg.ServerURL, url.PathEscape(id))
	var beep Beep
	if err := c.deleteAuthJSON(ctx, url, &beep); err != nil {
		return nil, err
	}
	return &beep, nil
}

func (c *Client) RunBeep(ctx context.Context, id string) (*BeepRun, error) {
	url := fmt.Sprintf("%s/api/v1/beeps/%s/runs", c.cfg.ServerURL, url.PathEscape(id))
	var run BeepRun
	if err := c.postAuthJSON(ctx, url, nil, http.StatusCreated, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (c *Client) ProposeBeep(ctx context.Context, prompt, timezone string) (*BeepProposal, error) {
	url := fmt.Sprintf("%s/api/v1/beep_proposals", c.cfg.ServerURL)
	payload := map[string]string{
		"prompt": prompt,
	}
	if timezone != "" {
		payload["timezone"] = timezone
	}
	var proposal BeepProposal
	if err := c.postAuthJSON(ctx, url, payload, http.StatusCreated, &proposal); err != nil {
		return nil, err
	}
	return &proposal, nil
}
