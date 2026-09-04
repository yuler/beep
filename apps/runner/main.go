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
	"beep-runner/internal/ui"
	"beep-runner/internal/version"
	"beep-runner/internal/workspace"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("%s version %s (%s, %s)\n", ui.Bold(ui.Cyan("beep-runner")), ui.Bold(version.Version), ui.Dim(version.GitCommit), ui.Dim(version.BuildDate))
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
		case "push":
			runJobPush(os.Args[2:])
			return
		case "pull":
			runJobPull(os.Args[2:])
			return
		case "sync":
			runJobPush(os.Args[2:])
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	runDaemon(os.Args[1:])
}

func printUsage() {
	fmt.Println(ui.Bold(ui.Cyan("Beep self-hosted runner")))
	fmt.Println()
	fmt.Println("A runner executes scheduled jobs locally in your workspace and reports logs/results to Core.")
	fmt.Println()
	fmt.Println(ui.Section("Usage:"))
	fmt.Printf("  %s %s %s\n", ui.Bold("beep-runner"), ui.Green("[command]"), ui.Cyan("[flags]"))
	fmt.Println()
	fmt.Println(ui.Section("Commands:"))
	fmt.Printf("  %-24s %s\n", ui.Green("run"), "Start the runner daemon (default)")
	fmt.Printf("  %-24s %s\n", ui.Green("config"), "Show current configuration")
	fmt.Printf("  %-24s %s\n", ui.Green("config set"), "Set configuration parameters (server, token, etc.)")
	fmt.Printf("  %-24s %s\n", ui.Green("config unset <key>"), "Remove a configuration parameter")
	fmt.Printf("  %-24s %s\n", ui.Green("config path"), "Print path to config file")
	fmt.Printf("  %-24s %s\n", ui.Green("ping"), "Test connectivity and authentication with Beep Core")
	fmt.Printf("  %-24s %s\n", ui.Green("job create <slug>"), "Create a local job script scaffold")
	fmt.Printf("  %-24s %s\n", ui.Green("job push [slug]"), "Push local workspace job(s) to Beep Core (like git push)")
	fmt.Printf("  %-24s %s\n", ui.Green("job pull [slug]"), "Pull server job(s) to local workspace (like git pull)")
	fmt.Printf("  %-24s %s\n", ui.Green("job remove <slug>"), "Remove a local job script and delete from server")
	fmt.Printf("  %-24s %s\n", ui.Green("job list"), "List local jobs in workspace and on server")
	fmt.Printf("  %-24s %s\n", ui.Green("version"), "Print version information")
	fmt.Println()
	fmt.Println(ui.Section("Flags:"))
	fmt.Printf("  %-24s %s\n", ui.Cyan("--server"), "Beep server URL (env: BEEP_SERVER)")
	fmt.Printf("  %-24s %s\n", ui.Cyan("--token"), "Runner token (env: BEEP_RUNNER_TOKEN)")
	fmt.Printf("  %-24s %s\n", ui.Cyan("--workspace"), "Local job workspace (env: BEEP_WORKSPACE, default ~/.beep-runner)")
	fmt.Printf("  %-24s %s\n", ui.Cyan("--concurrency"), "Max concurrent jobs (default 5)")
	fmt.Printf("  %-24s %s\n", ui.Cyan("--poll-interval"), "Poll interval (default 3s)")
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
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}
	if workspaceFlag != "" {
		wsHint = workspaceFlag
	}

	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error loading config: %v", err))
	}

	tokenStr := config.MaskToken(cfg.RunnerToken)
	if showToken && cfg.RunnerToken != "" {
		tokenStr = cfg.RunnerToken
	}

	fmt.Println(ui.Bold(ui.Cyan("Beep Runner Configuration:")))
	fmt.Println(ui.KeyValue("Config File", ui.Dim(cfg.ConfigFile)))
	fmt.Println(ui.KeyValue("Server URL", ui.Bold(cfg.ServerURL)))
	fmt.Println(ui.KeyValue("Runner Token", ui.Yellow(tokenStr)))
	fmt.Println(ui.KeyValue("Workspace", ui.Dim(cfg.Workspace)))
	fmt.Println(ui.KeyValue("Concurrency", ui.Bold(strconv.Itoa(cfg.Concurrency))))
	fmt.Println(ui.KeyValue("Poll Interval", ui.Bold(cfg.PollInterval.String())))
	fmt.Println(ui.KeyValue("Hostname", ui.Dim(cfg.Hostname)))
}

