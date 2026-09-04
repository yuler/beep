package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"beep-runner/internal/client"
	"beep-runner/internal/config"
	"beep-runner/internal/daemon"
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
		case "run":
			runDaemon(os.Args[2:])
			return
		case "config":
			runConfigCommand(os.Args[2:])
			return
		case "auth":
			runConfigSet(os.Args[2:])
			return
		case "job", "jobs":
			runJobCommand(os.Args[2:])
			return
		case "sync":
			runJobSync(os.Args[2:])
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
	fmt.Println("A runner executes scheduled jobs locally in your workspace and reports logs/results to Core.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  beep-runner [command] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run                  Start the runner daemon (default)")
	fmt.Println("  config               Show current configuration")
	fmt.Println("  config set           Set configuration parameters (server, token, allow-exec, etc.)")
	fmt.Println("  config unset <key>   Remove a configuration parameter")
	fmt.Println("  config path          Print path to config file")
	fmt.Println("  ping                 Test connectivity and authentication with Beep Core")
	fmt.Println("  job create <slug>    Create a local job script and sync to server")
	fmt.Println("  job sync             Sync all local workspace jobs to Beep Core")
	fmt.Println("  job list             List local jobs in workspace and on server")
	fmt.Println("  version              Print version information")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --server             Beep server URL (env: BEEP_SERVER)")
	fmt.Println("  --token              Runner token (env: BEEP_RUNNER_TOKEN)")
	fmt.Println("  --workspace          Local job workspace (env: BEEP_WORKSPACE, default ~/.beep-runner)")
	fmt.Println("  --allow-exec         Allow running workspace scripts (env: BEEP_ALLOW_EXEC)")
	fmt.Println("  --concurrency        Max concurrent jobs (default 5)")
	fmt.Println("  --poll-interval      Poll interval (default 3s)")
}

func peekWorkspace(args []string) string {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--workspace=") {
			return strings.TrimPrefix(arg, "--workspace=")
		}
		if arg == "--workspace" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func parseFlags(args []string) (*config.Config, error) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
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

func runConfigCommand(args []string) {
	if len(args) == 0 {
		runConfigShow(nil)
		return
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "show", "get", "list", "ls":
		runConfigShow(rest)
	case "set", "save":
		runConfigSet(rest)
	case "unset", "rm", "delete":
		runConfigUnset(rest)
	case "path", "file":
		wsHint := peekWorkspace(rest)
		fmt.Println(config.GetConfigPath(wsHint))
	default:
		if strings.HasPrefix(sub, "-") {
			runConfigShow(args)
		} else {
			runConfigSet(args)
		}
	}
}

func runConfigShow(args []string) {
	var showToken bool
	var workspaceFlag string
	fs := flag.NewFlagSet("config show", flag.ExitOnError)
	fs.BoolVar(&showToken, "show-token", false, "Display unmasked runner token")
	fs.StringVar(&workspaceFlag, "workspace", "", "Workspace directory")
	wsHint := peekWorkspace(args)
	if err := fs.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}
	if workspaceFlag != "" {
		wsHint = workspaceFlag
	}

	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	tokenStr := config.MaskToken(cfg.RunnerToken)
	if showToken && cfg.RunnerToken != "" {
		tokenStr = cfg.RunnerToken
	}

	fmt.Println("Beep Runner Configuration:")
	fmt.Printf("  Config File:    %s\n", cfg.ConfigFile)
	fmt.Printf("  Server URL:     %s\n", cfg.ServerURL)
	fmt.Printf("  Runner Token:   %s\n", tokenStr)
	fmt.Printf("  Workspace:      %s\n", cfg.Workspace)
	fmt.Printf("  Allow Exec:     %t\n", cfg.AllowExec)
	fmt.Printf("  Concurrency:    %d\n", cfg.Concurrency)
	fmt.Printf("  Poll Interval:  %s\n", cfg.PollInterval)
	fmt.Printf("  Hostname:       %s\n", cfg.Hostname)
}

