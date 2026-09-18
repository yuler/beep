package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"beep/internal/schedule"
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
	Intent               string         `json:"intent,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

// ProposalErrors represents errors returned in BeepProposal, which can be an object
// (map of field to error message, e.g. {"cron": "can't be blank"} or {}), an array
// of strings, or a map of string to array of strings.
type ProposalErrors []string

func (pe *ProposalErrors) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*pe = nil
		return nil
	}

	// 1. Array of strings: ["error 1", "error 2"]
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*pe = list
		return nil
	}

	// 2. Map of string to string: {"cron": "can't be blank"}
	var mapStr map[string]string
	if err := json.Unmarshal(data, &mapStr); err == nil {
		if len(mapStr) == 0 {
			*pe = nil
			return nil
		}
		keys := make([]string, 0, len(mapStr))
		for k := range mapStr {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		res := make([]string, 0, len(keys))
		for _, k := range keys {
			res = append(res, fmt.Sprintf("%s %s", k, mapStr[k]))
		}
		*pe = res
		return nil
	}

	// 3. Map of string to []string: {"cron": ["can't be blank"]}
	var mapList map[string][]string
	if err := json.Unmarshal(data, &mapList); err == nil {
		if len(mapList) == 0 {
			*pe = nil
			return nil
		}
		keys := make([]string, 0, len(mapList))
		for k := range mapList {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var res []string
		for _, k := range keys {
			for _, msg := range mapList[k] {
				res = append(res, fmt.Sprintf("%s %s", k, msg))
			}
		}
		*pe = res
		return nil
	}

	// 4. Map of string to any
	var mapAny map[string]any
	if err := json.Unmarshal(data, &mapAny); err == nil {
		if len(mapAny) == 0 {
			*pe = nil
			return nil
		}
		keys := make([]string, 0, len(mapAny))
		for k := range mapAny {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		res := make([]string, 0, len(keys))
		for _, k := range keys {
			res = append(res, fmt.Sprintf("%s %v", k, mapAny[k]))
		}
		*pe = res
		return nil
	}

	// 5. Single string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s != "" {
			*pe = []string{s}
		} else {
			*pe = nil
		}
		return nil
	}

	return nil
}

type BeepProposal struct {
	Action      string         `json:"action,omitempty"`
	Intent      string         `json:"intent,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Kind        string         `json:"kind"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	RunAt       string         `json:"run_at"`
	Cron        string         `json:"cron"`
	Timezone    string         `json:"timezone"`
	Channels    []string       `json:"notification_channels"`
	Errors      ProposalErrors `json:"errors"`
	Confirmable bool           `json:"confirmable"`
	Message     string         `json:"message"`
}

func (p *BeepProposal) UnmarshalJSON(data []byte) error {
	type Alias BeepProposal
	aux := struct {
		*Alias
		AltChannels []string `json:"channels"`
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(p.Channels) == 0 && len(aux.AltChannels) > 0 {
		p.Channels = aux.AltChannels
	}
	if p.Action == "" && (p.Intent == "create" || p.Intent == "other") {
		p.Action = p.Intent
		p.Intent = ""
	}
	return nil
}

func (p *BeepProposal) HasErrors() bool {
	return p != nil && len(p.Errors) > 0
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

// CreateBeepParams encapsulates user parameters for creating a Beep.
type CreateBeepParams struct {
	Title        string
	Body         string
	ScheduleKind string // "instant", "delay", "at", "cron"
	ScheduleVal  string
	Timezone     string
	Channels     string
	Intent       string
	Metadata     map[string]any
}

// ToRequest validates and transforms CreateBeepParams into a CreateBeepRequest.
func (p *CreateBeepParams) ToRequest() (*CreateBeepRequest, error) {
	if strings.TrimSpace(p.Title) == "" {
		return nil, errors.New("beep title is required")
	}

	tz := strings.TrimSpace(p.Timezone)
	if tz == "" {
		tz = "UTC"
	}

	req := &CreateBeepRequest{
		Title:    strings.TrimSpace(p.Title),
		Body:     strings.TrimSpace(p.Body),
		Timezone: tz,
		Intent:   strings.TrimSpace(p.Intent),
		Metadata: p.Metadata,
	}

	if trimmedChannels := strings.TrimSpace(p.Channels); trimmedChannels != "" {
		seen := make(map[string]bool)
		for _, ch := range strings.Split(trimmedChannels, ",") {
			if trimmed := strings.TrimSpace(ch); trimmed != "" && !seen[trimmed] {
				seen[trimmed] = true
				req.NotificationChannels = append(req.NotificationChannels, trimmed)
			}
		}
	}

	kind := p.ScheduleKind
	if kind == "" {
		kind = "instant"
	}

	switch kind {
	case "cron":
		if err := schedule.Validate(p.ScheduleVal); err != nil {
			return nil, fmt.Errorf("invalid --cron expression: %w", err)
		}
		req.Kind = "recurring"
		req.Cron = strings.TrimSpace(p.ScheduleVal)
	case "delay":
		d, err := ParseInDuration(p.ScheduleVal)
		if err != nil {
			return nil, fmt.Errorf("invalid --in duration (e.g. 10m, 2h, 1d): %w", err)
		}
		req.Kind = "once"
		req.RunAt = time.Now().Add(d).Format(time.RFC3339)
	case "at":
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
		}
		t, err := ParseAtTime(p.ScheduleVal, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid --at time (e.g. 15:30, 2026-10-01 10:00): %w", err)
		}
		req.Kind = "once"
		req.RunAt = t.Format(time.RFC3339)
	case "instant":
		req.Kind = "once"
		req.RunAt = time.Now().Format(time.RFC3339)
	default:
		return nil, fmt.Errorf("unknown schedule kind %q", kind)
	}

	return req, nil
}

// ParseInDuration parses a delay duration string supporting 'd' suffix (e.g. 1d, 2d, 15m, 2h).
func ParseInDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty duration")
	}
	var d time.Duration
	var err error
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, aErr := strconv.Atoi(daysStr)
		if aErr != nil {
			return 0, fmt.Errorf("invalid days format %q: %w", s, aErr)
		}
		d = time.Duration(days) * 24 * time.Hour
	} else {
		d, err = time.ParseDuration(s)
		if err != nil {
			return 0, err
		}
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero: %q", s)
	}
	return d, nil
}

// ParseAtTime parses a specific time/datetime string relative to location.
func ParseAtTime(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range formats {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}

	// Try time-only "15:04"
	if t, err := time.ParseInLocation("15:04", s, loc); err == nil {
		now := time.Now().In(loc)
		target := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
		if target.Before(now) {
			target = target.AddDate(0, 0, 1)
		}
		return target, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized time format %q", s)
}