func runConfigSet(args []string) {
	wsHint := peekWorkspace(args)
	configPath := config.GetConfigPath(wsHint)

	fc, err := config.LoadFile(configPath)
	if err != nil {
		fc = &config.FileConfig{}
	}

	boolFlags := map[string]bool{}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	var (
		serverFlag       string
		tokenFlag        string
		workspaceFlag    string
		concurrencyFlag  int
		pollIntervalFlag string
	)

	fs := flag.NewFlagSet("config set", flag.ExitOnError)
	fs.StringVar(&serverFlag, "server", "", "Beep server URL")
	fs.StringVar(&tokenFlag, "token", "", "Runner token")
	fs.StringVar(&workspaceFlag, "workspace", "", "Workspace directory")
	fs.IntVar(&concurrencyFlag, "concurrency", 0, "Max concurrency")
	fs.StringVar(&pollIntervalFlag, "poll-interval", "", "Poll interval (e.g. 3s)")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
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
		}
	}

	if !updated {
		fmt.Println(ui.Warn("No configuration options provided."))
		fmt.Println()
		fmt.Println(ui.Section("Usage:"))
		fmt.Printf("  %s\n", ui.Cyan("beep-runner config set --server <url> --token <token>"))
		fmt.Printf("  %s\n", ui.Cyan("beep-runner config set server <url>"))
		fmt.Printf("  %s\n", ui.Cyan("beep-runner config set token <token>"))
		return
	}

	if err := config.SaveFile(configPath, fc); err != nil {
		log.Fatalf("%s", ui.Error("Failed to save config: %v", err))
	}

	fmt.Println(ui.Success("Saved configuration to %s", ui.Bold(configPath)))
	fmt.Println()
	if wsHint != "" {
		runConfigShow([]string{"--workspace=" + wsHint})
	} else {
		runConfigShow(nil)
	}
}

func runConfigUnset(args []string) {
	if len(args) == 0 {
		fmt.Println(ui.Section("Usage:"))
		fmt.Printf("  %s\n", ui.Cyan("beep-runner config unset <key>"))
		fmt.Printf("  %s: server, token, workspace, concurrency, poll-interval\n", ui.Dim("Keys"))
		return
	}

	wsHint := peekWorkspace(args)
	configPath := config.GetConfigPath(wsHint)

	fc, err := config.LoadFile(configPath)
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to load config file: %v", err))
	}

	key := strings.ToLower(args[0])
	switch key {
	case "server", "server_url", "server-url":
		fc.ServerURL = ""
	case "token", "runner_token", "runner-token":
		fc.RunnerToken = ""
	case "workspace", "dir":
		fc.Workspace = ""
	case "concurrency":
		fc.Concurrency = 0
	case "poll_interval", "poll-interval":
		fc.PollInterval = ""
	default:
		log.Fatalf("%s", ui.Error("Unknown config key %q", key))
	}

	if err := config.SaveFile(configPath, fc); err != nil {
		log.Fatalf("%s", ui.Error("Failed to save config: %v", err))
	}

	fmt.Println(ui.Success("Unset %s in %s", ui.Bold(key), ui.Dim(configPath)))
	fmt.Println()
	if wsHint != "" {
		runConfigShow([]string{"--workspace=" + wsHint})
	} else {
		runConfigShow(nil)
	}
}

