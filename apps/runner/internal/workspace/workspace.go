package workspace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type JobSpec struct {
	Command        []string `json:"command"`
	Name           string   `json:"name,omitempty"`
	Cron           string   `json:"cron,omitempty"`
	Schedule       string   `json:"schedule,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Description    string   `json:"description,omitempty"`
}

type LocalJob struct {
	ID             string   `json:"id,omitempty"`
	Slug           string   `json:"slug"`
	Name           string   `json:"name"`
	Cron           string   `json:"cron"`
	Timezone       string   `json:"timezone,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Description    string   `json:"description,omitempty"`
	FilePath       string   `json:"file_path,omitempty"`
	Command        []string `json:"command,omitempty"`
	Source         string   `json:"source"` // "file" or "jobs.json"
}

type JobMetadata struct {
	ID             string
	Name           string
	Cron           string
	Timezone       string
	TimeoutSeconds int
	Description    string
}

type Workspace struct {
	Root string
	jobs map[string]JobSpec
}

func Open(root string) (*Workspace, error) {
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("workspace path required: %w", err)
		}
		root = filepath.Join(home, ".beep-runner")
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(abs, "jobs"), 0o755); err != nil {
		return nil, err
	}

	ws := &Workspace{Root: abs, jobs: map[string]JobSpec{}}
	if err := ws.loadJobsJSON(); err != nil {
		return nil, err
	}
	return ws, nil
}

func (w *Workspace) JobsDir() string {
	return filepath.Join(w.Root, "jobs")
}

func HumanizeSlug(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || r == ' '
	})
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	if len(parts) == 0 {
		return slug
	}
	return strings.Join(parts, " ")
}

func ParseTimeoutSeconds(val string) int {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "" {
		return 0
	}
	if d, err := time.ParseDuration(val); err == nil {
		return int(d.Seconds())
	}
	if n, err := strconv.Atoi(strings.TrimSuffix(val, "s")); err == nil && n > 0 {
		return n
	}
	return 0
}

func ParseScriptMetadata(filePath string) (JobMetadata, error) {
	var meta JobMetadata

	file, err := os.Open(filePath)
	if err != nil {
		return meta, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 60 {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#!") {
			continue
		}

		var comment string
		if strings.HasPrefix(line, "#") {
			comment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else if strings.HasPrefix(line, "//") {
			comment = strings.TrimSpace(strings.TrimPrefix(line, "//"))
		} else if strings.HasPrefix(line, "--") {
			comment = strings.TrimSpace(strings.TrimPrefix(line, "--"))
		} else {
			// Stop scanning when reaching non-comment script lines
			break
		}

		if !strings.HasPrefix(comment, "@") {
			continue
		}

		directive := strings.TrimPrefix(comment, "@")
		var key, val string
		if idx := strings.IndexAny(directive, ":="); idx != -1 {
			key = strings.TrimSpace(directive[:idx])
			val = strings.TrimSpace(directive[idx+1:])
		} else if idx := strings.Index(directive, " "); idx != -1 {
			key = strings.TrimSpace(directive[:idx])
			val = strings.TrimSpace(directive[idx+1:])
		} else {
			continue
		}

		switch strings.ToLower(key) {
		case "id", "job_id", "jobid":
			meta.ID = val
		case "name", "title":
			meta.Name = val
		case "schedule", "cron":
			meta.Cron = val
		case "timezone", "tz":
			meta.Timezone = val
		case "timeout", "timeout_seconds":
			meta.TimeoutSeconds = ParseTimeoutSeconds(val)
		case "description", "desc":
			meta.Description = val
		}
	}

	return meta, scanner.Err()
}

func DetectTimezone() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	loc := time.Now().Location().String()
	if loc != "" && loc != "Local" {
		return loc
	}
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		tz := strings.TrimSpace(string(data))
		if tz != "" {
			return tz
		}
	}
	return "UTC"
}

