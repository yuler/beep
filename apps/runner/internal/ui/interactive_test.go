package ui

import (
	"testing"

	"beep-runner/internal/client"
	"beep-runner/internal/workspace"
)

func TestCompareJob(t *testing.T) {
	// 1. Local only
	lj := &workspace.LocalJob{
		Slug: "local-only",
		Name: "Local Only",
		Cron: "*/5 * * * *",
	}
	item := CompareJob("local-only", lj, nil)
	if item.Status != StatusLocalOnly {
		t.Fatalf("expected StatusLocalOnly, got %s", item.Status)
	}

	// 2. Remote only
	sj := &client.ServerJob{
		Slug: "remote-only",
		Name: "Remote Only",
		Cron: "0 * * * *",
	}
	item = CompareJob("remote-only", nil, sj)
	if item.Status != StatusRemoteOnly {
		t.Fatalf("expected StatusRemoteOnly, got %s", item.Status)
	}

	// 3. Synced
	ljSynced := &workspace.LocalJob{
		Slug:           "health-check",
		Name:           "Health Check",
		Cron:           "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		TimeoutSeconds: 30,
		Description:    "Check health",
	}
	sjSynced := &client.ServerJob{
		Slug:           "health-check",
		Name:           "Health Check",
		Cron:           "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		TimeoutSeconds: 30,
		Config: map[string]any{
			"description": "Check health",
		},
	}
	item = CompareJob("health-check", ljSynced, sjSynced)
	if item.Status != StatusSynced {
		t.Fatalf("expected StatusSynced, got %s, diffs: %v", item.Status, item.Diffs)
	}

	// 4. Modified
	sjModified := &client.ServerJob{
		Slug:           "health-check",
		Name:           "Health Check Server",
		Cron:           "*/10 * * * *",
		Timezone:       "UTC",
		TimeoutSeconds: 60,
		Config: map[string]any{
			"description": "Check health updated",
		},
	}
	item = CompareJob("health-check", ljSynced, sjModified)
	if item.Status != StatusModified {
		t.Fatalf("expected StatusModified, got %s", item.Status)
	}
	if len(item.Diffs) != 5 {
		t.Fatalf("expected 5 diffs, got %d: %v", len(item.Diffs), item.Diffs)
	}
}
