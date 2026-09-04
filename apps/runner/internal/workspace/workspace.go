package workspace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Slug           string
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

// ValidateSlug verifies that slug is a valid single file name and not a path traversal attempt.
func ValidateSlug(slug string) error {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return fmt.Errorf("job slug cannot be empty")
	}
	if filepath.Base(slug) != slug || strings.Contains(slug, "..") || strings.ContainsAny(slug, `/\`) {
		return fmt.Errorf("invalid slug %q: must not contain path separators or parent directory references", slug)
	}
	return nil
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
	base := filepath.Base(filePath)
	meta.Slug = strings.ToLower(base)

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

// DetectTimezone returns the host IANA timezone, or "UTC" if it cannot be determined.
func DetectTimezone() string {
	if tz, ok := DetectTimezoneOK(); ok {
		return tz
	}
	return "UTC"
}

// DetectTimezoneOK returns the host IANA timezone when it can be resolved confidently.
func DetectTimezoneOK() (string, bool) {
	if tz := strings.TrimSpace(os.Getenv("TZ")); tz != "" && tz != "Local" {
		if ValidIANATimezone(tz) {
			return tz, true
		}
	}
	loc := time.Now().Location().String()
	if loc != "" && loc != "Local" && ValidIANATimezone(loc) {
		return loc, true
	}
	if tz, ok := timezoneFromLocaltime(); ok {
		return tz, true
	}
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		tz := strings.TrimSpace(string(data))
		if ValidIANATimezone(tz) {
			return tz, true
		}
	}
	return "", false
}

// ValidIANATimezone reports whether name is loadable via the system tz database.
func ValidIANATimezone(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "Local" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}

// timezoneFromLocaltime resolves /etc/localtime when it is a symlink into zoneinfo.
func timezoneFromLocaltime() (string, bool) {
	const localtime = "/etc/localtime"
	target, err := filepath.EvalSymlinks(localtime)
	if err != nil {
		return "", false
	}
	target = filepath.ToSlash(target)
	for _, prefix := range []string{
		"/usr/share/zoneinfo/",
		"/var/db/timezone/zoneinfo/",
		"/etc/zoneinfo/",
	} {
		if strings.HasPrefix(target, prefix) {
			tz := strings.TrimPrefix(target, prefix)
			tz = strings.TrimPrefix(tz, "posix/")
			tz = strings.TrimPrefix(tz, "right/")
			if ValidIANATimezone(tz) {
				return tz, true
			}
		}
	}
	return "", false
}

// ListIANATimezones returns IANA zone names from the system zoneinfo database.
func ListIANATimezones() []string {
	roots := []string{
		"/usr/share/zoneinfo",
		"/var/db/timezone/zoneinfo",
		"/etc/zoneinfo",
	}
	seen := map[string]struct{}{"UTC": {}}
	var zones []string
	zones = append(zones, "UTC")

	skipDir := map[string]struct{}{
		"posix": {}, "right": {},
	}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				base := filepath.Base(path)
				if _, skip := skipDir[base]; skip {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.Contains(filepath.Base(rel), ".") {
				return nil
			}
			if !strings.Contains(rel, "/") && (rel == "Factory" || rel == "localtime" || rel == "leapseconds" || rel == "tzdata.zi" || rel == "zone.tab" || rel == "zone1970.tab" || rel == "zonenow.tab" || rel == "iso3166.tab") {
				return nil
			}
			if !ValidIANATimezone(rel) {
				return nil
			}
			if _, ok := seen[rel]; ok {
				return nil
			}
			seen[rel] = struct{}{}
			zones = append(zones, rel)
			return nil
		})
	}

	if len(zones) == 1 {
		for _, z := range commonTimezones {
			if _, ok := seen[z]; !ok && ValidIANATimezone(z) {
				zones = append(zones, z)
			}
		}
	}

	sort.Strings(zones)
	// Keep UTC first for quicker selection.
	for i, z := range zones {
		if z == "UTC" {
			zones = append(append([]string{"UTC"}, zones[:i]...), zones[i+1:]...)
			break
		}
	}

	return zones
}

var commonTimezones = []string{
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Sao_Paulo",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Moscow",
	"Asia/Shanghai",
	"Asia/Chongqing",
	"Asia/Tokyo",
	"Asia/Singapore",
	"Asia/Kolkata",
	"Asia/Dubai",
	"Australia/Sydney",
	"Pacific/Auckland",
}

func GenerateShebang(scriptType string) (shebang string, commentPrefix string, templateBody string) {
	st := strings.TrimSpace(scriptType)
	stLower := strings.ToLower(st)

	switch stLower {
	case "bash", "sh":
		return "#!/usr/bin/env bash", "# ", `set -euo pipefail

# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

echo "[%s] Starting health check..."

# Put your check logic here. For example:
# curl -s -f -m 10 "http://127.0.0.1:8080/health" || exit 1

echo "Check passed successfully."
exit 0
`
	case "node", "nodejs", "node.js", "js":
		return "#!/usr/bin/env node", "// ", `// Environment variables available:
//   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
//   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

console.log("[%s] Starting health check...");

// Put your check logic here.
// Exit 0 for ok, non-zero for error/alerting.
console.log("Check passed successfully.");
process.exit(0);
`
	case "bun":
		return "#!/usr/bin/env bun", "// ", `// Environment variables available:
//   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
//   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

console.log("[%s] Starting health check with bun...");

// Put your check logic here.
console.log("Check passed successfully.");
process.exit(0);
`
	case "python", "python3", "py":
		return "#!/usr/bin/env python3", "# ", `import os
import sys

# Environment variables available:
#   BEEP_SERVER, BEEP_RUNNER_TOKEN, BEEP_RUN_ID, BEEP_JOB_SLUG
#   BEEP_LOG_URL, BEEP_RESULT_URL, BEEP_CONFIG, BEEP_CONFIG_*

print(f"[{os.getenv('BEEP_JOB_SLUG', '%s')}] Starting health check...")

# Put your check logic here.
# Exit 0 for ok, non-zero for error/alerting.
print("Check passed successfully.")
sys.exit(0)
`
	case "ruby", "rb":
		return "#!/usr/bin/env ruby", "# ", `puts "[%s] Starting health check..."
# Put your check logic here.
puts "Check passed successfully."
exit 0
`
	default:
		// Custom shebang or executable definition
		if strings.HasPrefix(st, "#!") {
			return st, "# ", `echo "[%s] Starting check..."
exit 0
`
		}
		if st != "" {
			return "#!/usr/bin/env " + st, "# ", `echo "[%s] Starting check..."
exit 0
`
		}
		return "#!/usr/bin/env bash", "# ", `set -euo pipefail

echo "[%s] Starting health check..."
echo "Check passed successfully."
exit 0
`
	}
}

func (w *Workspace) CreateScript(slug, scriptType, id, name, cron, timezone, description string, timeoutSeconds int) (filePath string, created bool, err error) {
	if err := ValidateSlug(slug); err != nil {
		return "", false, err
	}
	slug = strings.TrimSpace(strings.ToLower(slug))

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
	description = strings.TrimSpace(description)

	targetPath := filepath.Join(w.JobsDir(), slug)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, false, nil
	}

	shebang, commentPrefix, templateBody := GenerateShebang(scriptType)

	var sb strings.Builder
	sb.WriteString(shebang)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s@id: %s\n", commentPrefix, id))
	sb.WriteString(fmt.Sprintf("%s@name: %s\n", commentPrefix, name))
	sb.WriteString(fmt.Sprintf("%s@schedule: %s\n", commentPrefix, cron))
	sb.WriteString(fmt.Sprintf("%s@timeout: %ds\n", commentPrefix, timeoutSeconds))
	sb.WriteString(fmt.Sprintf("%s@timezone: %s\n", commentPrefix, timezone))
	sb.WriteString(fmt.Sprintf("%s@description: %s\n\n", commentPrefix, description))
	sb.WriteString(fmt.Sprintf(templateBody, slug))

	if err := os.WriteFile(targetPath, []byte(sb.String()), 0o755); err != nil {
		return "", false, fmt.Errorf("failed to write script %s: %w", targetPath, err)
	}

	return targetPath, true, nil
}

func (w *Workspace) UpdateScriptHeader(filePath, id, name, cron, timezone, description string, timeoutSeconds int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var shebang string
	startIndex := 0

	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		shebang = lines[0]
		startIndex = 1
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	commentPrefix := "# "
	if ext == ".js" || ext == ".ts" || strings.Contains(shebang, "node") || strings.Contains(shebang, "bun") || strings.Contains(shebang, "deno") {
		commentPrefix = "// "
	} else if ext == ".sql" {
		commentPrefix = "-- "
	}

	// Filter out all existing @ directive comment lines from the remaining body
	var remainingLines []string
	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "--")
		if isComment {
			c := strings.TrimLeft(trimmed, "#/- ")
			if strings.HasPrefix(c, "@") {
				continue
			}
		}
		remainingLines = append(remainingLines, line)
	}

	remainingBody := strings.Join(remainingLines, "\n")
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
	if err := ValidateSlug(slug); err != nil {
		return "", false
	}
	slug = strings.TrimSpace(strings.ToLower(slug))
	target := filepath.Join(w.JobsDir(), slug)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		return target, true
	}
	return "", false
}

func (w *Workspace) FindScriptFileByID(id string) (string, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", false
	}
	entries, err := os.ReadDir(w.JobsDir())
	if err != nil {
		return "", false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(w.JobsDir(), entry.Name())
		meta, err := ParseScriptMetadata(filePath)
		if err == nil && strings.TrimSpace(meta.ID) == id {
			return filePath, true
		}
	}
	return "", false
}

func (w *Workspace) PullJob(slug, scriptType, id, name, cron, timezone, description string, timeoutSeconds int, overwrite bool) (filePath string, created bool, err error) {
	if err := ValidateSlug(slug); err != nil {
		return "", false, err
	}
	slug = strings.TrimSpace(strings.ToLower(slug))

	// 1. Try finding by unique ID first to avoid duplicate files when name/slug changed
	var existingPath string
	var found bool
	if id != "" {
		existingPath, found = w.FindScriptFileByID(id)
	}

	// 2. Fallback to finding by slug
	if !found {
		existingPath, found = w.FindScriptFile(slug)
	}

	if found {
		// If local filename does not match server slug, rename to server slug (without extension)
		targetPath := filepath.Join(w.JobsDir(), slug)
		if existingPath != targetPath {
			if _, statErr := os.Stat(targetPath); statErr == nil {
				_ = os.Remove(targetPath)
			}
			if renameErr := os.Rename(existingPath, targetPath); renameErr != nil {
				return existingPath, false, fmt.Errorf("failed to rename %s to %s: %w", existingPath, targetPath, renameErr)
			}
			existingPath = targetPath
		}

		if err := w.UpdateScriptHeader(existingPath, id, name, cron, timezone, description, timeoutSeconds); err != nil {
			return existingPath, false, fmt.Errorf("failed to update script header for %s: %w", existingPath, err)
		}
		return existingPath, false, nil
	}

	return w.CreateScript(slug, scriptType, id, name, cron, timezone, description, timeoutSeconds)
}

func (w *Workspace) RemoveJob(slug string) (removedFiles []string, removedFromJSON bool, err error) {
	if err := ValidateSlug(slug); err != nil {
		return nil, false, err
	}
	slug = strings.TrimSpace(strings.ToLower(slug))

	targetFile := filepath.Join(w.JobsDir(), slug)
	if _, statErr := os.Stat(targetFile); statErr == nil {
		if rmErr := os.Remove(targetFile); rmErr == nil {
			removedFiles = append(removedFiles, targetFile)
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
			if strings.HasPrefix(name, ".") {
				continue
			}
			slug := strings.ToLower(name)
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
	if err := ValidateSlug(slug); err != nil {
		return nil, err
	}
	slug = strings.TrimSpace(slug)

	if spec, ok := w.jobs[slug]; ok && len(spec.Command) > 0 {
		return w.expandArgv(spec.Command), nil
	}

	jobsDir := w.JobsDir()
	target := filepath.Join(jobsDir, slug)
	if info, statErr := os.Stat(target); statErr == nil && !info.IsDir() {
		return []string{target}, nil
	}

	return nil, fmt.Errorf("no local executable for job %q in %s (add jobs/%s or jobs.json)", slug, jobsDir, slug)
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