func runDaemon(args []string) {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", ui.Error("Configuration error: %v", err))
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	d := daemon.New(cfg, ws)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println(ui.Dim("[beep-runner] Received termination signal..."))
		cancel()
	}()

	if err := d.Start(ctx); err != nil {
		log.Fatalf("%s", ui.Error("Runner daemon error: %v", err))
	}
}

func runPing(args []string) {
	cfg, err := parseFlags(args)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", ui.Error("Configuration error: %v", err))
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.Ping(ctx)
	if err != nil {
		log.Fatalf("%s", ui.Error("Ping failed: %v", err))
	}

	fmt.Println(ui.Success("Ping successful!"))
	fmt.Println(ui.KeyValue("Runner ID", ui.Bold(res.RunnerID)))
	fmt.Println(ui.KeyValue("Runner Name", ui.Bold(res.RunnerName)))
	fmt.Println(ui.KeyValue("Server Time", ui.Dim(res.ServerTime)))
}

func runJobCommand(args []string) {
	if len(args) == 0 {
		printJobUsage()
		return
	}

	sub := strings.ToLower(args[0])
	rest := args[1:]

	switch sub {
	case "create", "new", "add":
		runJobCreate(rest)
	case "push":
		runJobPush(rest)
	case "pull":
		runJobPull(rest)
	case "sync":
		runJobPush(rest)
	case "remove", "rm", "delete", "del":
		runJobRemove(rest)
	case "list", "ls":
		runJobList(rest)
	case "help", "-h", "--help":
		printJobUsage()
	default:
		fmt.Fprintln(os.Stderr, ui.Error("unknown job command %q.\n", args[0]))
		printJobUsage()
		os.Exit(1)
	}
}

