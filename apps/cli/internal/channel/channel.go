package channel

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/exec"
	"beep/internal/ui"
	"beep/internal/workspace"
)

// Channel manages listening for notifications from Beep Core and dispatching local hooks.
type Channel struct {
	cfg       *config.Config
	client    *client.Client
	workspace *workspace.Workspace
	OnReady   func()
}

// New creates a new Channel instance.
func New(cfg *config.Config, ws *workspace.Workspace) *Channel {
	return &Channel{
		cfg:       cfg,
		client:    client.New(cfg),
		workspace: ws,
	}
}

// Token returns the active channel token (channel, cli, or device token).
func (c *Channel) Token() string {
	token := c.cfg.ChannelToken
	if token == "" {
		token = c.cfg.CliToken
	}
	if token == "" {
		token = c.cfg.DeviceToken
	}
	return token
}

// Run starts listening for notifications until ctx is canceled.
func (c *Channel) Run(ctx context.Context) error {
	token := c.Token()
	if token == "" {
		return fmt.Errorf("channel token is not configured (run 'beep channel connect' or configure channel_token in config.json)")
	}

	wsRoot := ""
	if c.workspace != nil {
		wsRoot = c.workspace.Root
	}

	if wsRoot != "" {
		log.Printf("%s %s %s=%s %s=%s",
			ui.Bold(ui.Cyan("[beep-channel]")),
			ui.Green("Channel listening active:"),
			ui.Dim("server"), ui.Bold(c.cfg.ServerURL),
			ui.Dim("workspace"), ui.Dim(wsRoot),
		)
	} else {
		log.Printf("%s %s %s=%s",
			ui.Bold(ui.Cyan("[beep-channel]")),
			ui.Green("Channel listening active:"),
			ui.Dim("server"), ui.Bold(c.cfg.ServerURL),
		)
	}

	if c.OnReady != nil {
		c.OnReady()
	}

	interval := c.cfg.PollInterval
	if interval <= 0 {
		interval = 3 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s", ui.Bold(ui.Cyan("[beep-channel]")), ui.Yellow("Shutting down..."))
			return nil
		case <-ticker.C:
			c.PollInbox(ctx)
		}
	}
}

// PollInbox fetches pending deliveries and dispatches local hooks.
func (c *Channel) PollInbox(ctx context.Context) {
	if c.Token() == "" {
		return
	}

	deliveries, err := c.client.FetchCliInbox(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-channel]")), ui.Red("Inbox error:"), err)
		}
		return
	}

	for _, delivery := range deliveries {
		if delivery.ExpiresAt != nil && time.Now().After(*delivery.ExpiresAt) {
			log.Printf("%s %s %s (expired at %s)",
				ui.Bold(ui.Cyan("[beep-channel]")),
				ui.Yellow("Dropped expired delivery:"),
				ui.Bold(delivery.ID),
				delivery.ExpiresAt.Format(time.RFC3339),
			)
			_ = c.client.AckCliDelivery(ctx, delivery.ID, "failed", "expired")
			continue
		}

		title, _ := delivery.Payload["title"].(string)
		eventName, _ := delivery.Payload["event"].(string)
		if eventName == "" {
			eventName = "beep.fired"
		}

		log.Printf("%s %s %s (%s)",
			ui.Bold(ui.Cyan("[beep-channel]")),
			ui.Green("Received notification:"),
			ui.Bold(delivery.ID),
			ui.Dim(title),
		)

		wsRoot := ""
		if c.workspace != nil {
			wsRoot = c.workspace.Root
		}

		out, hookName, hookErr := exec.DispatchHook(ctx, wsRoot, delivery)
		if hookErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-channel]")), ui.Red("Hook execution failed:"), hookErr)
			_ = c.client.AckCliDelivery(ctx, delivery.ID, "failed", hookErr.Error())
			continue
		}
		if hookName == "" {
			log.Printf("%s %s %s (configure %s to handle notifications)",
				ui.Bold(ui.Cyan("[beep-channel]")),
				ui.Yellow("Warning: No on_channel hook found for event:"),
				ui.Bold(eventName),
				ui.Cyan("hooks/on_channel"),
			)
			_ = c.client.AckCliDelivery(ctx, delivery.ID, "failed", "no on_channel hook configured")
			continue
		} else if strings.TrimSpace(out) != "" {
			log.Printf("%s %s %s", ui.Bold(ui.Cyan("[beep-channel]")), ui.Dim("Hook output:"), strings.TrimSpace(out))
		}
		_ = c.client.AckCliDelivery(ctx, delivery.ID, "succeeded", "")
	}
}
