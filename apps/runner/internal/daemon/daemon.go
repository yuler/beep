package daemon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/config"
	"beep-runner/internal/exec"
	"beep-runner/internal/probe"
)

type Daemon struct {
	cfg        *config.Config
	client     *client.Client
	dispatcher *probe.Dispatcher
	executor   *exec.ScriptExecutor
	sem        chan struct{}
	wg         sync.WaitGroup
}

func New(cfg *config.Config) *Daemon {
	executor := exec.NewScriptExecutor(cfg.AllowExec)
	dispatcher := probe.NewDispatcher(executor.Run)

	return &Daemon{
		cfg:        cfg,
		client:     client.New(cfg),
		dispatcher: dispatcher,
		executor:   executor,
		sem:        make(chan struct{}, cfg.Concurrency),
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	log.Printf("[beep-runner] Connecting to server %s (concurrency=%d, allow_exec=%v)...",
		d.cfg.ServerURL, d.cfg.Concurrency, d.cfg.AllowExec)

	// Initial Ping
	pingRes, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("initial handshake failed: %w", err)
	}

	log.Printf("[beep-runner] Connected successfully! Runner ID: %s (%s)", pingRes.RunnerID, pingRes.RunnerName)

	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[beep-runner] Shutting down, waiting for running tasks to complete...")
			d.wg.Wait()
			log.Println("[beep-runner] All tasks completed. Goodbye.")
			return nil

		case <-ticker.C:
			d.pollAndExecute(ctx)
		}
	}
}

func (d *Daemon) pollAndExecute(ctx context.Context) {
	// Attempt to poll a task
	task, err := d.client.Poll(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("[beep-runner] Poll error: %v", err)
		}
		return
	}

	if task == nil {
		return // No tasks pending
	}

	// Acquire concurrency slot
	d.sem <- struct{}{}
	d.wg.Add(1)

	go func(t *probe.RunnerTask) {
		defer func() {
			<-d.sem
			d.wg.Done()
		}()

		log.Printf("[beep-runner] Executing task %s (%s - %s)...", t.ID, t.AppSlug, t.Title)

		taskCtx, cancel := context.WithTimeout(context.Background(), time.Duration(t.TimeoutSeconds+5)*time.Second)
		defer cancel()

		signal := d.dispatcher.Dispatch(taskCtx, t)

		log.Printf("[beep-runner] Task %s finished: status=%s title=%q", t.ID, signal.Status, signal.Title)

		if err := d.client.ReportResult(taskCtx, t.ID, signal); err != nil {
			log.Printf("[beep-runner] Error reporting result for task %s: %v", t.ID, err)
		}
	}(task)
}