func printJobUsage() {
	fmt.Println(ui.Section("Usage:"))
	fmt.Printf("  %s %s %s   %s\n", ui.Bold("beep-runner"), ui.Green("job create <slug>"), ui.Cyan("[flags]"), "Create local script scaffold")
	fmt.Printf("  %s %s %s     %s\n", ui.Bold("beep-runner"), ui.Green("job push [slug]"), ui.Cyan("[flags]"), "Push local workspace job(s) to Beep Core (like git push)")
	fmt.Printf("  %s %s %s     %s\n", ui.Bold("beep-runner"), ui.Green("job pull [slug]"), ui.Cyan("[flags]"), "Pull server job(s) to local workspace (like git pull)")
	fmt.Printf("  %s %s %s   %s\n", ui.Bold("beep-runner"), ui.Green("job remove <slug>"), ui.Cyan("[flags]"), "Remove local script & delete job from server")
	fmt.Printf("  %s %s %s            %s\n", ui.Bold("beep-runner"), ui.Green("job list"), ui.Cyan("[flags]"), "List local and server jobs")
	fmt.Println()
	fmt.Println(ui.Section("Flags for job create:"))
	fmt.Printf("  %-16s %s\n", ui.Cyan("--name"), "Job name (default: humanized slug or script @name)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--cron"), "Cron expression (default: */5 * * * * or script @schedule)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--timezone"), "Timezone (default: system timezone or script @timezone)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--description"), "Job description (default: script @description)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--type"), "Script type: sh, py, js, rb (default: sh)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--timeout"), "Timeout in seconds (default: 30 or script @timeout)")
	fmt.Printf("  %-16s %s\n", ui.Cyan("--no-sync"), "Create local script only without syncing to server")
	fmt.Println()
	fmt.Println(ui.Section("Flags for job pull:"))
	fmt.Printf("  %-16s %s\n", ui.Cyan("--force"), "Overwrite existing local scripts with server definition")
	fmt.Println()
	fmt.Println(ui.Section("Flags for job remove:"))
	fmt.Printf("  %-16s %s\n", ui.Cyan("--no-sync"), "Remove local script only without deleting from server")
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
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	boolFlags := map[string]bool{"no-sync": true}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	fs := flag.NewFlagSet("job create", flag.ExitOnError)
	var (
		name           string
		cron           string
		timezone       string
		description    string
		scriptType     string
		timeoutSeconds int
		noSync         bool
	)
	fs.StringVar(&name, "name", "", "Display name of the job")
	fs.StringVar(&cron, "cron", "*/5 * * * *", "Cron schedule expression")
	fs.StringVar(&timezone, "timezone", "", "Timezone (e.g. Asia/Shanghai, UTC)")
	fs.StringVar(&timezone, "tz", "", "Timezone alias")
	fs.StringVar(&description, "description", "", "Description of the job")
	fs.StringVar(&description, "desc", "", "Description alias")
	fs.StringVar(&scriptType, "type", "sh", "Script type (sh, py, js, rb)")
	fs.IntVar(&timeoutSeconds, "timeout", 30, "Timeout in seconds")
	fs.BoolVar(&noSync, "no-sync", false, "Do not sync to server")
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	var slug string
	if len(positional) > 0 {
		slug = positional[0]
	}
	if slug == "" {
		fmt.Println(ui.Error("job slug is required."))
		fmt.Printf("Example: %s\n", ui.Cyan("beep-runner job create intranet-http --cron '*/5 * * * *'"))
		os.Exit(1)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	filePath, created, err := ws.CreateScript(slug, scriptType, name, cron, timezone, description, timeoutSeconds)
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to create script: %v", err))
	}

	if created {
		fmt.Println(ui.Success("Created local script: %s", ui.Cyan(filePath)))
	} else {
		fmt.Println(ui.Info("Local script already exists: %s", ui.Cyan(filePath)))
		if meta, metaErr := workspace.ParseScriptMetadata(filePath); metaErr == nil {
			if name == "" && meta.Name != "" {
				name = meta.Name
			}
			if cron == "*/5 * * * *" && meta.Cron != "" {
				cron = meta.Cron
			}
			if timezone == "" && meta.Timezone != "" {
				timezone = meta.Timezone
			}
			if description == "" && meta.Description != "" {
				description = meta.Description
			}
			if timeoutSeconds == 30 && meta.TimeoutSeconds > 0 {
				timeoutSeconds = meta.TimeoutSeconds
			}
		}
	}

	if noSync {
		fmt.Println(ui.Dim("Skipping server sync (--no-sync specified)."))
		return
	}

	if cfg.ServerURL == "" || cfg.RunnerToken == "" {
		fmt.Println(ui.Info("Server URL or Runner Token not configured; skipping server sync."))
		fmt.Printf("  %s %s\n", ui.Dim("Tip: Configure once with"), ui.Cyan("beep-runner config set --server <url> --token <token>"))
		return
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverJob, err := c.CreateJob(ctx, &client.CreateJobRequest{
		Slug:           slug,
		Name:           name,
		Cron:           cron,
		Timezone:       timezone,
		TimeoutSeconds: timeoutSeconds,
		Description:    description,
	})
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to sync job to server: %v", err))
	}

	fmt.Println(ui.Success("Successfully registered job on server: %s (%s: %s, %s: %s)",
		ui.Bold(serverJob.Name), ui.Dim("ID"), ui.Dim(serverJob.ID), ui.Dim("Cron"), ui.Yellow(serverJob.Cron)))
	fmt.Println()
	fmt.Println(ui.Section("Next steps:"))
	fmt.Printf("  1. Edit your script: %s\n", ui.Cyan(filePath))
	fmt.Printf("  2. Start runner daemon: %s\n", ui.Green("beep-runner run"))
}

