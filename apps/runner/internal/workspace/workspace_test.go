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

	path, created, err := ws.CreateScript("my-ping", "sh", "job_001", "My Custom Ping", "*/2 * * * *", "Asia/Shanghai", "Health check for ping", 45)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected created to be true")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected script file to exist at %s", path)
	}

	// Verify metadata parsing
	meta, err := ParseScriptMetadata(path)
	if err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if meta.ID != "job_001" {
		t.Fatalf("expected id 'job_001', got %q", meta.ID)
	}
	if meta.Name != "My Custom Ping" {
		t.Fatalf("expected name 'My Custom Ping', got %q", meta.Name)
	}
	if meta.Cron != "*/2 * * * *" {
		t.Fatalf("expected cron '*/2 * * * *', got %q", meta.Cron)
	}
	if meta.Timezone != "Asia/Shanghai" {
		t.Fatalf("expected timezone 'Asia/Shanghai', got %q", meta.Timezone)
	}
	if meta.Description != "Health check for ping" {
		t.Fatalf("expected description 'Health check for ping', got %q", meta.Description)
	}
	if meta.TimeoutSeconds != 45 {
		t.Fatalf("expected timeout 45, got %d", meta.TimeoutSeconds)
	}

	// Calling again should not overwrite
	_, created2, err := ws.CreateScript("my-ping", "sh", "", "", "", "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatalf("expected created to be false for existing script")
	}

	// Test Python script creation
	pyPath, createdPy, err := ws.CreateScript("py-worker", "py", "job_002", "", "0 * * * *", "UTC", "", 60)
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

	// Test RemoveJob
	removedFiles, inJSON, err := ws.RemoveJob("my-ping")
	if err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}
	if len(removedFiles) != 1 || inJSON {
		t.Fatalf("expected 1 removed file, got %v", removedFiles)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted")
	}

	jobsAfter, err := ws.ListJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobsAfter) != 1 {
		t.Fatalf("expected 1 job after remove, got %d", len(jobsAfter))
	}
}

func TestPullJob(t *testing.T) {
	root := t.TempDir()
	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	// Pull new job
	filePath, created, err := ws.PullJob("cloud-db-check", "sh", "job_db_1", "Cloud DB Check", "*/10 * * * *", "UTC", "Ping remote DB", 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected created to be true for new job")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file to exist at %s", filePath)
	}

	// Pull existing job updates the header and metadata without losing body
	_, created2, err := ws.PullJob("cloud-db-check", "sh", "job_db_1_updated", "Cloud DB Check Updated", "0 * * * *", "Asia/Tokyo", "Updated DB check", 60, false)
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatalf("expected created to be false for existing job update")
	}

	meta, err := ParseScriptMetadata(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if meta.ID != "job_db_1_updated" || meta.Name != "Cloud DB Check Updated" || meta.Cron != "0 * * * *" || meta.Timezone != "Asia/Tokyo" || meta.TimeoutSeconds != 60 || meta.Description != "Updated DB check" {
		t.Fatalf("unexpected pulled metadata: %+v", meta)
	}
}
