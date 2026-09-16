package ui

import (
	"strings"
	"testing"
)

func TestColorFormatting(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	res := Green("hello")
	if !strings.Contains(res, "\033[32m") || !strings.Contains(res, "hello") {
		t.Fatalf("expected green ANSI code, got %q", res)
	}

	success := Success("done %d", 1)
	if !strings.Contains(success, "✓") || !strings.Contains(success, "done 1") {
		t.Fatalf("expected success prefix, got %q", success)
	}

	SetEnabled(false)
	resNoColor := Green("hello")
	if resNoColor != "hello" {
		t.Fatalf("expected plain string when disabled, got %q", resNoColor)
	}
}

func TestPrintErrorList(t *testing.T) {
	SetEnabled(false)
	// Should not panic on empty or populated lists
	PrintErrorList("Creation failed", nil)
	PrintErrorList("Creation failed", []string{"Title can't be blank", "Run at must be in the future"})
}
