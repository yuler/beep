package beep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"beep/internal/client"

	"github.com/charmbracelet/huh"
)

func TestShouldPromptBeepCreateForm(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		schedule   string
		argCount   int
		aiFallback bool
		wantPrompt bool
	}{
		{
			name:       "ai fallback with title from args must still prompt",
			title:      "哈哈",
			schedule:   "",
			argCount:   1,
			aiFallback: true,
			wantPrompt: true,
		},
		{
			name:       "empty interactive create prompts form",
			title:      "",
			schedule:   "",
			argCount:   0,
			aiFallback: false,
			wantPrompt: true,
		},
		{
			name:       "title from args without fallback skips form",
			title:      "哈哈",
			schedule:   "",
			argCount:   1,
			aiFallback: false,
			wantPrompt: false,
		},
		{
			name:       "explicit schedule with title skips form",
			title:      "Standup",
			schedule:   "cron",
			argCount:   1,
			aiFallback: false,
			wantPrompt: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldPromptBeepCreateForm(tc.title, tc.schedule, tc.argCount, tc.aiFallback)
			if got != tc.wantPrompt {
				t.Fatalf("shouldPromptBeepCreateForm(...) = %v, want %v", got, tc.wantPrompt)
			}
		})
	}
}

func TestIsUserAbort(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	activeCtx := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		err       error
		wantAbort bool
	}{
		{
			name:      "nil error and active ctx is not abort",
			ctx:       activeCtx,
			err:       nil,
			wantAbort: false,
		},
		{
			name:      "canceled ctx is abort",
			ctx:       canceledCtx,
			err:       nil,
			wantAbort: true,
		},
		{
			name:      "context.Canceled error is abort",
			ctx:       activeCtx,
			err:       context.Canceled,
			wantAbort: true,
		},
		{
			name:      "wrapped context.Canceled error is abort",
			ctx:       activeCtx,
			err:       fmt.Errorf("request failed: %w", context.Canceled),
			wantAbort: true,
		},
		{
			name:      "huh.ErrUserAborted is abort",
			ctx:       activeCtx,
			err:       huh.ErrUserAborted,
			wantAbort: true,
		},
		{
			name:      "user aborted text is abort",
			ctx:       activeCtx,
			err:       errors.New("prompt: user aborted"),
			wantAbort: true,
		},
		{
			name:      "cancelled text is abort",
			ctx:       activeCtx,
			err:       errors.New("action cancelled by user"),
			wantAbort: true,
		},
		{
			name:      "ai proposal failure is not abort",
			ctx:       activeCtx,
			err:       &aiProposalError{err: errors.New("ai model timeout")},
			wantAbort: false,
		},
		{
			name:      "generic network error is not abort",
			ctx:       activeCtx,
			err:       errors.New("connection refused"),
			wantAbort: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isUserAbort(tc.ctx, tc.err)
			if got != tc.wantAbort {
				t.Fatalf("isUserAbort() = %v, want %v", got, tc.wantAbort)
			}
		})
	}
}

func TestFormatBeepSchedule(t *testing.T) {
	bRecurring := &client.Beep{
		Kind: "recurring",
		Cron: "0 9 * * 1-5",
	}
	got := FormatBeepSchedule(bRecurring)
	want := "cron: 0 9 * * 1-5 (Every weekday at 09:00)"
	if got != want {
		t.Errorf("FormatBeepSchedule() = %q, want %q", got, want)
	}

	bInstant := &client.Beep{
		Kind: "once",
	}
	gotInstant := FormatBeepSchedule(bInstant)
	if !strings.Contains(gotInstant, "instant") {
		t.Errorf("FormatBeepSchedule() = %q, expected 'instant'", gotInstant)
	}
}

