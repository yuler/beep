package cliservice

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"beep/internal/config"
	"beep/internal/daemon"
	"beep/internal/proc"
	"beep/internal/ui"
)

var osExecutable = os.Executable

// SpawnForegroundFn builds an unstarted foreground child. Tests replace it.
var SpawnForegroundFn = ForegroundCommand

// RunForegroundServicesFn supervises foreground children for combined `service start`.
var RunForegroundServicesFn = RunForegroundServices

// RunForegroundServices starts each service as an attached child, prefixes
// stdout/stderr with the service name, and waits until all children exit.
func RunForegroundServices(cfg *config.Config, services []string, rawArgs []string) error {
	return runForegroundServices(cfg, services, rawArgs, os.Stdout)
}

func runForegroundServices(cfg *config.Config, services []string, rawArgs []string, out io.Writer) error {
	type child struct {
		service string
		cmd     *exec.Cmd
	}

	built := make([]child, 0, len(services))
	for _, service := range services {
		cmd, err := SpawnForegroundFn(service, rawArgs, cfg)
		if err != nil {
			return fmt.Errorf("%s: %w", service, err)
		}
		built = append(built, child{service: service, cmd: cmd})
	}

	var started []*exec.Cmd
	stopStarted := func() {
		for _, cmd := range started {
			if cmd.Process != nil {
				_ = proc.Terminate(cmd.Process)
			}
		}
	}

	var copyWG sync.WaitGroup
	var writeMu sync.Mutex
	prefixed := func(p []byte) (int, error) {
		writeMu.Lock()
		defer writeMu.Unlock()
		return out.Write(p)
	}

	for _, c := range built {
		stdout, err := c.cmd.StdoutPipe()
		if err != nil {
			stopStarted()
			return fmt.Errorf("stdout pipe for %s: %w", c.service, err)
		}
		stderr, err := c.cmd.StderrPipe()
		if err != nil {
			stopStarted()
			return fmt.Errorf("stderr pipe for %s: %w", c.service, err)
		}
		if err := c.cmd.Start(); err != nil {
			stopStarted()
			return fmt.Errorf("start %s: %w", c.service, err)
		}
		started = append(started, c.cmd)

		copyWG.Add(2)
		go func(service string, r io.Reader) {
			defer copyWG.Done()
			_ = CopyPrefixedLines(service, r, writerFunc(prefixed))
		}(c.service, stdout)
		go func(service string, r io.Reader) {
			defer copyWG.Done()
			_ = CopyPrefixedLines(service, r, writerFunc(prefixed))
		}(c.service, stderr)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, proc.ShutdownSignals...)
	defer signal.Stop(sigChan)

	errCh := make(chan error, len(started))
	for _, cmd := range started {
		go func(cmd *exec.Cmd) {
			errCh <- cmd.Wait()
		}(cmd)
	}

	var firstErr error
	remaining := len(started)
	for remaining > 0 {
		select {
		case <-sigChan:
			stopStarted()
		case err := <-errCh:
			remaining--
			if err != nil && firstErr == nil {
				firstErr = err
				stopStarted()
			}
		}
	}
	copyWG.Wait()
	return firstErr
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// ForegroundCommand builds an attached child process for a service.
// It does not set BEEP_DAEMON_CHILD so the child logs to stdout and its prefixed daily file.
func ForegroundCommand(service string, rawArgs []string, cfg *config.Config) (*exec.Cmd, error) {
	exe, err := osExecutable()
	if err != nil {
		return nil, fmt.Errorf("failed to determine executable path: %w", err)
	}
	cmd := exec.Command(exe, BuildServiceChildArgs(service, rawArgs)...)
	cmd.Env = childServiceEnv(service, cfg, false)
	return cmd, nil
}

func childServiceEnv(service string, cfg *config.Config, daemonChild bool) []string {
	env := append([]string{}, os.Environ()...)
	if daemonChild {
		env = append(env, "BEEP_DAEMON_CHILD=1")
	}
	if service == daemon.ServiceChannel {
		if token := cfg.ChannelAuthToken(); token != "" {
			env = append(env, "BEEP_CHANNEL_TOKEN="+token)
		}
	} else if service == daemon.ServiceRunner && cfg.RunnerToken != "" {
		env = append(env, "BEEP_RUNNER_TOKEN="+cfg.RunnerToken)
	}
	return env
}

// CopyPrefixedLines copies r to w, prefixing each line with "SOURCE | ".
func CopyPrefixedLines(source string, r io.Reader, w io.Writer) error {
	prefix := sourceLabel(source) + " | "
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if _, err := io.WriteString(w, prefix+scanner.Text()+"\n"); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func sourceLabel(source string) string {
	switch source {
	case daemon.ServiceRunner:
		return ui.Cyan(source)
	case daemon.ServiceChannel:
		return ui.Magenta(source)
	default:
		return source
	}
}
