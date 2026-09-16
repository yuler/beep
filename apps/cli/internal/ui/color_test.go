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

func TestTruncate(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := Truncate("hello world", 8); got != "hello..." {
		t.Errorf("expected 'hello...', got %q", got)
	}
	// Multi-byte Chinese characters
	chinese := "这是一个非常长的测试提醒标题用于验证截断"
	truncated := Truncate(chinese, 10)
	if len([]rune(truncated)) != 10 {
		t.Errorf("expected 10 runes, got %d runes (%q)", len([]rune(truncated)), truncated)
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Errorf("expected suffix '...', got %q", truncated)
	}
	// Small maxRunes
	if got := Truncate("abcde", 2); got != "ab" {
		t.Errorf("expected 'ab', got %q", got)
	}
}