func (w *Workspace) CreateScript(slug, scriptType, id, name, cron, timezone, description string, timeoutSeconds int) (filePath string, created bool, err error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return "", false, fmt.Errorf("job slug cannot be empty")
	}

	if name == "" {
		name = HumanizeSlug(slug)
	}
	if cron == "" {
		cron = "*/5 * * * *"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	if timezone == "" {
		timezone = DetectTimezone()
	}
	if description == "" {
		description = fmt.Sprintf("Health check job for %s", slug)
	}

	ext := ".sh"
	switch strings.ToLower(scriptType) {
	case "py", "python":
		ext = ".py"
	case "js", "node", "javascript":
		ext = ".js"
	case "rb", "ruby":
		ext = ".rb"
	case "sh", "bash", "":
		ext = ".sh"
	default:
		ext = "." + strings.TrimPrefix(scriptType, ".")
	}

	targetPath := filepath.Join(w.JobsDir(), slug+ext)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, false, nil
	}

	var content string
	switch ext {
	case ".py":
		content = fmt.Sprintf(`#!/usr/bin/env python3
# @id: %s
# @name: %s
# @schedule: %s
# @timeout: %ds
# @timezone: %s
# @description: %s

import os
import sys

# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

print(f"[{os.getenv('BEEP_JOB_SLUG', '%s')}] Starting health check...")

# Put your check logic here.
# Exit 0 for ok, non-zero for error/alerting.
print("Check passed successfully.")
sys.exit(0)
`, id, name, cron, timeoutSeconds, timezone, description, slug)
	case ".js":
		content = fmt.Sprintf(`#!/usr/bin/env node
// @id: %s
// @name: %s
// @schedule: %s
// @timeout: %ds
// @timezone: %s
// @description: %s

// Environment variables available:
//   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
//   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

console.log("[%s] Starting health check...");

// Put your check logic here.
// Exit 0 for ok, non-zero for error/alerting.
console.log("Check passed successfully.");
process.exit(0);
`, id, name, cron, timeoutSeconds, timezone, description, slug)
	case ".rb":
		content = fmt.Sprintf(`#!/usr/bin/env ruby
# @id: %s
# @name: %s
# @schedule: %s
# @timeout: %ds
# @timezone: %s
# @description: %s

puts "[%s] Starting health check..."
# Put your check logic here.
puts "Check passed successfully."
exit 0
`, id, name, cron, timeoutSeconds, timezone, description, slug)
	default:
		content = fmt.Sprintf(`#!/usr/bin/env bash
# @id: %s
# @name: %s
# @schedule: %s
# @timeout: %ds
# @timezone: %s
# @description: %s

set -euo pipefail

# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

echo "[%s] Starting health check..."

# Put your check logic here. For example:
# curl -s -f -m 10 "http://127.0.0.1:8080/health" || exit 1

echo "Check passed successfully."
exit 0
`, id, name, cron, timeoutSeconds, timezone, description, slug)
	}

	if err := os.WriteFile(targetPath, []byte(content), 0o755); err != nil {
		return "", false, fmt.Errorf("failed to write script %s: %w", targetPath, err)
	}

	return targetPath, true, nil
}

