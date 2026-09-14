package runner

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

// Runner manages local execution of jobs received from Beep Core.
type Runner struct {
	cfg       *config.Config
	client    *client.Client
	workspace *workspace.Workspace
	executor  *exec.JobExecutor
	sem       chan struct{}
	wg        sync.WaitGroup
	logMu     sync.Mutex
	OnReady   func()
}

// New creates a new Runner instance.
func New(cfg *config.Config, ws *workspace.Workspace) *Runner {
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}
	return &Runner{
		cfg:       cfg,
		client:    client.New(cfg),
		workspace: ws,
		executor:  exec.NewJobExecutor(),
		sem:       make(chan struct{}, concurrency),
	}
}

// Run starts the runner loop until ctx is canceled.
func (r *Runner) Run(ctx context.Context) error {
	if r.cfg.RunnerToken == "" {
		return fmt.Errorf("runner token is not configured (set via BEEP_RUNNER_TOKEN or config.json)")
	}

	log.Printf("%s Connecting to %s %s=%s %s=%d",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Bold(r.cfg.ServerURL),
		ui.Dim("workspace"), ui.Dim(r.workspace.Root),
		ui.Dim("concurrency"), r.cfg.Concurrency,
	)

	pingRes, err := r.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("runner handshake failed: %w", err)
	}
	log.Printf("%s %s %s (%s)",
		ui.Bold(ui.Cyan("[beep-runner]")),
		ui.Green("Connected:"),
		ui.Bold(pingRes.RunnerID),
		ui.Dim(pingRes.RunnerName),
	)

	if r.OnReady != nil {
		r.OnReady()
	}

	interval := r.cfg.PollInterval
	if interval <= 0 {
		interval = 3 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s %s", ui.Bold(ui.Cyan("[beep-runner]")), ui.Yellow("Shutting down..."))
			r.wg.Wait()
			return nil
		case <-ticker.C:
			r.PollAndExecute(ctx)
		}
	}
}

// PollAndExecute fetches pending jobs and executes them concurrently up to configured concurrency limit.
func (r *Runner) PollAndExecute(ctx context.Context) {
	if r.cfg.RunnerToken == "" {
		return
	}

	for {
		if len(r.sem) >= cap(r.sem) {
			pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
			_, err := r.client.Ping(pingCtx)
			pingCancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("Heartbeat error:"), err)
			}
			return
		}
		t, err := r.client.Poll(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("Poll error:"), err)
			}
			return
		}
		if t == nil {
			return
		}

		r.sem <- struct{}{}
		r.wg.Add(1)
		go func(job *task.Task) {
			defer func() {
				<-r.sem
				r.wg.Done()
			}()
			r.execute(ctx, job)
		}(t)
	}
}

func (r *Runner) execute(ctx context.Context, job *task.Task) {
	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	argv, err := r.workspace.Resolve(job.JobSlug)
	if err != nil {
		result := task.Error("Unknown local job", err.Error(), nil)
		r.logJobBlock(job, []string{err.Error()}, result)
		errCtx, errCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer errCancel()
		_ = r.client.ReportLog(errCtx, job.LogURL, err.Error()+"\n")
		if reportErr := r.client.ReportResult(errCtx, job.ResultURL, result); reportErr != nil {
			log.Printf("%s %s %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error:"), reportErr)
		}
		return
	}

	env, err := r.JobEnv(job)
	if err != nil {
		result := task.Error("Workspace environment", err.Error(), nil)
		r.logJobBlock(job, []string{err.Error()}, result)
		errCtx, errCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer errCancel()
		_ = r.client.ReportLog(errCtx, job.LogURL, err.Error()+"\n")
		if reportErr := r.client.ReportResult(errCtx, job.ResultURL, result); reportErr != nil {
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
			if err := r.client.ReportLog(uploadCtx, job.LogURL, chunk); err != nil {
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

	result := r.executor.Run(taskCtx, argv, env, timeout, func(line string) {
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

	r.logJobBlock(job, localLines, result)

	resultCtx, resultCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer resultCancel()
	if err := r.client.ReportResult(resultCtx, job.ResultURL, result); err != nil {
		log.Printf("%s %s for %s: %v", ui.Bold(ui.Cyan("[beep-runner]")), ui.Red("result error"), job.ID, err)
	}
}

func (r *Runner) logJobBlock(job *task.Task, lines []string, result *task.Result) {
	r.logMu.Lock()
	defer r.logMu.Unlock()

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

// JobEnv resolves the execution environment variables for a job.
func (r *Runner) JobEnv(job *task.Task) ([]string, error) {
	configJSON, _ := json.Marshal(job.Config)

	var wsEnv []string
	if r.workspace != nil {
		var err error
		wsEnv, err = r.workspace.LoadEnv()
		if err != nil {
			return nil, err
		}
	}

	var extras []string
	extras = append(extras, wsEnv...)
	extras = append(extras, exec.ConfigEnv(job.Config)...)
	extras = append(extras,
		"BEEP_RUNNER_SERVER="+r.cfg.ServerURL,
		"BEEP_RUNNER_RUN_ID="+job.ID,
		"BEEP_RUNNER_JOB_SLUG="+job.JobSlug,
		"BEEP_RUNNER_LOG_URL="+job.LogURL,
		"BEEP_RUNNER_RESULT_URL="+job.ResultURL,
		"BEEP_RUNNER_CONFIG="+string(configJSON),
	)

	return exec.WithJobEnv(extras), nil
}
