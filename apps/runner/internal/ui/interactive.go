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

	if res.Name == "" {
		res.Name = workspace.HumanizeSlug(res.Slug)
	}

	cronChoice := "*/5 * * * *"
	customCron := ""
	if res.Cron != "" && res.Cron != "*/5 * * * *" {
		cronChoice = "custom"
		customCron = res.Cron
	}

	if res.ScriptType == "" {
		res.ScriptType = "sh"
	}
	if res.Timezone == "" {
		res.Timezone = workspace.DetectTimezone()
	}
	if res.TimeoutSeconds <= 0 {
		res.TimeoutSeconds = 30
	}
	if res.Description == "" {
		res.Description = fmt.Sprintf("Health check job for %s", res.Slug)
	}

	timeoutStr := strconv.Itoa(res.TimeoutSeconds) + "s"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Display Name").
				Description("Human-readable title for the job").
				Value(&res.Name),

			huh.NewSelect[string]().
				Title("Script Type").
				Description("Language template for the local executable").
				Options(
					huh.NewOption("Bash / Shell (.sh)", "sh"),
					huh.NewOption("Python 3 (.py)", "py"),
					huh.NewOption("Node.js (.js)", "js"),
					huh.NewOption("Ruby (.rb)", "rb"),
				).
				Value(&res.ScriptType),

			huh.NewSelect[string]().
				Title("Schedule (Cron)").
				Description("Execution frequency").
				Options(
					huh.NewOption("Every 5 minutes (*/5 * * * *) - Recommended", "*/5 * * * *"),
					huh.NewOption("Every minute (* * * * *)", "* * * * *"),
					huh.NewOption("Every 10 minutes (*/10 * * * *)", "*/10 * * * *"),
					huh.NewOption("Every hour (0 * * * *)", "0 * * * *"),
					huh.NewOption("Every day at midnight (0 0 * * *)", "0 0 * * *"),
					huh.NewOption("Custom cron expression...", "custom"),
				).
				Value(&cronChoice),

			huh.NewSelect[string]().
				Title("Timeout").
				Description("Max execution time before considering task timed out").
				Options(
					huh.NewOption("30 seconds (Default)", "30s"),
					huh.NewOption("10 seconds (Fast check)", "10s"),
					huh.NewOption("60 seconds (1 minute)", "60s"),
					huh.NewOption("120 seconds (2 minutes)", "120s"),
				).
				Value(&timeoutStr),

			huh.NewInput().
				Title("Timezone").
				Description("IANA timezone for cron evaluation").
				Value(&res.Timezone),

			huh.NewInput().
				Title("Description").
				Description("Optional summary of what this job checks").
				Value(&res.Description),

			huh.NewConfirm().
				Title("Register on server now?").
				Description("Sync and register this job to Beep Core immediately").
				Value(&res.SyncToServer),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	if cronChoice == "custom" {
		err := huh.NewInput().
			Title("Custom Cron Expression").
			Placeholder("e.g. */15 * * * *").
			Value(&customCron).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("cron expression cannot be empty")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
		res.Cron = strings.TrimSpace(customCron)
	} else {
		res.Cron = cronChoice
	}

	res.TimeoutSeconds = workspace.ParseTimeoutSeconds(timeoutStr)
	if res.TimeoutSeconds <= 0 {
		res.TimeoutSeconds = 30
	}

	return &res, nil
}

// PromptJobRemove asks the user to select which job(s) to remove.
func PromptJobRemove(localJobs []workspace.LocalJob) (slugs []string, syncServer bool, err error) {
	if len(localJobs) == 0 {
		return nil, false, errors.New("no local jobs found to remove")
	}

	options := make([]huh.Option[string], 0, len(localJobs))
	for _, j := range localJobs {
		label := fmt.Sprintf("%s (%s, %s)", j.Slug, j.Name, j.Cron)
		options = append(options, huh.NewOption(label, j.Slug))
	}

	var selected []string
	syncServer = true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Jobs to Remove").
				Description("Choose one or more local jobs to delete").
				Options(options...).
				Value(&selected).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return errors.New("please select at least one job")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Delete from server as well?").
				Description("Remove corresponding job definitions from Beep Core").
				Value(&syncServer),
		),
	)

	if err := form.Run(); err != nil {
		return nil, false, err
	}

	return selected, syncServer, nil
}

