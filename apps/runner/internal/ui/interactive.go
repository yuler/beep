package ui

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"beep-runner/internal/client"
	"beep-runner/internal/config"
	"beep-runner/internal/schedule"
	"beep-runner/internal/workspace"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// IsInteractive checks if stdin and stdout are attached to a terminal.
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) &&
		isatty.IsTerminal(os.Stdout.Fd()) &&
		os.Getenv("CI") == ""
}

type JobCreateParams struct {
	Slug           string
	Name           string
	ScriptType     string
	Cron           string
	Timezone       string
	Description    string
	TimeoutSeconds int
	SyncToServer   bool
}

// PromptJobCreate runs an interactive wizard for creating a job.
func PromptJobCreate(defaults JobCreateParams) (*JobCreateParams, error) {
	res := defaults

	// 1. Slug
	if res.Slug == "" {
		err := huh.NewInput().
			Title("Job Slug").
			Description("Unique identifier for the job (lowercase alphanumeric, dashes, underscores)").
			Placeholder("e.g. intranet-gateway, db-backup-check").
			Value(&res.Slug).
			Validate(func(s string) error {
				s = strings.TrimSpace(strings.ToLower(s))
				if s == "" {
					return errors.New("slug is required")
				}
				if !slugRegex.MatchString(s) {
					return errors.New("slug must start with alphanumeric and contain only letters, numbers, '-' or '_'")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Slug = strings.TrimSpace(strings.ToLower(res.Slug))

	// 2. Name
	if res.Name == "" {
		res.Name = workspace.HumanizeSlug(res.Slug)
	}
	err := huh.NewInput().
		Title("Display Name").
		Description("Human-readable title for the job").
		Placeholder("e.g. " + workspace.HumanizeSlug(res.Slug)).
		Value(&res.Name).
		Run()
	if err != nil {
		return nil, err
	}
	res.Name = strings.TrimSpace(res.Name)
	if res.Name == "" {
		res.Name = workspace.HumanizeSlug(res.Slug)
	}

	// 3. Description
	err = huh.NewInput().
		Title("Description").
		Description("Optional summary of what this job checks").
		Placeholder("e.g. Ping internal health check endpoint").
		Value(&res.Description).
		Run()
	if err != nil {
		return nil, err
	}
	res.Description = strings.TrimSpace(res.Description)

	// 4. Schedule (Cron)
	cronChoice := "*/5 * * * *"
	customCron := ""
	if res.Cron != "" && res.Cron != "*/5 * * * *" {
		cronChoice = "custom"
		customCron = res.Cron
	}

	err = huh.NewSelect[string]().
		Title("Schedule (Cron)").
		Description("Classic cron or Fugit semantic phrases (same as Beep Core)").
		Options(
			huh.NewOption("Every 5 minutes (*/5 * * * *) - Recommended", "*/5 * * * *"),
			huh.NewOption("Every minute (* * * * *)", "* * * * *"),
			huh.NewOption("Every 10 minutes (*/10 * * * *)", "*/10 * * * *"),
			huh.NewOption("Every hour (0 * * * *)", "0 * * * *"),
			huh.NewOption("Every day at midnight (0 0 * * *)", "0 0 * * *"),
			huh.NewOption("Custom (cron or \"every 5 minutes\")...", "custom"),
		).
		Value(&cronChoice).
		Run()
	if err != nil {
		return nil, err
	}

	if cronChoice == "custom" {
		err := huh.NewInput().
			Title("Custom Schedule").
			Description("Fugit-compatible: */15 * * * * or every 15 minutes / every day at noon").
			Placeholder("e.g. every 15 minutes").
			Value(&customCron).
			Validate(func(s string) error {
				return schedule.Validate(s)
			}).
			Run()
		if err != nil {
			return nil, err
		}
		res.Cron = strings.TrimSpace(customCron)
	} else {
		res.Cron = cronChoice
	}

	// 5. Script Runtime / Shebang
	runtimeChoice := res.ScriptType
	if runtimeChoice == "" {
		runtimeChoice = "bash"
	}
	customShebang := ""
	if strings.HasPrefix(runtimeChoice, "#!") || (runtimeChoice != "bash" && runtimeChoice != "node" && runtimeChoice != "bun" && runtimeChoice != "python" && runtimeChoice != "ruby") {
		customShebang = runtimeChoice
		runtimeChoice = "custom"
	}

	err = huh.NewSelect[string]().
		Title("Script Runtime & Shebang").
		Description("Select execution interpreter or specify custom shebang").
		Options(
			huh.NewOption("Bash (#!/usr/bin/env bash)", "bash"),
			huh.NewOption("Node.js (#!/usr/bin/env node)", "node"),
			huh.NewOption("Bun (#!/usr/bin/env bun)", "bun"),
			huh.NewOption("Python 3 (#!/usr/bin/env python3)", "python"),
			huh.NewOption("Ruby (#!/usr/bin/env ruby)", "ruby"),
			huh.NewOption("Custom Shebang / Interpreter...", "custom"),
		).
		Value(&runtimeChoice).
		Run()
	if err != nil {
		return nil, err
	}

	if runtimeChoice == "custom" {
		if customShebang == "" {
			customShebang = "#!/usr/bin/env "
		}
		err = huh.NewInput().
			Title("Custom Shebang / Interpreter").
			Description("Enter full shebang (e.g. #!/usr/bin/env deno) or binary name (e.g. deno, php)").
			Placeholder("#!/usr/bin/env deno").
			Value(&customShebang).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("shebang cannot be empty")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
		res.ScriptType = strings.TrimSpace(customShebang)
	} else {
		res.ScriptType = runtimeChoice
	}

	// 6. Timeout
	if res.TimeoutSeconds <= 0 {
		res.TimeoutSeconds = 30
	}
	timeoutStr := strconv.Itoa(res.TimeoutSeconds) + "s"
	err = huh.NewSelect[string]().
		Title("Timeout").
		Description("Max execution time before considering task timed out").
		Options(
			huh.NewOption("30 seconds (Default)", "30s"),
			huh.NewOption("10 seconds (Fast check)", "10s"),
			huh.NewOption("60 seconds (1 minute)", "60s"),
			huh.NewOption("120 seconds (2 minutes)", "120s"),
		).
		Value(&timeoutStr).
		Run()
	if err != nil {
		return nil, err
	}
	res.TimeoutSeconds = workspace.ParseTimeoutSeconds(timeoutStr)
	if res.TimeoutSeconds <= 0 {
		res.TimeoutSeconds = 30
	}

	// 7. Timezone
	detectedTZ, tzOK := workspace.DetectTimezoneOK()
	if res.Timezone == "" {
		if tzOK {
			res.Timezone = detectedTZ
		}
	}

	if res.Timezone == "" {
		zones := workspace.ListIANATimezones()
		options := make([]huh.Option[string], 0, len(zones))
		for _, z := range zones {
			options = append(options, huh.NewOption(z, z))
		}
		err = huh.NewSelect[string]().
			Title("Timezone").
			Description("Could not detect host timezone — pick an IANA zone (type to filter)").
			Options(options...).
			Value(&res.Timezone).
			Run()
		if err != nil {
			return nil, err
		}
	} else {
		tzDesc := "IANA timezone for cron evaluation"
		if tzOK && res.Timezone == detectedTZ {
			tzDesc = fmt.Sprintf("Detected from this machine (%s) — edit if needed", detectedTZ)
		}
		err = huh.NewInput().
			Title("Timezone").
			Description(tzDesc).
			Value(&res.Timezone).
			Validate(func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return errors.New("timezone is required")
				}
				if !workspace.ValidIANATimezone(s) {
					return fmt.Errorf("%q is not a valid IANA timezone", s)
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
	}
	res.Timezone = strings.TrimSpace(res.Timezone)

	// 8. SyncToServer
	err = huh.NewConfirm().
		Title("Register on server now?").
		Description("Sync and register this job to Beep Core immediately").
		Value(&res.SyncToServer).
		Run()
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// PromptJobRemove asks the user to select which job(s) to remove.
func PromptJobRemove(localJobs []workspace.LocalJob) (slugs []string, err error) {
	if len(localJobs) == 0 {
		return nil, errors.New("no local jobs found to remove")
	}

	options := make([]huh.Option[string], 0, len(localJobs))
	for _, j := range localJobs {
		label := j.Slug
		if j.ID != "" {
			label = fmt.Sprintf("%s (id: %s)", j.Slug, j.ID)
		}
		options = append(options, huh.NewOption(label, j.Slug))
	}

	var selected []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Jobs to Remove").
				Description("Choose one or more local jobs to delete (also deletes from server unless --no-sync)").
				Options(options...).
				Value(&selected).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("please select at least one job")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return selected, nil
}

type SyncStatus string

const (
	StatusSynced     SyncStatus = "synced"
	StatusModified   SyncStatus = "modified"
	StatusLocalOnly  SyncStatus = "local only"
	StatusRemoteOnly SyncStatus = "remote only"
)

type JobCompareItem struct {
	ID        string
	Slug      string
	Status    SyncStatus
	LocalJob  *workspace.LocalJob
	ServerJob *client.ServerJob
	Diffs     []string
}

func CompareJob(slug string, lj *workspace.LocalJob, sj *client.ServerJob) JobCompareItem {
	id := ""
	if lj != nil && lj.ID != "" {
		id = lj.ID
	} else if sj != nil && sj.ID != "" {
		id = sj.ID
	}

	item := JobCompareItem{
		ID:        id,
		Slug:      slug,
		LocalJob:  lj,
		ServerJob: sj,
	}

	if lj != nil && sj == nil {
		item.Status = StatusLocalOnly
		return item
	}
	if lj == nil && sj != nil {
		item.Status = StatusRemoteOnly
		return item
	}
	if lj == nil && sj == nil {
		return item
	}

	serverDesc := ""
	if sj.Config != nil {
		if d, ok := sj.Config["description"].(string); ok {
			serverDesc = d
		}
	}

	var diffs []string
	if lj.Slug != "" && sj.Slug != "" && !strings.EqualFold(lj.Slug, sj.Slug) {
		diffs = append(diffs, fmt.Sprintf("slug: local %q != server %q", lj.Slug, sj.Slug))
	}
	if lj.Name != "" && sj.Name != "" && lj.Name != sj.Name {
		diffs = append(diffs, fmt.Sprintf("name: local %q != server %q", lj.Name, sj.Name))
	}
	if lj.Cron != "" && sj.Cron != "" && lj.Cron != sj.Cron {
		diffs = append(diffs, fmt.Sprintf("schedule: local %q != server %q", lj.Cron, sj.Cron))
	}
	ltz := lj.Timezone
	if ltz == "" {
		ltz = "UTC"
	}
	stz := sj.Timezone
	if stz == "" {
		stz = "UTC"
	}
	if !strings.EqualFold(ltz, stz) {
		diffs = append(diffs, fmt.Sprintf("timezone: local %q != server %q", ltz, stz))
	}
	lt := lj.TimeoutSeconds
	if lt <= 0 {
		lt = 30
	}
	st := sj.TimeoutSeconds
	if st <= 0 {
		st = 30
	}
	if lt != st {
		diffs = append(diffs, fmt.Sprintf("timeout: local %ds != server %ds", lt, st))
	}
	if lj.Description != "" && serverDesc != "" && lj.Description != serverDesc {
		diffs = append(diffs, fmt.Sprintf("description: local %q != server %q", lj.Description, serverDesc))
	}

	item.Diffs = diffs
	if len(diffs) > 0 {
		item.Status = StatusModified
	} else {
		item.Status = StatusSynced
	}
	return item
}

// PairJobs matches local and server jobs strictly by @id first (unique dimension),
// and only falls back to slug for local jobs without an assigned ID.
func PairJobs(localJobs []workspace.LocalJob, serverJobs []*client.ServerJob) []JobCompareItem {
	type localRef struct {
		job  workspace.LocalJob
		used bool
	}
	type serverRef struct {
		job  *client.ServerJob
		used bool
	}

	locals := make([]localRef, 0, len(localJobs))
	for _, lj := range localJobs {
		locals = append(locals, localRef{job: lj})
	}
	servers := make([]serverRef, 0, len(serverJobs))
	for _, sj := range serverJobs {
		if sj == nil {
			continue
		}
		servers = append(servers, serverRef{job: sj})
	}

	var items []JobCompareItem

	// Pass 1: Primary match strictly by @id (unique dimension).
	for i := range locals {
		if locals[i].job.ID == "" {
			continue
		}
		for j := range servers {
			if servers[j].used || servers[j].job.ID == "" {
				continue
			}
			if locals[i].job.ID == servers[j].job.ID {
				lj := locals[i].job
				sj := servers[j].job
				items = append(items, CompareJob(sj.Slug, &lj, sj))
				locals[i].used = true
				servers[j].used = true
				break
			}
		}
	}

	// Pass 2: Fallback match by slug only for remaining local jobs without an ID.
	for i := range locals {
		if locals[i].used {
			continue
		}
		slug := strings.ToLower(locals[i].job.Slug)
		for j := range servers {
			if servers[j].used {
				continue
			}
			if strings.EqualFold(servers[j].job.Slug, slug) {
				lj := locals[i].job
				sj := servers[j].job
				items = append(items, CompareJob(slug, &lj, sj))
				locals[i].used = true
				servers[j].used = true
				break
			}
		}
	}

	// Pass 3: Remaining unmatched local jobs (local only).
	for i := range locals {
		if locals[i].used {
			continue
		}
		lj := locals[i].job
		items = append(items, CompareJob(lj.Slug, &lj, nil))
	}

	// Pass 4: Remaining unmatched server jobs (remote only).
	for j := range servers {
		if servers[j].used {
			continue
		}
		sj := servers[j].job
		items = append(items, CompareJob(sj.Slug, nil, sj))
	}

	return items
}

func formatPushOption(item JobCompareItem) huh.Option[string] {
	key := item.ID
	if key == "" {
		if item.LocalJob != nil && item.LocalJob.Slug != "" {
			key = item.LocalJob.Slug
		} else {
			key = item.Slug
		}
	}
	label := formatSyncLabel(item)
	return huh.NewOption(label, key)
}

func formatPullOption(item JobCompareItem) huh.Option[string] {
	key := item.ID
	if key == "" {
		if item.ServerJob != nil && item.ServerJob.ID != "" {
			key = item.ServerJob.ID
		} else {
			key = item.Slug
		}
	}
	label := formatSyncLabel(item)
	return huh.NewOption(label, key)
}

func formatSyncLabel(item JobCompareItem) string {
	id := item.ID
	if id == "" {
		if item.LocalJob != nil && item.LocalJob.ID != "" {
			id = item.LocalJob.ID
		} else if item.ServerJob != nil && item.ServerJob.ID != "" {
			id = item.ServerJob.ID
		}
	}

	primary := item.Slug
	if id != "" {
		primary = fmt.Sprintf("%s (%s)", item.Slug, id)
	}

	state := string(item.Status)
	if state == "" {
		state = "unknown"
	}
	return fmt.Sprintf("%-36s  %s", primary, Gray(state))
}

// PromptJobPushSelection asks the user which local job(s) to push.
func PromptJobPushSelection(items []JobCompareItem) (selectedSlugs []string, err error) {
	if len(items) == 0 {
		return nil, errors.New("no local jobs found to push")
	}

	options := make([]huh.Option[string], 0, len(items))
	var preSelected []string
	for _, it := range items {
		opt := formatPushOption(it)
		options = append(options, opt)
		if it.Status == StatusModified || it.Status == StatusLocalOnly {
			preSelected = append(preSelected, opt.Value)
		}
	}

	selectedSlugs = preSelected

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Push Jobs to Server").
				Description("Select local job(s) to push to Beep Core (space to toggle, enter to confirm)").
				Options(options...).
				Value(&selectedSlugs).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("please select at least one job to push")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return selectedSlugs, nil
}

// PromptJobPullSelection asks the user which server job(s) to pull.
func PromptJobPullSelection(items []JobCompareItem) (selectedSlugs []string, err error) {
	if len(items) == 0 {
		return nil, errors.New("no server jobs available to pull")
	}

	options := make([]huh.Option[string], 0, len(items))
	var preSelected []string
	for _, it := range items {
		opt := formatPullOption(it)
		options = append(options, opt)
		if it.Status == StatusModified || it.Status == StatusRemoteOnly {
			preSelected = append(preSelected, opt.Value)
		}
	}

	selectedSlugs = preSelected

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Pull Jobs from Server").
				Description("Select server job(s) to pull into your workspace (space to toggle, enter to confirm)").
				Options(options...).
				Value(&selectedSlugs).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("please select at least one job to pull")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return selectedSlugs, nil
}

// PromptConfigSetWizard asks the user for missing runner configuration.
func PromptConfigSetWizard(fc *config.FileConfig) error {
	var (
		server       = fc.ServerURL
		token        = fc.RunnerToken
		workspaceDir = fc.Workspace
		concurrency  = strconv.Itoa(fc.Concurrency)
		interval     = fc.PollInterval
	)

	if concurrency == "0" {
		concurrency = "5"
	}
	if interval == "" {
		interval = "3s"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Beep Server URL").
				Description("API URL of your Beep instance").
				Placeholder("https://core.example.com or http://core.beep.localhost:3000").
				Value(&server).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("server URL is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Runner Token").
				Description("Authentication token from Beep Runners panel").
				Placeholder("beep_runner_...").
				EchoMode(huh.EchoModePassword).
				Value(&token).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("runner token is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Local Workspace Directory").
				Description("Where job scripts and configs are stored (default ~/.beep-runner)").
				Placeholder("~/.beep-runner").
				Value(&workspaceDir),

			huh.NewInput().
				Title("Max Concurrency").
				Description("Maximum parallel jobs executed concurrently").
				Value(&concurrency).
				Validate(func(s string) error {
					n, err := strconv.Atoi(strings.TrimSpace(s))
					if err != nil || n <= 0 {
						return errors.New("must be a positive number")
					}
					return nil
				}),

			huh.NewInput().
				Title("Poll Interval").
				Description("Interval to poll Core for due tasks").
				Value(&interval).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("poll interval is required (e.g. 3s)")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	fc.ServerURL = strings.TrimRight(strings.TrimSpace(server), "/")
	fc.RunnerToken = strings.TrimSpace(token)
	fc.Workspace = strings.TrimSpace(workspaceDir)
	if n, err := strconv.Atoi(strings.TrimSpace(concurrency)); err == nil && n > 0 {
		fc.Concurrency = n
	}
	fc.PollInterval = strings.TrimSpace(interval)

	return nil
}

// PromptConfigUnsetSelect asks which config key to unset.
func PromptConfigUnsetSelect(fc *config.FileConfig) (string, error) {
	options := []huh.Option[string]{}
	if fc.ServerURL != "" {
		options = append(options, huh.NewOption(fmt.Sprintf("Server URL (%s)", fc.ServerURL), "server"))
	}
	if fc.RunnerToken != "" {
		options = append(options, huh.NewOption("Runner Token (***)", "token"))
	}
	if fc.Workspace != "" {
		options = append(options, huh.NewOption(fmt.Sprintf("Workspace (%s)", fc.Workspace), "workspace"))
	}
	if fc.Concurrency > 0 {
		options = append(options, huh.NewOption(fmt.Sprintf("Concurrency (%d)", fc.Concurrency), "concurrency"))
	}
	if fc.PollInterval != "" {
		options = append(options, huh.NewOption(fmt.Sprintf("Poll Interval (%s)", fc.PollInterval), "poll-interval"))
	}

	if len(options) == 0 {
		return "", errors.New("no configured parameters to unset")
	}

	var choice string
	err := huh.NewSelect[string]().
		Title("Select Config Key to Unset").
		Options(options...).
		Value(&choice).
		Run()
	if err != nil {
		return "", err
	}
	return choice, nil
}
