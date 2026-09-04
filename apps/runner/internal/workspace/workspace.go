package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type JobSpec struct {
	Command []string `json:"command"`
}

type LocalJob struct {
	Slug     string   `json:"slug"`
	FilePath string   `json:"file_path,omitempty"`
	Command  []string `json:"command,omitempty"`
	Source   string   `json:"source"` // "file" or "jobs.json"
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

func (w *Workspace) CreateScript(slug, scriptType string) (filePath string, created bool, err error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return "", false, fmt.Errorf("job slug cannot be empty")
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
import os
import sys

# Beep Runner job: %s
# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

print(f"[{os.getenv('BEEP_JOB_SLUG', '%s')}] Starting health check...")

# Put your check logic here.
# Exit 0 for ok, non-zero for error/alerting.
print("Check passed successfully.")
sys.exit(0)
`, slug, slug)
	case ".js":
		content = fmt.Sprintf(`#!/usr/bin/env node
// Beep Runner job: %s
// Environment variables available:
//   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
//   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

console.log("[%s] Starting health check...");

// Put your check logic here.
// Exit 0 for ok, non-zero for error/alerting.
console.log("Check passed successfully.");
process.exit(0);
`, slug, slug)
	case ".rb":
		content = fmt.Sprintf(`#!/usr/bin/env ruby
# Beep Runner job: %s

puts "[%s] Starting health check..."
# Put your check logic here.
puts "Check passed successfully."
exit 0
`, slug, slug)
	default:
		content = fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

# Beep Runner job: %s
# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

echo "[%s] Starting health check..."

# Put your check logic here. For example:
# curl -s -f -m 10 "http://127.0.0.1:8080/health" || exit 1

echo "Check passed successfully."
exit 0
`, slug, slug)
	}

	if err := os.WriteFile(targetPath, []byte(content), 0o755); err != nil {
		return "", false, fmt.Errorf("failed to write script %s: %w", targetPath, err)
	}

	return targetPath, true, nil
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
			jobsMap[slug] = LocalJob{
				Slug:     slug,
				FilePath: filePath,
				Source:   "file",
			}
		}
	}

	for slug, spec := range w.jobs {
		jobsMap[slug] = LocalJob{
			Slug:    slug,
			Command: spec.Command,
			Source:  "jobs.json",
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