type SyncStatus string

const (
	StatusSynced     SyncStatus = "synced"
	StatusModified   SyncStatus = "modified"
	StatusLocalOnly  SyncStatus = "local only"
	StatusRemoteOnly SyncStatus = "remote only"
)

type JobCompareItem struct {
	Slug      string
	Status    SyncStatus
	LocalJob  *workspace.LocalJob
	ServerJob *client.ServerJob
	Diffs     []string
}

func CompareJob(slug string, lj *workspace.LocalJob, sj *client.ServerJob) JobCompareItem {
	item := JobCompareItem{
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

func formatPushOption(item JobCompareItem) huh.Option[string] {
	name := ""
	cron := ""
	if item.LocalJob != nil {
		name = item.LocalJob.Name
		cron = item.LocalJob.Cron
	}

	var label string
	switch item.Status {
	case StatusModified:
		label = fmt.Sprintf("%s [modified / out of sync] (%s, %s)", item.Slug, name, cron)
	case StatusLocalOnly:
		label = fmt.Sprintf("%s [local only] (%s, %s)", item.Slug, name, cron)
	case StatusSynced:
		label = fmt.Sprintf("%s [synced] (%s, %s)", item.Slug, name, cron)
	default:
		label = fmt.Sprintf("%s (%s, %s)", item.Slug, name, cron)
	}
	return huh.NewOption(label, item.Slug)
}

func formatPullOption(item JobCompareItem) huh.Option[string] {
	name := ""
	cron := ""
	if item.ServerJob != nil {
		name = item.ServerJob.Name
		cron = item.ServerJob.Cron
	}

	var label string
	switch item.Status {
	case StatusModified:
		label = fmt.Sprintf("%s [modified / out of sync] (%s, %s)", item.Slug, name, cron)
	case StatusRemoteOnly:
		label = fmt.Sprintf("%s [remote only] (%s, %s)", item.Slug, name, cron)
	case StatusSynced:
		label = fmt.Sprintf("%s [synced] (%s, %s)", item.Slug, name, cron)
	default:
		label = fmt.Sprintf("%s (%s, %s)", item.Slug, name, cron)
	}
	return huh.NewOption(label, item.Slug)
}

// PromptJobPushSelection asks the user which local job(s) to push.
func PromptJobPushSelection(items []JobCompareItem) (selectedSlugs []string, err error) {
	if len(items) == 0 {
		return nil, errors.New("no local jobs found to push")
	}

	options := make([]huh.Option[string], 0, len(items))
	var preSelected []string
	for _, it := range items {
		options = append(options, formatPushOption(it))
		if it.Status == StatusModified || it.Status == StatusLocalOnly {
			preSelected = append(preSelected, it.Slug)
		}
	}
	if len(preSelected) == 0 {
		for _, it := range items {
			preSelected = append(preSelected, it.Slug)
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
func PromptJobPullSelection(items []JobCompareItem) (selectedSlugs []string, force bool, err error) {
	if len(items) == 0 {
		return nil, false, errors.New("no server jobs available to pull")
	}

	options := make([]huh.Option[string], 0, len(items))
	var preSelected []string
	for _, it := range items {
		options = append(options, formatPullOption(it))
		if it.Status == StatusModified || it.Status == StatusRemoteOnly {
			preSelected = append(preSelected, it.Slug)
		}
	}
	if len(preSelected) == 0 {
		for _, it := range items {
			preSelected = append(preSelected, it.Slug)
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

			huh.NewConfirm().
				Title("Overwrite existing local scripts?").
				Description("If enabled, replaces existing script header and content").
				Value(&force),
		),
	)

	if err := form.Run(); err != nil {
		return nil, false, err
	}

	return selectedSlugs, force, nil
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
				Placeholder("beep_rt_...").
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
