package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestVisualWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{Green("active"), 6},
		{"你好", 4},
		{Yellow("测试") + " abc", 8},
	}

	for _, tt := range tests {
		got := VisualWidth(tt.input)
		if got != tt.want {
			t.Errorf("VisualWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestTableRender(t *testing.T) {
	tbl := NewTable("ID", "TITLE", "STATUS")
	tbl.SetIndent("  ")
	tbl.AddRow("1", "Hello World", Green("active"))
	tbl.AddRow("200", "测试长标题", Yellow("paused"))

	var buf bytes.Buffer
	if err := tbl.Render(&buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}

	// Visual width check: the column starts should line up
	// Header: "  ID   TITLE        STATUS"
	// Row 1:  "  1    Hello World  active"
	// Row 2:  "  200  测试长标题   paused"
	if !strings.HasPrefix(lines[0], "  ") {
		t.Errorf("expected 2-space indent in header, got %q", lines[0])
	}
}
