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

func (w *Workspace) Resolve(slug string) (argv []string, err error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("job slug is empty")
	}

	if spec, ok := w.jobs[slug]; ok && len(spec.Command) > 0 {
		return w.expandArgv(spec.Command), nil
	}

	jobsDir := filepath.Join(w.Root, "jobs")
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
