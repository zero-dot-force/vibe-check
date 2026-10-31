package ripwireadapter

import (
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Error-budget warning tests — openspec/changes/ripwire-adapter task 4.5.
// They pin the spike-measured budget text per language and the
// any-warning ⇒ partial status rule the graph assembly applies.

func TestErrorBudgetWarning_PerLanguage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		language string
		wantText []string
	}{
		{languagePython, []string{"1.9%", "2.2%"}},
		{languageTypeScript, []string{"0.16%"}},
	}

	for _, tc := range cases {
		t.Run(tc.language, func(t *testing.T) {
			w := errorBudgetWarning(tc.language)
			if w.Code != "error-budget" {
				t.Errorf("code: got %q, want error-budget", w.Code)
			}
			for _, want := range tc.wantText {
				if !strings.Contains(w.Message, want) {
					t.Errorf("message %q does not carry budget %q", w.Message, want)
				}
			}
			if w.Module != "" {
				t.Errorf("module: got %q, want empty (global warning)", w.Module)
			}
		})
	}
}

func TestErrorBudgetWarning_UnknownLanguage(t *testing.T) {
	t.Parallel()

	w := errorBudgetWarning("cobol")
	if w.Code != "error-budget" {
		t.Errorf("code: got %q, want error-budget", w.Code)
	}
	if !strings.Contains(w.Message, "cobol") {
		t.Errorf("message %q does not name the language", w.Message)
	}
}

func TestGraphStatusForWarnings(t *testing.T) {
	t.Parallel()

	if got := graphStatusForWarnings(nil); got != metrics.StatusComplete {
		t.Errorf("no warnings: got %q, want complete", got)
	}
	if got := graphStatusForWarnings([]metrics.Warning{}); got != metrics.StatusComplete {
		t.Errorf("empty warnings: got %q, want complete", got)
	}
	warned := []metrics.Warning{{Code: "error-budget", Message: "m"}}
	if got := graphStatusForWarnings(warned); got != metrics.StatusPartial {
		t.Errorf("with warnings: got %q, want partial", got)
	}
}