func runConfigSet(args []string) {
	wsHint := peekWorkspace(args)
	configPath := config.GetConfigPath(wsHint)

	fc, err := config.LoadFile(configPath)
	if err != nil {
		fc = &config.FileConfig{}
	}

	boolFlags := map[string]bool{"allow-exec": true}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	var (
		serverFlag       string
		tokenFlag        string
		workspaceFlag    string
		allowExecFlag    bool
		concurrencyFlag  int
		pollIntervalFlag string
	)

	fs := flag.NewFlagSet("config set", flag.ExitOnError)
	fs.StringVar(&serverFlag, "server", "", "Beep server URL")
	fs.StringVar(&tokenFlag, "token", "", "Runner token")
	fs.StringVar(&workspaceFlag, "workspace", "", "Workspace directory")
	fs.BoolVar(&allowExecFlag, "allow-exec", false, "Allow local script execution")
	fs.IntVar(&concurrencyFlag, "concurrency", 0, "Max concurrency")
	fs.StringVar(&pollIntervalFlag, "poll-interval", "", "Poll interval (e.g. 3s)")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("Error: %v", err)
	}

	updated := false

	// Handle flag updates
	if serverFlag != "" {
		fc.ServerURL = strings.TrimRight(serverFlag, "/")
		updated = true
	}
	if tokenFlag != "" {
		fc.RunnerToken = tokenFlag
		updated = true
	}
	if workspaceFlag != "" {
		fc.Workspace = workspaceFlag
		updated = true
	}
	if allowExecFlag {
		t := true
		fc.AllowExec = &t
		updated = true
	}
	if concurrencyFlag > 0 {
		fc.Concurrency = concurrencyFlag
		updated = true
	}
	if pollIntervalFlag != "" {
		fc.PollInterval = pollIntervalFlag
		updated = true
	}

	// Handle positional key-value pairs (e.g. beep-runner config set server https://...)
	for i := 0; i < len(positional); i++ {
		key := strings.ToLower(positional[i])
		if i+1 < len(positional) {
			val := positional[i+1]
			switch key {
			case "server", "server_url", "server-url", "url":
				fc.ServerURL = strings.TrimRight(val, "/")
				updated = true
				i++
			case "token", "runner_token", "runner-token", "auth":
				fc.RunnerToken = val
				updated = true
				i++
			case "workspace", "dir", "workdir":
				fc.Workspace = val
				updated = true
				i++
			case "allow_exec", "allow-exec", "exec":
				v, _ := strconv.ParseBool(val)
				fc.AllowExec = &v
				updated = true
				i++
			case "concurrency":
				if n, err := strconv.Atoi(val); err == nil && n > 0 {
					fc.Concurrency = n
					updated = true
					i++
				}
			case "poll_interval", "poll-interval", "interval":
				fc.PollInterval = val
				updated = true
				i++
			}
		} else {
			// Single key like "allow-exec"
			if key == "allow_exec" || key == "allow-exec" {
				t := true
				fc.AllowExec = &t
				updated = true
			}
		}
	}

	if !updated {
		fmt.Println("No configuration options provided.")
		fmt.Println("\nUsage:")
		fmt.Println("  beep-runner config set --server <url> --token <token> [--allow-exec]")
		fmt.Println("  beep-runner config set server <url>")
		fmt.Println("  beep-runner config set token <token>")
		fmt.Println("  beep-runner config set allow-exec true")
		return
	}

	if err := config.SaveFile(configPath, fc); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Printf("✓ Saved configuration to %s\n\n", configPath)
	if wsHint != "" {
		runConfigShow([]string{"--workspace=" + wsHint})
	} else {
		runConfigShow(nil)
	}
}

func runConfigUnset(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: beep-runner config unset <key>")
		fmt.Println("Keys: server, token, workspace, allow-exec, concurrency, poll-interval")
		return
	}

	wsHint := peekWorkspace(args)
	configPath := config.GetConfigPath(wsHint)

	fc, err := config.LoadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	key := strings.ToLower(args[0])
	switch key {
	case "server", "server_url", "server-url":
		fc.ServerURL = ""
	case "token", "runner_token", "runner-token":
		fc.RunnerToken = ""
	case "workspace", "dir":
		fc.Workspace = ""
	case "allow_exec", "allow-exec", "exec":
		fc.AllowExec = nil
	case "concurrency":
		fc.Concurrency = 0
	case "poll_interval", "poll-interval":
		fc.PollInterval = ""
	default:
		log.Fatalf("Unknown config key %q", key)
	}

	if err := config.SaveFile(configPath, fc); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Printf("✓ Unset %s in %s\n\n", key, configPath)
	if wsHint != "" {
		runConfigShow([]string{"--workspace=" + wsHint})
	} else {
		runConfigShow(nil)
	}
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

func runJobCommand(args []string) {
	if len(args) == 0 {
		printJobUsage()
		return
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "create", "new", "add":
		runJobCreate(rest)
	case "sync":
		runJobSync(rest)
	case "list", "ls":
		runJobList(rest)
	default:
		// If user typed: beep-runner job <slug> -> treat as create <slug>
		if !strings.HasPrefix(sub, "-") {
			runJobCreate(args)
		} else {
			printJobUsage()
		}
	}
}

func printJobUsage() {
	fmt.Println("Usage:")
	fmt.Println("  beep-runner job create <slug> [flags]   Create local script & register job on server")
	fmt.Println("  beep-runner job sync [flags]            Sync all local jobs in workspace to server")
	fmt.Println("  beep-runner job list [flags]            List local and server jobs")
	fmt.Println()
	fmt.Println("Flags for job create:")
	fmt.Println("  --name        Job name (default: humanized slug)")
	fmt.Println("  --cron        Cron expression (default: */5 * * * *)")
	fmt.Println("  --type        Script type: sh, py, js, rb (default: sh)")
	fmt.Println("  --timeout     Timeout in seconds (default: 30)")
	fmt.Println("  --no-sync     Create local script only without syncing to server")
}