func runJobRemove(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	boolFlags := map[string]bool{"no-sync": true}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	fs := flag.NewFlagSet("job remove", flag.ExitOnError)
	var noSync bool
	fs.BoolVar(&noSync, "no-sync", false, "Do not delete from server")
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	var slug string
	if len(positional) > 0 {
		slug = positional[0]
	}
	if slug == "" {
		fmt.Println(ui.Error("job slug is required."))
		fmt.Printf("Example: %s\n", ui.Cyan("beep-runner job remove intranet-http"))
		os.Exit(1)
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	removedFiles, removedFromJSON, err := ws.RemoveJob(slug)
	if err != nil {
		fmt.Println(ui.Info("%v", err))
	} else {
		for _, f := range removedFiles {
			fmt.Println(ui.Success("Removed local job script: %s", ui.Cyan(f)))
		}
		if removedFromJSON {
			fmt.Println(ui.Success("Removed %q from jobs.json", ui.Bold(slug)))
		}
	}

	if noSync {
		fmt.Println(ui.Dim("Skipping server delete (--no-sync specified)."))
		return
	}

	if cfg.ServerURL != "" && cfg.RunnerToken != "" {
		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := c.DeleteJob(ctx, slug); err != nil {
			fmt.Println(ui.Warn("Failed to delete job on server: %v", err))
		} else {
			fmt.Println(ui.Success("Deleted job %q from server", ui.Bold(slug)))
		}
	}
}

func runJobPush(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	boolFlags := map[string]bool{}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	fs := flag.NewFlagSet("job push", flag.ExitOnError)
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", ui.Error("Configuration error: %v", err))
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	localJobs, err := ws.ListJobs()
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to list local jobs: %v", err))
	}

	var targetSlug string
	if len(positional) > 0 {
		targetSlug = strings.TrimSpace(strings.ToLower(positional[0]))
	}

	var syncReqs []*client.CreateJobRequest
	for _, job := range localJobs {
		if targetSlug != "" && !strings.EqualFold(job.Slug, targetSlug) {
			continue
		}
		syncReqs = append(syncReqs, &client.CreateJobRequest{
			Slug:           job.Slug,
			Name:           job.Name,
			Cron:           job.Cron,
			Timezone:       job.Timezone,
			TimeoutSeconds: job.TimeoutSeconds,
			Description:    job.Description,
		})
	}

	if len(syncReqs) == 0 {
		if targetSlug != "" {
			log.Fatalf("%s", ui.Error("Job %q not found in local workspace %s", targetSlug, ws.Root))
		} else {
			fmt.Println(ui.Warn("No local jobs found in %s/jobs", ws.Root))
			fmt.Printf("Create one with: %s\n", ui.Cyan("beep-runner job create <slug>"))
			return
		}
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	synced, err := c.SyncJobs(ctx, syncReqs)
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to push jobs to server: %v", err))
	}

	fmt.Println(ui.Success("Successfully pushed %d job(s) to server (%s):", len(synced), ui.Dim(cfg.ServerURL)))
	for _, j := range synced {
		tz := j.Timezone
		if tz == "" {
			tz = "UTC"
		}
		fmt.Printf("  %s %s [%s, %s: %s] (%s)\n",
			ui.Bullet(), ui.Cyan(j.Slug), ui.Bold(j.Name), ui.Dim("cron"), ui.Yellow(j.Cron), ui.Dim(tz))
	}
}

func runJobPull(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	boolFlags := map[string]bool{"force": true}
	flagArgs, positional := splitFlagsAndArgs(args, boolFlags)

	var force bool
	fs := flag.NewFlagSet("job pull", flag.ExitOnError)
	fs.BoolVar(&force, "force", false, "Overwrite existing local scripts")
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(flagArgs); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", ui.Error("Configuration error: %v", err))
	}

	var targetSlug string
	if len(positional) > 0 {
		targetSlug = strings.TrimSpace(strings.ToLower(positional[0]))
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	c := client.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	serverJobs, err := c.ListJobs(ctx)
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to fetch server jobs: %v", err))
	}

	if len(serverJobs) == 0 {
		fmt.Println(ui.Info("No jobs registered on server (%s)", cfg.ServerURL))
		return
	}

	pulledCount := 0
	skippedCount := 0

	for _, sj := range serverJobs {
		if targetSlug != "" && !strings.EqualFold(sj.Slug, targetSlug) {
			continue
		}

		desc := ""
		if sj.Config != nil {
			if d, ok := sj.Config["description"].(string); ok {
				desc = d
			}
		}

		filePath, created, pullErr := ws.PullJob(sj.Slug, "sh", sj.Name, sj.Cron, sj.Timezone, desc, sj.TimeoutSeconds, force)
		if pullErr != nil {
			fmt.Printf("  %s %s: %v\n", ui.Error("Failed to pull"), ui.Cyan(sj.Slug), pullErr)
			continue
		}

		if created {
			pulledCount++
			fmt.Println(ui.Success("Pulled %s -> %s", ui.Bold(sj.Slug), ui.Cyan(filePath)))
		} else {
			skippedCount++
			fmt.Println(ui.Info("Local script already exists: %s %s", ui.Cyan(filePath), ui.Dim("(use --force to overwrite)")))
		}
	}

	fmt.Println()
	if pulledCount > 0 {
		fmt.Println(ui.Success("Successfully pulled %d job(s) from server (%s)", pulledCount, ui.Dim(cfg.ServerURL)))
	} else if skippedCount > 0 {
		fmt.Println(ui.Info("All %d server job(s) already exist locally in %s", skippedCount, ui.Dim(ws.JobsDir())))
	}
}

