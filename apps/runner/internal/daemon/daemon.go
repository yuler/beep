package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/config"
	"beep-runner/internal/exec"
	"beep-runner/internal/task"
	"beep-runner/internal/ui"
	"beep-runner/internal/workspace"
)

type Daemon struct {
	cfg       *config.Config
	client    *client.Client
	workspace *workspace.Workspace
	executor  *exec.JobExecutor
	sem       chan struct{}
	wg        sync.WaitGroup
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
	for {
		if len(d.sem) >= cap(d.sem) {
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
			d.execute(job)
		}(t)
	}
}

func (d *Daemon) execute(job *task.Task) {
	log.Printf("%s %s %s (%s: %s)",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Yellow("Running"),
		ui.Bold(ui.Cyan(job.JobSlug)),
		ui.Dim("run_id"),
		ui.Dim(job.ID),
	)

	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	reportCtx, reportCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer reportCancel()

	argv, err := d.workspace.Resolve(job.JobSlug)
	if err != nil {
		result := task.Error("Unknown local job", err.Error(), nil)
		_ = d.client.ReportLog(reportCtx, job.LogURL, err.Error()+"\n")
		if reportErr := d.client.ReportResult(reportCtx, job.ResultURL, result); reportErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error:"), reportErr)
		}
		return
	}

	env := d.jobEnv(job)

	var logBuf string
	var logMu sync.Mutex
	flush := func(ctx context.Context, force bool) {
		logMu.Lock()
		chunk := logBuf
		if !force && len(chunk) < 256 {
			logMu.Unlock()
			return
		}
		logBuf = ""
		logMu.Unlock()
		if chunk == "" {
			return
		}
		if err := d.client.ReportLog(ctx, job.LogURL, chunk); err != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("log upload:"), err)
		}
	}

	result := d.executor.Run(taskCtx, argv, env, timeout, func(line string) {
		log.Print(ui.Dim(fmt.Sprintf("[%s]", job.JobSlug)) + " " + line)
		logMu.Lock()
		logBuf += line
		logMu.Unlock()
		flush(taskCtx, false)
	})
	flush(reportCtx, true)

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

	if err := d.client.ReportResult(reportCtx, job.ResultURL, result); err != nil {
		log.Printf("%s %s for %s: %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error"), job.ID, err)
	}
}

func (d *Daemon) jobEnv(job *task.Task) []string {
	configJSON, _ := json.Marshal(job.Config)
	return exec.WithJobEnv(append(exec.ConfigEnv(job.Config),
		"BEEP_SERVER="+d.cfg.ServerURL,
		"BEEP_RUN_ID="+job.ID,
		"BEEP_JOB_SLUG="+job.JobSlug,
		"BEEP_LOG_URL="+job.LogURL,
		"BEEP_RESULT_URL="+job.ResultURL,
		"BEEP_CONFIG="+string(configJSON),
	))
}
