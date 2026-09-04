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
	log.Printf("[beep-runner] Connecting to %s workspace=%s concurrency=%d",
		d.cfg.ServerURL, d.workspace.Root, d.cfg.Concurrency)

	pingRes, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("initial handshake failed: %w", err)
	}
	log.Printf("[beep-runner] Connected: %s (%s)", pingRes.RunnerID, pingRes.RunnerName)

	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[beep-runner] Shutting down...")
			d.wg.Wait()
			return nil
		case <-ticker.C:
			d.pollAndExecute(ctx)
		}
	}
}

func (d *Daemon) pollAndExecute(ctx context.Context) {
	t, err := d.client.Poll(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("[beep-runner] Poll error: %v", err)
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

func (d *Daemon) execute(job *task.Task) {
	log.Printf("[beep-runner] Running %s (%s)", job.JobSlug, job.ID)

	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	taskCtx, cancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
	defer cancel()

	argv, err := d.workspace.Resolve(job.JobSlug)
	if err != nil {
		result := task.Error("Unknown local job", err.Error(), nil)
		_ = d.client.ReportLog(taskCtx, job.LogURL, err.Error()+"\n")
		if reportErr := d.client.ReportResult(taskCtx, job.ResultURL, result); reportErr != nil {
			log.Printf("[beep-runner] result error: %v", reportErr)
		}
		return
	}

	configJSON, _ := json.Marshal(job.Config)
	env := exec.WithJobEnv(append(exec.ConfigEnv(job.Config),
		"BEEP_SERVER="+d.cfg.ServerURL,
		"BEEP_RUNNER_TOKEN="+d.cfg.RunnerToken,
		"BEEP_RUN_ID="+job.ID,
		"BEEP_JOB_SLUG="+job.JobSlug,
		"BEEP_LOG_URL="+job.LogURL,
		"BEEP_RESULT_URL="+job.ResultURL,
		"BEEP_CONFIG="+string(configJSON),
	))

	var logBuf string
	var logMu sync.Mutex
	flush := func(force bool) {
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
		if err := d.client.ReportLog(taskCtx, job.LogURL, chunk); err != nil {
			log.Printf("[beep-runner] log upload: %v", err)
		}
	}

	result := d.executor.Run(taskCtx, argv, env, timeout, func(line string) {
		log.Print("[job " + job.JobSlug + "] " + line)
		logMu.Lock()
		logBuf += line
		logMu.Unlock()
		flush(false)
	})
	flush(true)

	if err := d.client.ReportResult(taskCtx, job.ResultURL, result); err != nil {
		log.Printf("[beep-runner] result error for %s: %v", job.ID, err)
	}
}