func runJobList(args []string) {
	wsHint := peekWorkspace(args)
	cfg, err := config.Load(wsHint)
	if err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	fs := flag.NewFlagSet("job list", flag.ExitOnError)
	fs.StringVar(&cfg.ServerURL, "server", cfg.ServerURL, "Beep server URL")
	fs.StringVar(&cfg.RunnerToken, "token", cfg.RunnerToken, "Runner token")
	fs.StringVar(&cfg.Workspace, "workspace", cfg.Workspace, "Workspace directory")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("%s", ui.Error("Error: %v", err))
	}

	ws, err := workspace.Open(cfg.Workspace)
	if err != nil {
		log.Fatalf("%s", ui.Error("Workspace error: %v", err))
	}

	localJobs, err := ws.ListJobs()
	if err != nil {
		log.Fatalf("%s", ui.Error("Failed to list local jobs: %v", err))
	}

	fmt.Printf("%s (%s):\n", ui.Bold(ui.Cyan("Local Workspace Jobs")), ui.Dim(ws.Root))
	if len(localJobs) == 0 {
		fmt.Println(ui.Dim("  (No local jobs found)"))
	} else {
		for _, j := range localJobs {
			desc := fmt.Sprintf("[%s, %s: %s]", ui.Bold(j.Name), ui.Dim("cron"), ui.Yellow(j.Cron))
			if j.TimeoutSeconds > 0 && j.TimeoutSeconds != 30 {
				desc += fmt.Sprintf(" (%ds)", j.TimeoutSeconds)
			}
			if j.FilePath != "" {
				fmt.Printf("  %s %-18s %-42s -> %s\n", ui.Bullet(), ui.Cyan(j.Slug), desc, ui.Dim(j.FilePath))
			} else {
				fmt.Printf("  %s %-18s %-42s -> %s (jobs.json)\n", ui.Bullet(), ui.Cyan(j.Slug), desc, ui.Dim(strings.Join(j.Command, " ")))
			}
		}
	}

	if cfg.ServerURL != "" && cfg.RunnerToken != "" {
		fmt.Println()
		fmt.Printf("%s (%s):\n", ui.Bold(ui.Cyan("Server Jobs")), ui.Dim(cfg.ServerURL))
		c := client.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		serverJobs, err := c.ListJobs(ctx)
		if err != nil {
			fmt.Printf("  %s\n", ui.Warn("Could not fetch server jobs: %v", err))
		} else if len(serverJobs) == 0 {
			fmt.Println(ui.Dim("  (No jobs configured on server)"))
		} else {
			for _, sj := range serverJobs {
				fmt.Printf("  %s %-18s [%s] %s: %-14s %s\n",
					ui.Bullet(), ui.Cyan(sj.Slug), ui.Bold(sj.Name), ui.Dim("cron"), ui.Yellow(sj.Cron), ui.StatusBadge(sj.Status))
			}
		}
	}
}


