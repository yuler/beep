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

// PromptJobPushSelection asks the user which jobs to push.
func PromptJobPushSelection(localJobs []workspace.LocalJob) (selectedSlugs []string, err error) {
	if len(localJobs) == 0 {
		return nil, errors.New("no local jobs found")
	}

	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("Push all jobs (%d total)", len(localJobs)), "__all__"),
	}
	for _, j := range localJobs {
		label := fmt.Sprintf("%s (%s)", j.Slug, j.Name)
		options = append(options, huh.NewOption(label, j.Slug))
	}

	var choice string
	err = huh.NewSelect[string]().
		Title("Push Jobs to Server").
		Description("Select which job(s) to push to Beep Core").
		Options(options...).
		Value(&choice).
		Run()
	if err != nil {
		return nil, err
	}

	if choice == "__all__" {
		for _, j := range localJobs {
			selectedSlugs = append(selectedSlugs, j.Slug)
		}
		return selectedSlugs, nil
	}

	return []string{choice}, nil
}

// PromptJobPullSelection asks the user which server jobs to pull.
func PromptJobPullSelection(serverJobs []*client.ServerJob) (selectedSlugs []string, force bool, err error) {
	if len(serverJobs) == 0 {
		return nil, false, errors.New("no server jobs available to pull")
	}

	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("Pull all jobs (%d total)", len(serverJobs)), "__all__"),
	}
	for _, j := range serverJobs {
		label := fmt.Sprintf("%s (%s, %s)", j.Slug, j.Name, j.Cron)
		options = append(options, huh.NewOption(label, j.Slug))
	}

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pull Jobs from Server").
				Description("Select which server job(s) to pull to your local workspace").
				Options(options...).
				Value(&choice),

			huh.NewConfirm().
				Title("Overwrite existing local scripts?").
				Description("If enabled, replaces existing script header and content").
				Value(&force),
		),
	)

	if err := form.Run(); err != nil {
		return nil, false, err
	}

	if choice == "__all__" {
		for _, j := range serverJobs {
			selectedSlugs = append(selectedSlugs, j.Slug)
		}
		return selectedSlugs, force, nil
	}

	return []string{choice}, force, nil
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
