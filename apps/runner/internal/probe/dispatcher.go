package probe

import (
	"context"
	"fmt"
	"strings"
)

type RunnerTask struct {
	ID             string         `json:"id"`
	BeeperID       string         `json:"beeper_id"`
	Title          string         `json:"title"`
	AppSlug        string         `json:"app_slug"`
	Manifest       map[string]any `json:"manifest,omitempty"`
	Config         map[string]any `json:"config"`
	ScheduledFor   string         `json:"scheduled_for"`
	TimeoutSeconds int            `json:"timeout_seconds"`
}

type Dispatcher struct {
	httpProbe   *HTTPProbe
	tlsProbe    *TLSProbe
	tcpProbe    *TCPProbe
	dnsProbe    *DNSProbe
	execHandler func(ctx context.Context, config map[string]any) *Signal
}

func NewDispatcher(execHandler func(ctx context.Context, config map[string]any) *Signal) *Dispatcher {
	return &Dispatcher{
		httpProbe:   NewHTTPProbe(),
		tlsProbe:    NewTLSProbe(),
		tcpProbe:    NewTCPProbe(),
		dnsProbe:    NewDNSProbe(),
		execHandler: execHandler,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, task *RunnerTask) *Signal {
	slug := strings.ToLower(strings.TrimSpace(task.AppSlug))
	cfg := task.Config
	if cfg == nil {
		cfg = make(map[string]any)
	}

	switch slug {
	case "site-uptime", "http", "https":
		return d.httpProbe.Run(ctx, cfg)

	case "ssl-tls-expiry", "tls", "ssl":
		return d.tlsProbe.Run(ctx, cfg)

	case "tcp-port", "tcp":
		return d.tcpProbe.Run(ctx, cfg)

	case "dns":
		return d.dnsProbe.Run(ctx, cfg)

	case "custom-script", "exec", "script", "shell", "command":
		if d.execHandler != nil {
			return d.execHandler(ctx, cfg)
		}
		return ErrorSignal("Script execution not configured", "No script executor registered", nil)

	default:
		// Fallback: If config contains target_url, try HTTP probe; if host, try TCP
		if _, ok := cfg["target_url"]; ok {
			return d.httpProbe.Run(ctx, cfg)
		}
		if _, ok := cfg["command"]; ok {
			if d.execHandler != nil {
				return d.execHandler(ctx, cfg)
			}
		}
		return ErrorSignal("Unknown probe type", fmt.Sprintf("Unsupported beeper app slug: %s", task.AppSlug), nil)
	}
}
