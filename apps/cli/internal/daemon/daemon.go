package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"beep/internal/client"
	"beep/internal/config"
	"beep/internal/exec"
	"beep/internal/task"
	"beep/internal/ui"
	"beep/internal/workspace"
)

type Daemon struct {
	cfg       *config.Config
	client    *client.Client
	workspace *workspace.Workspace
	executor  *exec.JobExecutor
	sem       chan struct{}
	wg        sync.WaitGroup
	logMu     sync.Mutex
	OnReady   func()
}

func New(cfg *config.Config, ws *workspace.Workspace) *Daemon {
	return &Daemon{
		cfg:       cfg,
		client:    client.New(cfg),
		workspace: ws,
		executor:  exec.NewJobExecutor(),
		sem:       make(chan struct{}, cfg.Concurrency),
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	log.Printf("%s Connecting to %s %s=%s %s=%s",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Bold(d.cfg.ServerURL),
		ui.Dim("workspace"), ui.Dim(d.workspace.Root),
		ui.Dim("concurrency"), ui.Yellow(fmt.Sprintf("%d", d.cfg.Concurrency)),
	)

	pingRes, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("initial handshake failed: %w", err)
	}
	if d.OnReady != nil {
		d.OnReady()
	}
	log.Printf("%s %s %s (%s)",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Green("Connected:"),
		ui.Bold(pingRes.RunnerID),
		ui.Dim(pingRes.RunnerName),
	)

	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s", ui.Bold(ui.Cyan("[beep-runner]")), ui.Yellow("Shutting down..."))
			d.wg.Wait()
			return nil
		case <-ticker.C:
			d.pollAndExecute(ctx)
		}
	}
}

func (d *Daemon) pollAndExecute(ctx context.Context) {
	d.pollDeviceInbox(ctx)

	for {
		if len(d.sem) >= cap(d.sem) {
			pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
			_, err := d.client.Ping(pingCtx)
			pingCancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("Heartbeat error:"), err)
			}
			return
		}
		t, err := d.client.Poll(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("Poll error:"), err)
			}
			return
		}
		if t == nil {
			return
		}

		d.sem <- struct{}{}
		d.wg.Add(1)
		go func(job *task.Task) {
			defer func() {
				<-d.sem
				d.wg.Done()
			}()
			d.execute(ctx, job)
		}(t)
	}
}

func (d *Daemon) execute(ctx context.Context, job *task.Task) {
	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	argv, err := d.workspace.Resolve(job.JobSlug)
	if err != nil {
		result := task.Error("Unknown local job", err.Error(), nil)
		d.logJobBlock(job, []string{err.Error()}, result)
		errCtx, errCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer errCancel()
		_ = d.client.ReportLog(errCtx, job.LogURL, err.Error()+"\n")
		if reportErr := d.client.ReportResult(errCtx, job.ResultURL, result); reportErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error:"), reportErr)
		}
		return
	}

	env, err := d.jobEnv(job)
	if err != nil {
		result := task.Error("Workspace environment", err.Error(), nil)
		d.logJobBlock(job, []string{err.Error()}, result)
		errCtx, errCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer errCancel()
		_ = d.client.ReportLog(errCtx, job.LogURL, err.Error()+"\n")
		if reportErr := d.client.ReportResult(errCtx, job.ResultURL, result); reportErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error:"), reportErr)
		}
		return
	}

	logChan := make(chan string, 1000)
	var logWg sync.WaitGroup
	logWg.Add(1)

	go func() {
		defer logWg.Done()
		var buf strings.Builder
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		flushChunk := func() {
			if buf.Len() == 0 {
				return
			}
			chunk := buf.String()
			buf.Reset()
			uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer uploadCancel()
			if err := d.client.ReportLog(uploadCtx, job.LogURL, chunk); err != nil {
				log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("log upload:"), err)
			}
		}

		for {
			select {
			case line, ok := <-logChan:
				if !ok {
					flushChunk()
					return
				}
				buf.WriteString(line)
				if buf.Len() >= 1024 {
					flushChunk()
				}
			case <-ticker.C:
				flushChunk()
			}
		}
	}()

	var localLines []string
	var localMu sync.Mutex

	result := d.executor.Run(taskCtx, argv, env, timeout, func(line string) {
		cleanLine := strings.TrimRight(line, "\r\n")
		if cleanLine != "" {
			localMu.Lock()
			localLines = append(localLines, cleanLine)
			localMu.Unlock()
		}
		select {
		case logChan <- line:
		case <-taskCtx.Done():
		}
	})
	close(logChan)
	logWg.Wait()

	d.logJobBlock(job, localLines, result)

	resultCtx, resultCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer resultCancel()
	if err := d.client.ReportResult(resultCtx, job.ResultURL, result); err != nil {
		log.Printf("%s %s for %s: %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error"), job.ID, err)
	}
}

