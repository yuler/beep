package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/config"
	"beep-runner/internal/daemon"
	"beep-runner/internal/exec"
	"beep-runner/internal/probe"
	"beep-runner/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("beep-runner version %s (%s, %s)\n", version.Version, version.GitCommit, version.BuildDate)
			return

		case "ping":
			runPing(os.Args[2:])
			return

		case "test":
			runTest(os.Args[2:])
			return

		case "run":
			runDaemon(os.Args[2:])
			return

		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	// Default to run daemon
	runDaemon(os.Args[1:])
}

func printUsage() {
	fmt.Println("Beep Self-hosted Runner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  beep-runner [command] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run               Start the runner daemon (default)")
	fmt.Println("  ping              Test connectivity and authentication with Beep server")
	fmt.Println("  test <type> <arg> Test a probe offline (http, tls, tcp, dns, exec)")
	fmt.Println("  version           Print runner version information")
	fmt.Println("  help              Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --server          Beep server URL (env: BEEP_SERVER)")
	fmt.Println("  --token           Runner authentication token (env: BEEP_RUNNER_TOKEN)")
	fmt.Println("  --allow-exec      Enable local command/script execution (env: BEEP_ALLOW_EXEC)")
	fmt.Println("  --concurrency     Maximum concurrent tasks (default 5, env: BEEP_CONCURRENCY)")
	fmt.Println("  --poll-interval   Polling interval (default 3s, env: BEEP_POLL_INTERVAL)")
}

func parseFlags(args []string) (*config.Config, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("beep-runner", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.BoolVar(&cfg.AllowExec, "allow-exec", cfg.AllowExec, "Allow local script execution")
	fs.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "Max concurrency")
	fs.DurationVar(&cfg.PollInterval, "poll-interval", cfg.PollInterval, "Poll interval")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return cfg, nil
}

func runDaemon(args []string) {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	d := daemon.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[beep-runner] Received termination signal...")
		cancel()
	}()

	if err := d.Start(ctx); err != nil {
		log.Fatalf("Runner daemon error: %v", err)
	}
}

func runPing(args []string) {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.Ping(ctx)
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	fmt.Println("Ping successful!")
	fmt.Printf("  Runner ID:   %s\n", res.RunnerID)
	fmt.Printf("  Runner Name: %s\n", res.RunnerName)
	fmt.Printf("  Server Time: %s\n", res.ServerTime)
}

func runTest(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: beep-runner test <http|tls|tcp|dns|exec> <target>")
		os.Exit(1)
	}

	probeType := args[0]
	target := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var sig *probe.Signal

	switch probeType {
	case "http", "https":
		p := probe.NewHTTPProbe()
		sig = p.Run(ctx, map[string]any{"target_url": target})

	case "tls", "ssl":
		p := probe.NewTLSProbe()
		sig = p.Run(ctx, map[string]any{"host": target})

	case "tcp":
		p := probe.NewTCPProbe()
		sig = p.Run(ctx, map[string]any{"host": target, "port": 80})

	case "dns":
		p := probe.NewDNSProbe()
		sig = p.Run(ctx, map[string]any{"hostname": target})

	case "exec", "script", "shell":
		e := exec.NewScriptExecutor(true)
		sig = e.Run(ctx, map[string]any{"command": target})

	default:
		fmt.Printf("Unknown probe test type: %s\n", probeType)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(sig, "", "  ")
	fmt.Println(string(out))
}