func (w *Workspace) UpdateScriptHeader(filePath, id, name, cron, timezone, description string, timeoutSeconds int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	commentPrefix := "# "
	if ext == ".js" || ext == ".ts" {
		commentPrefix = "// "
	} else if ext == ".sql" {
		commentPrefix = "-- "
	}

	lines := strings.Split(string(data), "\n")
	var shebang string
	startIndex := 0

	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		shebang = lines[0]
		startIndex = 1
	}

	// Skip existing @ comment lines
	for startIndex < len(lines) {
		trimmed := strings.TrimSpace(lines[startIndex])
		if trimmed == "" {
			startIndex++
			continue
		}
		isComment := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "--")
		if isComment {
			c := strings.TrimLeft(trimmed, "#/- ")
			if strings.HasPrefix(c, "@") {
				startIndex++
				continue
			}
		}
		break
	}

	remainingBody := strings.Join(lines[startIndex:], "\n")
	remainingBody = strings.TrimLeft(remainingBody, "\r\n")

	if cron == "" {
		cron = "*/5 * * * *"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	if timezone == "" {
		timezone = DetectTimezone()
	}

	var sb strings.Builder
	if shebang != "" {
		sb.WriteString(shebang)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%s@id: %s\n", commentPrefix, id))
	sb.WriteString(fmt.Sprintf("%s@name: %s\n", commentPrefix, name))
	sb.WriteString(fmt.Sprintf("%s@schedule: %s\n", commentPrefix, cron))
	sb.WriteString(fmt.Sprintf("%s@timeout: %ds\n", commentPrefix, timeoutSeconds))
	sb.WriteString(fmt.Sprintf("%s@timezone: %s\n", commentPrefix, timezone))
	sb.WriteString(fmt.Sprintf("%s@description: %s\n\n", commentPrefix, description))
	sb.WriteString(remainingBody)

	return os.WriteFile(filePath, []byte(sb.String()), 0o755)
}

func (w *Workspace) UpdateScriptID(filePath string, id string) error {
	meta, err := ParseScriptMetadata(filePath)
	if err != nil {
		return err
	}
	return w.UpdateScriptHeader(filePath, id, meta.Name, meta.Cron, meta.Timezone, meta.Description, meta.TimeoutSeconds)
}

func (w *Workspace) FindScriptFile(slug string) (string, bool) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	candidates := []string{
		filepath.Join(w.JobsDir(), slug),
		filepath.Join(w.JobsDir(), slug+".sh"),
		filepath.Join(w.JobsDir(), slug+".py"),
		filepath.Join(w.JobsDir(), slug+".rb"),
		filepath.Join(w.JobsDir(), slug+".js"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, true
		}
	}
	return "", false
}

func (w *Workspace) PullJob(slug, scriptType, id, name, cron, timezone, description string, timeoutSeconds int, overwrite bool) (filePath string, created bool, err error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return "", false, fmt.Errorf("job slug cannot be empty")
	}

	existingPath, found := w.FindScriptFile(slug)
	if found {
		if err := w.UpdateScriptHeader(existingPath, id, name, cron, timezone, description, timeoutSeconds); err != nil {
			return existingPath, false, fmt.Errorf("failed to update script header for %s: %w", existingPath, err)
		}
		return existingPath, false, nil
	}

	return w.CreateScript(slug, scriptType, id, name, cron, timezone, description, timeoutSeconds)
}

func (w *Workspace) RemoveJob(slug string) (removedFiles []string, removedFromJSON bool, err error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, false, fmt.Errorf("job slug cannot be empty")
	}

	entries, readErr := os.ReadDir(w.JobsDir())
	if readErr == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			fileSlug := strings.TrimSuffix(name, ext)
			if strings.EqualFold(fileSlug, slug) {
				fullPath := filepath.Join(w.JobsDir(), name)
				if rmErr := os.Remove(fullPath); rmErr == nil {
					removedFiles = append(removedFiles, fullPath)
				}
			}
		}
	}

	if _, ok := w.jobs[slug]; ok {
		delete(w.jobs, slug)
		removedFromJSON = true
		if saveErr := w.saveJobsJSON(); saveErr != nil {
			return removedFiles, removedFromJSON, saveErr
		}
	}

	if len(removedFiles) == 0 && !removedFromJSON {
		return nil, false, fmt.Errorf("job %q not found in workspace %s", slug, w.Root)
	}

	return removedFiles, removedFromJSON, nil
}