func (d *Daemon) logJobBlock(job *task.Task, lines []string, result *task.Result) {
	d.logMu.Lock()
	defer d.logMu.Unlock()

	log.Printf("%s %s %s (%s: %s)",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Yellow("Running"),
		ui.Bold(ui.Cyan(job.JobSlug)),
		ui.Dim("run_id"),
		ui.Dim(job.ID),
	)

	for _, line := range lines {
		log.Printf("%s %s", ui.Dim(fmt.Sprintf("[%s]", job.JobSlug)), line)
	}

	if result.Status == task.StatusOk {
		log.Printf("%s %s %s %s",
			ui.Bold(ui.Cyan("[beep-runner]")),
			ui.Green("✓"),
			ui.Bold(job.JobSlug),
			ui.Dim(result.Title),
		)
	} else {
		log.Printf("%s %s %s %s",
			ui.Bold(ui.Cyan("[beep-runner]")),
			ui.Red("✗"),
			ui.Bold(job.JobSlug),
			ui.Red(result.Title),
		)
	}
}

func (d *Daemon) jobEnv(job *task.Task) ([]string, error) {
	configJSON, _ := json.Marshal(job.Config)

	wsEnv, err := d.workspace.LoadEnv()
	if err != nil {
		return nil, err
	}

	var extras []string
	extras = append(extras, wsEnv...)
	extras = append(extras, exec.ConfigEnv(job.Config)...)
	extras = append(extras,
		"BEEP_RUNNER_SERVER="+d.cfg.ServerURL,
		"BEEP_RUNNER_RUN_ID="+job.ID,
		"BEEP_RUNNER_JOB_SLUG="+job.JobSlug,
		"BEEP_RUNNER_LOG_URL="+job.LogURL,
		"BEEP_RUNNER_RESULT_URL="+job.ResultURL,
		"BEEP_RUNNER_CONFIG="+string(configJSON),
	)

	return exec.WithJobEnv(extras), nil
}

func (d *Daemon) pollDeviceInbox(ctx context.Context) {
	if d.cfg.DeviceToken == "" {
		return
	}

	deliveries, err := d.client.FetchDeviceInbox(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-device]")), ui.Red("Inbox error:"), err)
		}
		return
	}

	for _, delivery := range deliveries {
		if delivery.ExpiresAt != nil && time.Now().After(*delivery.ExpiresAt) {
			log.Printf("%s %s %s (expired at %s)",
				ui.Bold(ui.Cyan("[beep-device]")),
				ui.Yellow("Dropped expired delivery:"),
				ui.Bold(delivery.ID),
				delivery.ExpiresAt.Format(time.RFC3339),
			)
			_ = d.client.AckDeviceDelivery(ctx, delivery.ID, "failed", "expired")
			continue
		}

		title, _ := delivery.Payload["title"].(string)
		log.Printf("%s %s %s (%s)",
			ui.Bold(ui.Cyan("[beep-device]")),
			ui.Green("Received notification:"),
			ui.Bold(delivery.ID),
			ui.Dim(title),
		)

		out, hookErr := exec.DispatchOnBeepHook(ctx, d.workspace.Root, delivery)
		if hookErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-device]")), ui.Red("Hook execution failed:"), hookErr)
			_ = d.client.AckDeviceDelivery(ctx, delivery.ID, "failed", hookErr.Error())
		} else {
			if strings.TrimSpace(out) != "" {
				log.Printf("%s %s %s", ui.Bold(ui.Cyan("[beep-device]")), ui.Dim("Hook output:"), strings.TrimSpace(out))
			}
			_ = d.client.AckDeviceDelivery(ctx, delivery.ID, "succeeded", "")
		}
	}
}