func splitFlagsAndArgs(args []string, boolFlags map[string]bool) (flagArgs []string, positional []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			name := strings.TrimLeft(strings.Split(arg, "=")[0], "-")
			if !strings.Contains(arg, "=") && !boolFlags[name] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			positional = append(positional, arg)
		}
	}
	return flagArgs, positional
}

func runJobCreate(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	boolFlags := map[string]bool{"no-sync": true}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	fs := flag.NewFlagSet("job create", flag.ExitOnError)
	var (
		name           string
		cron           string
		scriptType     string
		timeoutSeconds int
		noSync         bool
	)
	fs.StringVar(&name, "name", "", "Display name of the job")
	fs.StringVar(&cron, "cron", "*/5 * * * *", "Cron schedule expression")
	fs.StringVar(&scriptType, "type", "sh", "Script type (sh, py, js, rb)")
	fs.IntVar(&timeoutSeconds, "timeout", 30, "Timeout in seconds")
	fs.BoolVar(&noSync, "no-sync", false, "Do not sync to server")
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("Error: %v", err)
	}

	var slug string
	if len(positional) > 0 {
		slug = positional[0]
	}
	if slug == "" {
		fmt.Println("Error: job slug is required.")
		fmt.Println("Example: beep-runner job create intranet-http --cron '*/5 * * * *'")
		os.Exit(1)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("Workspace error: %v", err)
	}

	filePath, created, err := ws.CreateScript(slug, scriptType)
	if err != nil {
		log.Fatalf("Failed to create script: %v", err)
	}

	if created {
		fmt.Printf("✓ Created local script: %s\n", filePath)
	} else {
		fmt.Printf("ℹ Local script already exists: %s\n", filePath)
	}

	if noSync {
		fmt.Println("Skipping server sync (--no-sync specified).")
		return
	}

	if cfg.ServerURL == "" || cfg.RunnerToken == "" {
		fmt.Println("ℹ Server URL or Runner Token not configured; skipping server sync.")
		fmt.Println("  Tip: Configure once with 'beep-runner config set --server <url> --token <token>'")
		return
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverJob, err := c.CreateJob(ctx, &client.CreateJobRequest{
		Slug:           slug,
		Name:           name,
		Cron:           cron,
		TimeoutSeconds: timeoutSeconds,
	})
	if err != nil {
		log.Fatalf("Failed to sync job to server: %v", err)
	}

	fmt.Printf("✓ Successfully registered job on server: %s (ID: %s, Cron: %s)\n", serverJob.Name, serverJob.ID, serverJob.Cron)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Edit your script: %s\n", filePath)
	fmt.Printf("  2. Start runner daemon: beep-runner run --allow-exec\n")
}

func runJobSync(args []string) {
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

	localJobs, err := ws.ListJobs()
	if err != nil {
		log.Fatalf("Failed to list local jobs: %v", err)
	}

	if len(localJobs) == 0 {
		fmt.Printf("No local jobs found in %s/jobs\n", ws.Root)
		fmt.Println("Create one with: beep-runner job create <slug>")
		return
	}

	var syncReqs []*client.CreateJobRequest
	for _, job := range localJobs {
		syncReqs = append(syncReqs, &client.CreateJobRequest{
			Slug: job.Slug,
			Cron: "*/5 * * * *",
		})
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	synced, err := c.SyncJobs(ctx, syncReqs)
	if err != nil {
		log.Fatalf("Failed to sync jobs: %v", err)
	}

	fmt.Printf("✓ Successfully synced %d jobs to server:\n", len(synced))
	for _, j := range synced {
		fmt.Printf("  - %s (slug: %s, cron: %s)\n", j.Name, j.Slug, j.Cron)
	}
}

func runJobList(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fs := flag.NewFlagSet("job list", flag.ExitOnError)
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("Workspace error: %v", err)
	}

	localJobs, err := ws.ListJobs()
	if err != nil {
		log.Fatalf("Failed to list local jobs: %v", err)
	}

	fmt.Printf("Local Workspace Jobs (%s):\n", ws.Root)
	if len(localJobs) == 0 {
		fmt.Println("  (No local jobs found)")
	} else {
		for _, j := range localJobs {
			if j.FilePath != "" {
				fmt.Printf("  • %-20s -> %s\n", j.Slug, j.FilePath)
			} else {
				fmt.Printf("  • %-20s -> %s (jobs.json)\n", j.Slug, strings.Join(j.Command, " "))
			}
		}
	}

	if cfg.ServerURL != "" && cfg.RunnerToken != "" {
		fmt.Printf("\nServer Jobs (%s):\n", cfg.ServerURL)
		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		serverJobs, err := c.ListJobs(ctx)
		if err != nil {
			fmt.Printf("  (Could not fetch server jobs: %v)\n", err)
		} else if len(serverJobs) == 0 {
			fmt.Println("  (No jobs configured on server)")
		} else {
			for _, sj := range serverJobs {
				fmt.Printf("  • %-20s [%s] cron: %-12s (status: %s)\n", sj.Slug, sj.Name, sj.Cron, sj.Status)
			}
		}
	}
}
