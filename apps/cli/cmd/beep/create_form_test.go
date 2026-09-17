package beep

import "testing"

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
