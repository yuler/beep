package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExecutableAndJSON(t *testing.T) {
	root := t.TempDir()
	jobsDir := filepath.Join(root, "jobs")
	if err := os.MkdirAll(jobsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(jobsDir, "intranet-http.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	jobsJSON := `{"jobs":{"py-check":{"command":["python3","jobs/check.py"]}}}`
	if err := os.WriteFile(filepath.Join(root, "jobs.json"), []byte(jobsJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	argv, err := ws.Resolve("intranet-http")
	if err != nil {
		t.Fatal(err)
	}
	if argv[0] != script {
		t.Fatalf("expected %s, got %v", script, argv)
	}

	argv, err = ws.Resolve("py-check")
	if err != nil {
		t.Fatal(err)
	}
	if len(argv) != 2 || argv[0] != "python3" {
		t.Fatalf("unexpected argv %v", argv)
	}

	if _, err := ws.Resolve("missing"); err == nil {
		t.Fatal("expected missing job error")
	}
}

func TestCreateScriptAndListJobs(t *testing.T) {
	root := t.TempDir()
	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	path, created, err := ws.CreateScript("my-ping", "sh")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected created to be true")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected script file to exist at %s", path)
	}

	// Calling again should not overwrite
	_, created2, err := ws.CreateScript("my-ping", "sh")
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatalf("expected created to be false for existing script")
	}

	// Test Python script creation
	pyPath, createdPy, err := ws.CreateScript("py-worker", "py")
	if err != nil {
		t.Fatal(err)
	}
	if !createdPy || filepath.Ext(pyPath) != ".py" {
		t.Fatalf("expected .py script created")
	}

	jobs, err := ws.ListJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}
