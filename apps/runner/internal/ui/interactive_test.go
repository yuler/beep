package ui

import (
	"strings"
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

	// 5. Slug rename (same id, different slug) is modified
	ljRenamed := &workspace.LocalJob{
		Slug:           "health-check-v2",
		Name:           "Health Check",
		Cron:           "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		TimeoutSeconds: 30,
		Description:    "Check health",
		ID:             "job-1",
	}
	sjOldSlug := &client.ServerJob{
		ID:             "job-1",
		Slug:           "health-check",
		Name:           "Health Check",
		Cron:           "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		TimeoutSeconds: 30,
		Config: map[string]any{
			"description": "Check health",
		},
	}
	item = CompareJob("health-check-v2", ljRenamed, sjOldSlug)
	if item.Status != StatusModified {
		t.Fatalf("expected StatusModified for slug rename, got %s", item.Status)
	}
	foundSlugDiff := false
	for _, d := range item.Diffs {
		if strings.Contains(d, "slug:") {
			foundSlugDiff = true
		}
	}
	if !foundSlugDiff {
		t.Fatalf("expected slug diff, got %v", item.Diffs)
	}
}

func TestPairJobsMatchesRenamedSlugByID(t *testing.T) {
	local := []workspace.LocalJob{
		{Slug: "new-name", Name: "Health", Cron: "*/5 * * * *", ID: "abc", Timezone: "UTC"},
	}
	server := []*client.ServerJob{
		{ID: "abc", Slug: "old-name", Name: "Health", Cron: "*/5 * * * *", Timezone: "UTC"},
	}
	items := PairJobs(local, server)
	if len(items) != 1 {
		t.Fatalf("expected 1 paired item, got %d", len(items))
	}
	if items[0].Status != StatusModified {
		t.Fatalf("expected modified, got %s", items[0].Status)
	}
	if items[0].LocalJob == nil || items[0].ServerJob == nil {
		t.Fatal("expected both local and server set")
	}
}
