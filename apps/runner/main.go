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
	"beep-runner/internal/probe"
	"beep-runner/internal/version"
	"beep-runner/internal/workspace"
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

	runDaemon(os.Args[1:])
}

func printUsage() {
	fmt.Println("Beep self-hosted runner")
	fmt.Println()
	fmt.Println("The runner is a local workspace. Core schedules jobs; this process polls,")
	fmt.Println("runs a matching local script, and posts logs/results back to the server.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  beep-runner [command] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run               Start the runner daemon (default)")
	fmt.Println("  ping              Test connectivity and authentication")
	fmt.Println("  test <type> <arg> Local helper (http, tls, tcp, dns) for writing scripts")
	fmt.Println("  version           Print version")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --server          Beep server URL (env: BEEP_SERVER)")
	fmt.Println("  --token           Runner token (env: BEEP_RUNNER_TOKEN)")
	fmt.Println("  --workspace       Local job workspace (env: BEEP_WORKSPACE, default ~/.beep-runner)")
	fmt.Println("  --allow-exec      Allow running workspace scripts (env: BEEP_ALLOW_EXEC)")
	fmt.Println("  --concurrency     Max concurrent jobs (default 5)")
	fmt.Println("  --poll-interval   Poll interval (default 3s)")
}

func parseFlags(args []string) (*config.Config, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("beep-runner", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Local workspace directory")
	fs.BoolVar(&cfg.AllowExec, "allow-exec", cfg.AllowExec, "Allow local scripts")
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

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("Workspace error: %v", err)
	}

	d := daemon.New(cfg, ws)

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
		fmt.Println("Usage: beep-runner test <http|tls|tcp|dns> <target>")
		os.Exit(1)
	}

	probeType := args[0]
	target := args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var out any
	switch probeType {
	case "http", "https":
		out = probe.NewHTTPProbe().Run(ctx, map[string]any{"target_url": target})
	case "tls", "ssl":
		out = probe.NewTLSProbe().Run(ctx, map[string]any{"host": target})
	case "tcp":
		out = probe.NewTCPProbe().Run(ctx, map[string]any{"host": target, "port": 80})
	case "dns":
		out = probe.NewDNSProbe().Run(ctx, map[string]any{"hostname": target})
	default:
		fmt.Printf("Unknown helper type: %s\n", probeType)
		os.Exit(1)
	}

	encoded, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(encoded))
}
