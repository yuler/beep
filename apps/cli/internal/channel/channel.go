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

	log.Printf("%s %s %s=%s",
		ui.Bold(ui.Cyan("[beep-channel]")),
		ui.Green("Channel listening active:"),
		ui.Dim("server"), ui.Bold(c.cfg.ServerURL),
	)

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

// PollInbox fetches pending deliveries and dispatches notifications and hooks.
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
		body, _ := delivery.Payload["body"].(string)
		if body == "" {
			body, _ = delivery.Payload["message"].(string)
		}

		log.Printf("%s %s %s (%s)",
			ui.Bold(ui.Cyan("[beep-channel]")),
			ui.Green("Received notification:"),
			ui.Bold(delivery.ID),
			ui.Dim(title),
		)

		if title != "" || body != "" {
			notifTitle := title
			if notifTitle == "" {
				notifTitle = "Beep Notification"
			}
			exec.NotifyDesktop(notifTitle, body)
		}

		wsRoot := ""
		if c.workspace != nil {
			wsRoot = c.workspace.Root
		}

		out, hookErr := exec.DispatchOnBeepHook(ctx, wsRoot, delivery)
		if hookErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-channel]")), ui.Red("Hook execution failed:"), hookErr)
			_ = c.client.AckCliDelivery(ctx, delivery.ID, "failed", hookErr.Error())
		} else {
			if strings.TrimSpace(out) != "" {
				log.Printf("%s %s %s", ui.Bold(ui.Cyan("[beep-channel]")), ui.Dim("Hook output:"), strings.TrimSpace(out))
			}
			_ = c.client.AckCliDelivery(ctx, delivery.ID, "succeeded", "")
		}
	}
}