func (w *Workspace) ListJobs() ([]LocalJob, error) {
	jobsMap := make(map[string]LocalJob)

	entries, err := os.ReadDir(w.JobsDir())
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			slug := strings.TrimSuffix(name, ext)
			if slug == "" {
				continue
			}
			filePath := filepath.Join(w.JobsDir(), name)
			meta, _ := ParseScriptMetadata(filePath)

			jobName := meta.Name
			if jobName == "" {
				jobName = HumanizeSlug(slug)
			}
			cron := meta.Cron
			if cron == "" {
				cron = "*/5 * * * *"
			}
			timeout := meta.TimeoutSeconds
			if timeout <= 0 {
				timeout = 30
			}

			jobsMap[slug] = LocalJob{
				ID:             meta.ID,
				Slug:           slug,
				Name:           jobName,
				Cron:           cron,
				Timezone:       meta.Timezone,
				TimeoutSeconds: timeout,
				Description:    meta.Description,
				FilePath:       filePath,
				Source:         "file",
			}
		}
	}

	for slug, spec := range w.jobs {
		jobName := spec.Name
		if jobName == "" {
			jobName = HumanizeSlug(slug)
		}
		cron := spec.Cron
		if cron == "" && spec.Schedule != "" {
			cron = spec.Schedule
		}
		if cron == "" {
			cron = "*/5 * * * *"
		}
		timeout := spec.TimeoutSeconds
		if timeout <= 0 {
			timeout = 30
		}

		jobsMap[slug] = LocalJob{
			Slug:           slug,
			Name:           jobName,
			Cron:           cron,
			Timezone:       spec.Timezone,
			TimeoutSeconds: timeout,
			Description:    spec.Description,
			Command:        spec.Command,
			Source:         "jobs.json",
		}
	}

	result := make([]LocalJob, 0, len(jobsMap))
	for _, job := range jobsMap {
		result = append(result, job)
	}
	return result, nil
}

func (w *Workspace) Resolve(slug string) (argv []string, err error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("job slug is empty")
	}

	if spec, ok := w.jobs[slug]; ok && len(spec.Command) > 0 {
		return w.expandArgv(spec.Command), nil
	}

	jobsDir := w.JobsDir()
	candidates := []string{
		filepath.Join(jobsDir, slug),
		filepath.Join(jobsDir, slug+".sh"),
		filepath.Join(jobsDir, slug+".py"),
		filepath.Join(jobsDir, slug+".rb"),
		filepath.Join(jobsDir, slug+".js"),
	}
	if runtime.GOOS == "windows" {
		candidates = append([]string{
			filepath.Join(jobsDir, slug+".cmd"),
			filepath.Join(jobsDir, slug+".bat"),
			filepath.Join(jobsDir, slug+".ps1"),
		}, candidates...)
	}

	for _, path := range candidates {
		info, statErr := os.Stat(path)
		if statErr != nil || info.IsDir() {
			continue
		}
		return []string{path}, nil
	}

	return nil, fmt.Errorf("no local script for job %q in %s (add jobs/%s or jobs.json)", slug, jobsDir, slug)
}

func (w *Workspace) loadJobsJSON() error {
	path := filepath.Join(w.Root, "jobs.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var parsed struct {
		Jobs map[string]JobSpec `json:"jobs"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		var flat map[string]JobSpec
		if err2 := json.Unmarshal(data, &flat); err2 != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		parsed.Jobs = flat
	}
	for slug, spec := range parsed.Jobs {
		w.jobs[slug] = spec
	}
	return nil
}

func (w *Workspace) saveJobsJSON() error {
	path := filepath.Join(w.Root, "jobs.json")
	if len(w.jobs) == 0 {
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(path)
		}
		return nil
	}
	payload := map[string]any{
		"jobs": w.jobs,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (w *Workspace) expandArgv(argv []string) []string {
	out := make([]string, len(argv))
	for i, arg := range argv {
		if strings.HasPrefix(arg, "/") || filepath.IsAbs(arg) {
			out[i] = arg
			continue
		}
		if strings.Contains(arg, string(filepath.Separator)) || strings.HasPrefix(arg, "jobs/") || strings.HasPrefix(arg, "./") {
			out[i] = filepath.Join(w.Root, arg)
			continue
		}
		out[i] = arg
	}
	return out
}
