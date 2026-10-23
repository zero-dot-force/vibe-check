package report

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// withProvenance attaches a provenance object to a graph fixture and fails
// the test unless the augmented graph still passes metrics.Validate, keeping
// the every-fixture-is-validated discipline of mustGraph. generatedAt is
// carried even though rendering MUST NOT emit it — the no-timestamp tests
// assert its absence.
func withProvenance(t *testing.T, g *metrics.ModuleGraph, producer, version, generatedAt string) *metrics.ModuleGraph {
	t.Helper()
	g.Provenance = &metrics.Provenance{
		Producer:    producer,
		Version:     version,
		GeneratedAt: generatedAt,
		Input:       metrics.ProvenanceInput{Path: ".", ModulePath: "example.com/mod"},
	}
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := metrics.Validate(data); err != nil {
		t.Fatalf("fixture failed schema validation: %v", err)
	}
	return g
}

// statusSummaryLine returns the report line that begins with an overall
// status word, failing the test when no such line exists. No other report
// line can begin with a status word: headings start with "#", the legend
// with 📖, and recommendation items with a list number.
func statusSummaryLine(t *testing.T, out string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, StatusAllClear) || strings.HasPrefix(line, StatusNeedsAttention) {
			return line
		}
	}
	t.Fatalf("no status summary line begins with a status word in output:\n%s", out)
	return ""
}

// cleanMods builds the two-module fixture shared by the header, legend,
// no-timestamp, and exact-output tests: no rule fires (the all-clear path),
// and the aggregate means are exact in binary floating point.
func cleanMods(t *testing.T) []metrics.ModuleResult {
	t.Helper()
	mods := []metrics.ModuleResult{
		mod("pkg/alpha", 1, 1, 0.25, 0, 0.125, 1),
		mod("pkg/beta", 2, 2, 0.45, 0, 0.375, 2),
	}
	mods[0].TotalLines = 120
	mods[1].TotalLines = 180
	return mods
}

// violationMods builds the two-module fixture shared by the violations
// tests: pkg/monolith fires the red lcom4-split rule (LCOM4 5) and pkg/hub
// fires the yellow ce-decouple rule (Ce 21); no other rule fires. Both
// distances are 0.25 so the mean is exact in binary floating point.
func violationMods(t *testing.T) []metrics.ModuleResult {
	t.Helper()
	mods := []metrics.ModuleResult{
		mod("pkg/monolith", 2, 2, 0.5, 0, 0.25, 5),
		mod("pkg/hub", 1, 21, 0.8, 0, 0.25, 1),
	}
	mods[0].TotalLines = 700
	mods[1].TotalLines = 500
	return mods
}

// TestRecommendationSeverity_Emoji pins the severity-to-emoji mapping of the
// report's fixed emoji vocabulary: the machine-readable strings red, yellow,
// and green render as 🔴, 🟡, and 🟢, and unknown severities render as the
// empty string.
func TestRecommendationSeverity_Emoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity RecommendationSeverity
		want     string
	}{
		{SeverityRed, "🔴"},
		{SeverityYellow, "🟡"},
		{SeverityGreen, "🟢"},
		{RecommendationSeverity("bogus"), ""},
	}
	for _, tt := range tests {
		if got := tt.severity.Emoji(); got != tt.want {
			t.Errorf("%q.Emoji(): got %q, want %q", tt.severity, got, tt.want)
		}
	}
}

// TestHasViolations pins the violation predicate that drives the status word
// and the report command's exit code: red and yellow are violations, green
// is not, and an empty list is not.
func TestHasViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		recs []Recommendation
		want bool
	}{
		{name: "nil list", recs: nil, want: false},
		{name: "empty list", recs: []Recommendation{}, want: false},
		{name: "green only", recs: []Recommendation{wantAllClear}, want: false},
		{
			name: "red only",
			recs: []Recommendation{{Severity: SeverityRed, Rule: RuleLCOM4Split, Message: "m"}},
			want: true,
		},
		{
			name: "yellow only",
			recs: []Recommendation{{Severity: SeverityYellow, Rule: RuleCeDecouple, Message: "m"}},
			want: true,
		},
		{
			name: "green yellow red",
			recs: []Recommendation{
				wantAllClear,
				{Severity: SeverityYellow, Rule: RuleInstabilityProtect, Message: "m"},
				{Severity: SeverityRed, Rule: RuleCycleBreak, Message: "m"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasViolations(tt.recs); got != tt.want {
				t.Errorf("HasViolations(%s): got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestReportStatus pins the status-word contract: "Needs Attention" when any
// red or yellow recommendation fires, "All Clear" otherwise — including the
// empty list.
func TestReportStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		recs []Recommendation
		want string
	}{
		{name: "empty list is all clear", recs: nil, want: StatusAllClear},
		{name: "green only is all clear", recs: []Recommendation{wantAllClear}, want: StatusAllClear},
		{
			name: "red fires needs attention",
			recs: []Recommendation{{Severity: SeverityRed, Rule: RuleLCOM4Split, Message: "m"}},
			want: StatusNeedsAttention,
		},
		{
			name: "yellow fires needs attention",
			recs: []Recommendation{{Severity: SeverityYellow, Rule: RuleInstabilityIsolate, Message: "m"}},
			want: StatusNeedsAttention,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ReportStatus(tt.recs); got != tt.want {
				t.Errorf("ReportStatus(%s): got %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestRenderSummary_HeaderFields pins the report spec's header-contents
// scenario: the first heading contains the 🏗️ report title, and the header
// line carries the producer name and version, the package count, and the
// total analyzed lines taken from the snapshot's provenance.
func TestRenderSummary_HeaderFields(t *testing.T) {
	t.Parallel()

	g := withProvenance(t, mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
		"vibe-check", "1.2.3", "2026-01-15T10:30:00Z")
	out := RenderSummary(g, RenderOptions{})

	if !strings.HasPrefix(out, "# 🏗️ Architectural Health Report\n") {
		t.Errorf("report does not open with the title heading: got prefix %q", firstLine(out))
	}
	wantHeader := "**Provenance**: vibe-check 1.2.3 | 2 packages | 300 lines"
	if !strings.Contains(out, wantHeader) {
		t.Errorf("header line: output does not contain %q", wantHeader)
	}
	if !strings.Contains(out, "## Prioritized Recommendations") {
		t.Errorf("summary output does not contain the recommendations section heading")
	}
}

// TestRenderSummary_NoTimestamp pins the byte-stability rule that no
// generation timestamp appears anywhere in the output, even though the
// snapshot's provenance carries one.
func TestRenderSummary_NoTimestamp(t *testing.T) {
	t.Parallel()

	g := withProvenance(t, mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
		"vibe-check", "1.2.3", "2026-01-15T10:30:00Z")
	out := RenderSummary(g, RenderOptions{})

	if strings.Contains(out, "2026-01-15T10:30:00Z") {
		t.Errorf("output contains the snapshot's generatedAt timestamp: %q", "2026-01-15T10:30:00Z")
	}
	rfc3339 := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	if match := rfc3339.FindString(out); match != "" {
		t.Errorf("output contains an RFC3339-shaped timestamp %q", match)
	}
}

// TestRenderSummary_Legend pins the fixed-formatting-vocabulary scenario:
// every report renders the fixed legend, it defines all six metric terms
// (Instability, Abstractness, Distance, LCOM4, Ca, Ce), and it links the
// metrics guide by absolute URL.
func TestRenderSummary_Legend(t *testing.T) {
	t.Parallel()

	out := RenderSummary(mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil), RenderOptions{})

	if !strings.Contains(out, reportLegend) {
		t.Errorf("output does not contain the fixed legend line")
	}
	for _, term := range []string{"Instability =", "Abstractness =", "Distance =", "LCOM4 =", "Ca =", "Ce ="} {
		if !strings.Contains(reportLegend, term) {
			t.Errorf("legend does not define the metric term %q", term)
		}
	}
	if !strings.Contains(reportLegend, metricsGuideURL) {
		t.Errorf("legend does not link the metrics guide URL %q", metricsGuideURL)
	}
}

// TestRenderSummary_StatusLine pins the status-summary scenarios: a clean
// graph's line begins with "All Clear" and a violations graph's line begins
// with "Needs Attention", each followed by the package count, average
// instability, average distance, max LCOM4, and cycle count.
func TestRenderSummary_StatusLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		want string
	}{
		{
			name: "clean graph is all clear",
			g:    mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
			want: "All Clear | 2 packages | Avg Instability: 0.35 | Avg Distance: 0.25 | Max LCOM4: 2 | Cycles: 0",
		},
		{
			name: "red cycle fires needs attention",
			g: mustGraph(t, metrics.StatusComplete,
				[]metrics.ModuleResult{
					mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1),
					mod("pkg/b", 1, 1, 0.5, 0, 0.25, 1),
				},
				[]metrics.Cycle{{"pkg/a", "pkg/b"}}, nil),
			want: "Needs Attention | 2 packages | Avg Instability: 0.50 | Avg Distance: 0.25 | Max LCOM4: 1 | Cycles: 1",
		},
		{
			name: "yellow only fires needs attention",
			g: mustGraph(t, metrics.StatusComplete,
				[]metrics.ModuleResult{mod("pkg/hub", 1, 21, 0.8, 0, 0.2, 1)},
				nil, nil),
			want: "Needs Attention | 1 package | Avg Instability: 0.80 | Avg Distance: 0.20 | Max LCOM4: 1 | Cycles: 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := RenderSummary(tt.g, RenderOptions{})
			if got := statusSummaryLine(t, out); got != tt.want {
				t.Errorf("status summary line: got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRenderSummary_Recommendations pins the recommendation rendering: the
// engine's prioritized list renders as a numbered markdown list, each entry
// carrying its severity emoji and its message verbatim, ordered red before
// yellow before green; the all-clear path renders exactly one green entry.
func TestRenderSummary_Recommendations(t *testing.T) {
	t.Parallel()

	t.Run("all three severities in priority order", func(t *testing.T) {
		t.Parallel()
		// 16 duplicated lines over 400 total lines is 4%: strictly between
		// 0% and 5%, so the green watch rule fires instead of the red
		// eliminate rule.
		dups := []metrics.Duplication{{
			ModulePath: "pkg/dup",
			Blocks: []metrics.DuplicateBlock{
				dupBlock("dup/a.go", 10, 25, 8),
				dupBlock("dup/b.go", 10, 25, 8),
			},
			Similarity: 0.9,
		}}
		mods := append(violationMods(t), dupMod("pkg/dup", 400, dups...))
		g := mustGraph(t, metrics.StatusComplete, mods, nil, nil)

		out := RenderSummary(g, RenderOptions{})
		wantLines := []string{
			"1. 🔴 Split `pkg/monolith` — LCOM4 of `5` suggests multiple unrelated responsibilities",
			"2. 🟡 Decouple `pkg/hub` — `21` efferent couplings; consider interface extraction",
			"3. 🟢 Watch duplication in `dup/a.go`, `dup/b.go` — `2` blocks at `90`% similarity",
		}
		for _, want := range wantLines {
			if !strings.Contains(out, want+"\n") {
				t.Errorf("recommendations: output does not contain line %q", want)
			}
		}
		for i := range wantLines {
			for j := i + 1; j < len(wantLines); j++ {
				if gotI, gotJ := strings.Index(out, wantLines[i]), strings.Index(out, wantLines[j]); gotI > gotJ {
					t.Errorf("recommendation %d renders after recommendation %d", j+1, i+1)
				}
			}
		}
		if strings.Contains(out, wantAllClear.Message) {
			t.Errorf("all-clear rendered alongside the fired watch rule; the watch rule must suppress it")
		}
		if got := statusSummaryLine(t, out); !strings.HasPrefix(got, StatusNeedsAttention) {
			t.Errorf("status summary line: got %q, want it to begin with %q", got, StatusNeedsAttention)
		}
	})

	t.Run("all clear renders exactly one green entry", func(t *testing.T) {
		t.Parallel()
		g := mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil)

		out := RenderSummary(g, RenderOptions{})
		want := "1. 🟢 " + wantAllClear.Message + "\n"
		if !strings.Contains(out, want) {
			t.Errorf("all-clear recommendation: output does not contain %q", want)
		}
		for _, emoji := range []string{"🔴", "🟡"} {
			if strings.Contains(out, emoji) {
				t.Errorf("clean report contains violation emoji %q", emoji)
			}
		}
		if got, wantCount := strings.Count(out, "\n🟢")+strings.Count(out, ". 🟢"), 1; got != wantCount {
			t.Errorf("green recommendation count: got %v, want %v", got, wantCount)
		}
	})
}

// TestRenderSummary_NumberFormats pins the report's fixed formatting
// vocabulary on the summary surfaces: ratio metrics with %.2f (the mean of
// 0.25, 0.25, and 0.50 is 1/3, which %.2f renders as 0.33 where %v would
// render full float precision) and LCOM4, cycle, package, and line counts
// as integers.
func TestRenderSummary_NumberFormats(t *testing.T) {
	t.Parallel()

	mods := []metrics.ModuleResult{
		mod("pkg/a", 1, 1, 0.25, 0, 0.1, 1),
		mod("pkg/b", 1, 1, 0.25, 0, 0.2, 2),
		mod("pkg/c", 1, 1, 0.50, 0, 0.4, 3),
	}
	mods[0].TotalLines = 100
	mods[1].TotalLines = 200
	mods[2].TotalLines = 300
	g := mustGraph(t, metrics.StatusComplete, mods, nil, nil)

	out := RenderSummary(g, RenderOptions{})
	for _, want := range []string{
		"All Clear | 3 packages | Avg Instability: 0.33 | Avg Distance: 0.23 | Max LCOM4: 3 | Cycles: 0",
		"3 packages | 600 lines",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("number format: output does not contain %q", want)
		}
	}
	if strings.Contains(out, "0.3333") {
		t.Errorf("instability rendered with more than two decimals")
	}
}

// TestRenderSummary_EmptyModuleSet pins the spec's empty-module-set scenario:
// zero modules render zero aggregates, the all-clear status, and the single
// green all-clear recommendation. A nil graph renders identically, mirroring
// the engine's nil handling.
func TestRenderSummary_EmptyModuleSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		opts RenderOptions
	}{
		{name: "zero modules", g: mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{}, nil, nil), opts: RenderOptions{}},
		{name: "nil graph", g: nil, opts: RenderOptions{FallbackProducer: "x-renderer", FallbackVersion: "0.9.0"}},
	}
	wantStatus := "All Clear | 0 packages | Avg Instability: 0.00 | Avg Distance: 0.00 | Max LCOM4: 0 | Cycles: 0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := RenderSummary(tt.g, tt.opts)
			if got := statusSummaryLine(t, out); got != wantStatus {
				t.Errorf("status summary line: got %q, want %q", got, wantStatus)
			}
			if want := "1. 🟢 " + wantAllClear.Message + "\n"; !strings.Contains(out, want) {
				t.Errorf("all-clear recommendation: output does not contain %q", want)
			}
			if strings.Contains(out, StatusNeedsAttention) {
				t.Errorf("empty module set rendered %q", StatusNeedsAttention)
			}
		})
	}
}

// TestRenderSummary_ProvenanceFallback pins the header's producer
// resolution: snapshot provenance wins when present, an omitted provenance
// falls back to the RenderOptions-injected rendering-binary identity, empty
// provenance fields fall back per field, and the zero options fall back to
// the package defaults.
func TestRenderSummary_ProvenanceFallback(t *testing.T) {
	t.Parallel()

	injected := RenderOptions{FallbackProducer: "my-renderer", FallbackVersion: "1.0.0"}
	twoMods := func(t *testing.T) *metrics.ModuleGraph {
		t.Helper()
		return mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil)
	}

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		opts RenderOptions
		want string
	}{
		{
			name: "snapshot provenance wins over options",
			g:    withProvenance(t, twoMods(t), "snapshot-tool", "9.9.9", "2026-01-15T10:30:00Z"),
			opts: injected,
			want: "**Provenance**: snapshot-tool 9.9.9 | 2 packages | 300 lines",
		},
		{
			name: "omitted provenance uses injected identity",
			g:    twoMods(t),
			opts: injected,
			want: "**Provenance**: my-renderer 1.0.0 | 2 packages | 300 lines",
		},
		{
			name: "omitted provenance with zero options uses defaults",
			g:    twoMods(t),
			opts: RenderOptions{},
			want: "**Provenance**: " + DefaultRenderProducer + " " + DefaultRenderVersion + " | 2 packages | 300 lines",
		},
		{
			name: "empty provenance fields fall back per field",
			g:    withProvenance(t, twoMods(t), "", "", "2026-01-15T10:30:00Z"),
			opts: injected,
			want: "**Provenance**: my-renderer 1.0.0 | 2 packages | 300 lines",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := RenderSummary(tt.g, tt.opts)
			if !strings.Contains(out, tt.want) {
				t.Errorf("header provenance: output does not contain %q", tt.want)
			}
		})
	}
}

// TestRenderSummary_ExactOutput pins the full byte format of the summary
// report — section order, separators, emoji, and every formatted value — for
// both the all-clear and the violations path. This is the format contract
// the detailed-mode renderer extends and the CLI golden tests build on.
func TestRenderSummary_ExactOutput(t *testing.T) {
	t.Parallel()

	legend := "📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md"

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		want string
	}{
		{
			name: "clean graph with provenance",
			g: withProvenance(t, mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
				"vibe-check", "1.2.3", "2026-01-15T10:30:00Z"),
			want: "# 🏗️ Architectural Health Report\n" +
				"\n" +
				"**Provenance**: vibe-check 1.2.3 | 2 packages | 300 lines\n" +
				"\n" +
				legend + "\n" +
				"\n" +
				"All Clear | 2 packages | Avg Instability: 0.35 | Avg Distance: 0.25 | Max LCOM4: 2 | Cycles: 0\n" +
				"\n" +
				"## Prioritized Recommendations\n" +
				"\n" +
				"1. 🟢 Architecture is healthy — no significant issues detected; keep monitoring\n",
		},
		{
			name: "violations graph without provenance",
			g:    mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil),
			want: "# 🏗️ Architectural Health Report\n" +
				"\n" +
				"**Provenance**: vibe-check dev | 2 packages | 1200 lines\n" +
				"\n" +
				legend + "\n" +
				"\n" +
				"Needs Attention | 2 packages | Avg Instability: 0.65 | Avg Distance: 0.25 | Max LCOM4: 5 | Cycles: 0\n" +
				"\n" +
				"## Prioritized Recommendations\n" +
				"\n" +
				"1. 🔴 Split `pkg/monolith` — LCOM4 of `5` suggests multiple unrelated responsibilities\n" +
				"2. 🟡 Decouple `pkg/hub` — `21` efferent couplings; consider interface extraction\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := RenderSummary(tt.g, RenderOptions{}); got != tt.want {
				t.Errorf("RenderSummary(%s):\ngot:\n%s\nwant:\n%s", tt.name, got, tt.want)
			}
		})
	}
}

// firstLine returns the first line of a multi-line string, for failure
// messages about a report's opening line.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// detailedFixture builds the rich detailed-mode fixture shared by the exact
// output, prefix, and section tests: two modules fed in reverse path order
// (pinning the canonical path-order rendering), one cycle, one duplication
// group of two blocks, one module-scoped warning, and one Zone of Uselessness
// package. Metric values are chosen so every detailed section has content
// while only the cycle rule (red, once per member) and the duplication watch
// rule (green) fire, and the aggregate means are exact in binary floating
// point.
func detailedFixture(t *testing.T) *metrics.ModuleGraph {
	t.Helper()
	alpha := mod("pkg/alpha", 12, 0, 0.06, 0.50, 0.44, 1)
	alpha.Zone = metrics.ZoneMainSequence
	alpha.TotalLines = 500
	omega := mod("pkg/omega", 0, 3, 0.94, 0.90, 0.20, 2)
	omega.Zone = metrics.ZoneOfUselessness
	omega.TotalLines = 300
	omega.Duplications = []metrics.Duplication{{
		ModulePath: "pkg/omega",
		Blocks: []metrics.DuplicateBlock{
			dupBlock("x/a.go", 10, 25, 16),
			dupBlock("x/b.go", 40, 55, 16),
		},
		Similarity: 0.9,
	}}
	warnings := []metrics.Warning{{
		Code:    "load-error",
		Message: "package pkg/gone: type information unavailable",
		Module:  "pkg/gone",
	}}
	return mustGraph(t, metrics.StatusComplete,
		[]metrics.ModuleResult{omega, alpha}, // reverse path order on purpose
		[]metrics.Cycle{{"pkg/alpha", "pkg/omega"}},
		warnings)
}

// TestRenderDetailed_SummaryPrefix pins the composability contract: detailed
// mode renders the summary mode's contents first, byte-for-byte, then
// continues with the detailed sections — the bytes immediately after the
// summary output are the blank line before the 📊 heading.
func TestRenderDetailed_SummaryPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		opts RenderOptions
	}{
		{name: "rich fixture", g: detailedFixture(t), opts: RenderOptions{}},
		{
			name: "clean with provenance",
			g: withProvenance(t, mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
				"vibe-check", "1.2.3", "2026-01-15T10:30:00Z"),
			opts: RenderOptions{},
		},
		{name: "violations", g: mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil), opts: RenderOptions{}},
		{name: "zero modules", g: mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{}, nil, nil), opts: RenderOptions{}},
		{name: "nil graph", g: nil, opts: RenderOptions{FallbackProducer: "x-renderer", FallbackVersion: "0.9.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			summary := RenderSummary(tt.g, tt.opts)
			detailed := RenderDetailed(tt.g, tt.opts)
			if !strings.HasPrefix(detailed, summary) {
				t.Errorf("detailed output does not start with the summary output:\nsummary:\n%s\ndetailed:\n%s", summary, detailed)
			}
			if !strings.HasPrefix(detailed[len(summary):], "\n## 📊 Package Metrics\n\n") {
				t.Errorf("detailed output does not continue with the package metrics heading after the summary sections: got %q",
					firstLine(detailed[len(summary):]))
			}
		})
	}
}

// TestRenderDetailed_SectionOrder pins the report spec's detailed-mode
// scenario: the seven detailed sections render after the summary contents,
// each marked with its emoji from the closed vocabulary, in the fixed
// section order.
func TestRenderDetailed_SectionOrder(t *testing.T) {
	t.Parallel()

	out := RenderDetailed(detailedFixture(t), RenderOptions{})
	headings := []string{
		"## 📊 Package Metrics",
		"## 🔗 Coupling Summary",
		"## 🧩 Cohesion Summary",
		"## 🔄 Circular Dependencies",
		"## 📋 Code Duplication",
		"## 🏥 Health Scorecard",
		"## ⚠️ Warnings",
	}
	last := strings.Index(out, "## Prioritized Recommendations")
	if last < 0 {
		t.Fatalf("detailed output does not contain the recommendations heading")
	}
	for _, h := range headings {
		i := strings.Index(out, h)
		if i < 0 {
			t.Errorf("detailed output does not contain heading %q", h)
			continue
		}
		if i <= last {
			t.Errorf("heading %q renders at %d, want after the previous section at %d", h, i, last)
		}
		last = i
	}
}

// TestRenderDetailed_PackageMetrics pins the package metrics table: full-word
// headers, one row per module in path order regardless of input order, the
// fixed number formats, and the emoji-labeled zone column for every defined
// zone. The unknown-zone fallback is pinned directly on zoneLabel because a
// zone outside the schema enum cannot ride through mustGraph.
func TestRenderDetailed_PackageMetrics(t *testing.T) {
	t.Parallel()

	wantHeader := "| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce | Zone |\n" +
		"|---------|-------------|--------------|----------|-------|----|----|------|\n"

	t.Run("populated table renders rows in path order", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(detailedFixture(t), RenderOptions{})
		if !strings.Contains(out, wantHeader) {
			t.Errorf("package metrics table header: output does not contain %q", wantHeader)
		}
		wantAlpha := "| `pkg/alpha` | 0.06 | 0.50 | 0.44 | 1 | 12 | 0 | 🟢 Main Sequence |"
		wantOmega := "| `pkg/omega` | 0.94 | 0.90 | 0.20 | 2 | 0 | 3 | 🔴 Zone of Uselessness |"
		for _, want := range []string{wantAlpha, wantOmega} {
			if !strings.Contains(out, want) {
				t.Errorf("package metrics row: output does not contain %q", want)
			}
		}
		if a, o := strings.Index(out, wantAlpha), strings.Index(out, wantOmega); a > o {
			t.Errorf("package metrics rows render out of path order: alpha at %d, omega at %d", a, o)
		}
	})

	t.Run("every defined zone renders its human label", func(t *testing.T) {
		t.Parallel()
		zoned := func(path string, zone metrics.Zone) metrics.ModuleResult {
			m := mod(path, 1, 1, 0.5, 0.25, 0.25, 1)
			m.Zone = zone
			return m
		}
		mods := []metrics.ModuleResult{
			zoned("pkg/delta", metrics.ZoneOfUselessness),
			zoned("pkg/gamma", metrics.ZoneOfPain),
			zoned("pkg/beta", metrics.ZoneNormal),
			zoned("pkg/alpha", metrics.ZoneMainSequence),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"| `pkg/alpha` | 0.50 | 0.25 | 0.25 | 1 | 1 | 1 | 🟢 Main Sequence |",
			"| `pkg/beta` | 0.50 | 0.25 | 0.25 | 1 | 1 | 1 | 🟡 Balanced |",
			"| `pkg/gamma` | 0.50 | 0.25 | 0.25 | 1 | 1 | 1 | 🔴 Zone of Pain |",
			"| `pkg/delta` | 0.50 | 0.25 | 0.25 | 1 | 1 | 1 | 🔴 Zone of Uselessness |",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("zone row: output does not contain %q", want)
			}
		}
	})

	t.Run("unknown zone falls back to the machine string", func(t *testing.T) {
		t.Parallel()
		if got := zoneLabel(metrics.Zone("mystery")); got != "mystery" {
			t.Errorf("zoneLabel(mystery): got %q, want %q", got, "mystery")
		}
	})

	t.Run("no modules renders the empty note", func(t *testing.T) {
		t.Parallel()
		for _, g := range []*metrics.ModuleGraph{
			nil,
			mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{}, nil, nil),
		} {
			out := RenderDetailed(g, RenderOptions{})
			if !strings.Contains(out, "## 📊 Package Metrics\n\nNo packages to analyze.\n") {
				t.Errorf("empty package metrics section: output does not contain the no-packages note:\n%s", out)
			}
		}
	})
}

// TestRenderDetailed_CouplingSummary pins the coupling summary facts: the
// max-Ca and max-Ce modules (first by path among ties), the leaf modules
// (Ce == 0), and the heavily-depended-on list with its exclusive Ca > 10
// bound — a module at exactly 10 is not heavily depended on.
func TestRenderDetailed_CouplingSummary(t *testing.T) {
	t.Parallel()

	t.Run("max ca, max ce, leaves, and heavily depended-on", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/cli", 0, 20, 0.5, 0, 0.25, 1),
			mod("pkg/db", 12, 0, 0.5, 0, 0.25, 1),
			mod("pkg/edge", 10, 1, 0.5, 0, 0.25, 1),
			mod("pkg/core", 15, 2, 0.5, 0, 0.25, 1),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"- **Max Ca**: `pkg/core` (15 dependents)",
			"- **Max Ce**: `pkg/cli` (20 dependencies)",
			"- **Leaves (Ce=0)**: `pkg/db`",
			"- **Heavily depended-on (Ca > 10)**: `pkg/core`, `pkg/db`",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("coupling summary: output does not contain %q", want)
			}
		}
	})

	t.Run("max ties name the first module in path order", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/bbb", 5, 7, 0.5, 0, 0.25, 1),
			mod("pkg/aaa", 5, 7, 0.5, 0, 0.25, 1),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"- **Max Ca**: `pkg/aaa` (5 dependents)",
			"- **Max Ce**: `pkg/aaa` (7 dependencies)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("coupling summary tie: output does not contain %q", want)
			}
		}
	})

	t.Run("empty lists render none", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{mod("pkg/mid", 3, 4, 0.5, 0, 0.25, 1)}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"- **Leaves (Ce=0)**: none",
			"- **Heavily depended-on (Ca > 10)**: none",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("coupling summary empty list: output does not contain %q", want)
			}
		}
	})

	t.Run("no modules renders the empty note", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(nil, RenderOptions{})
		if !strings.Contains(out, "## 🔗 Coupling Summary\n\nNo packages to analyze.\n") {
			t.Errorf("empty coupling section: output does not contain the no-packages note")
		}
	})
}

// TestRenderDetailed_CohesionSummary pins the cohesion summary: best (lowest
// LCOM4) and worst (highest LCOM4) modules — first by path among ties — and
// the review list at the LCOM4 >= 4 boundary (3 excluded, 4 included), with
// values, in path order.
func TestRenderDetailed_CohesionSummary(t *testing.T) {
	t.Parallel()

	t.Run("best, worst, and the review boundary", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/worst", 1, 1, 0.5, 0, 0.25, 7),
			mod("pkg/mid", 1, 1, 0.5, 0, 0.25, 3),
			mod("pkg/edge", 1, 1, 0.5, 0, 0.25, 4),
			mod("pkg/best", 1, 1, 0.5, 0, 0.25, 1),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"- **Best**: `pkg/best` (LCOM4=1)",
			"- **Worst**: `pkg/worst` (LCOM4=7)",
			"- **Packages needing review (LCOM4 ≥ 4)**: `pkg/edge` (4), `pkg/worst` (7)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("cohesion summary: output does not contain %q", want)
			}
		}
	})

	t.Run("ties name the first module in path order", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/bbb", 1, 1, 0.5, 0, 0.25, 2),
			mod("pkg/aaa", 1, 1, 0.5, 0, 0.25, 2),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		for _, want := range []string{
			"- **Best**: `pkg/aaa` (LCOM4=2)",
			"- **Worst**: `pkg/aaa` (LCOM4=2)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("cohesion tie: output does not contain %q", want)
			}
		}
	})

	t.Run("no module at the review threshold renders none", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{mod("pkg/ok", 1, 1, 0.5, 0, 0.25, 3)}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		if !strings.Contains(out, "- **Packages needing review (LCOM4 ≥ 4)**: none") {
			t.Errorf("cohesion summary: output does not contain the none review list")
		}
	})

	t.Run("no modules renders the empty note", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(nil, RenderOptions{})
		if !strings.Contains(out, "## 🧩 Cohesion Summary\n\nNo packages to analyze.\n") {
			t.Errorf("empty cohesion section: output does not contain the no-packages note")
		}
	})
}

// TestRenderDetailed_Cycles pins the circular dependency chains: each cycle
// renders as one closed chain per line — members joined with →, first member
// repeated to close the loop — inside a fenced code block, multiple chains
// ordered by their first path, and the zero-cycle green line for clean
// graphs. Zero-member cycles are schema-valid but name no members; they must
// not panic and do not render a chain.
func TestRenderDetailed_Cycles(t *testing.T) {
	t.Parallel()

	t.Run("one cycle renders a closed chain in a fence", func(t *testing.T) {
		t.Parallel()
		g := mustGraph(t, metrics.StatusComplete, nil, []metrics.Cycle{{"pkg/a", "pkg/b"}}, nil)
		out := RenderDetailed(g, RenderOptions{})
		want := "## 🔄 Circular Dependencies\n\n" +
			"🔴 **1 cycle detected:**\n\n" +
			"```\npkg/a → pkg/b → pkg/a\n```\n"
		if !strings.Contains(out, want) {
			t.Errorf("cycles section: output does not contain %q", want)
		}
	})

	t.Run("multiple cycles render one chain per line ordered by first path", func(t *testing.T) {
		t.Parallel()
		g := mustGraph(t, metrics.StatusComplete, nil,
			[]metrics.Cycle{{"pkg/c", "pkg/d"}, {"pkg/a", "pkg/b", "pkg/e"}}, nil)
		out := RenderDetailed(g, RenderOptions{})
		for _, want := range []string{
			"🔴 **2 cycles detected:**",
			"```\npkg/a → pkg/b → pkg/e → pkg/a\npkg/c → pkg/d → pkg/c\n```\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("cycles section: output does not contain %q", want)
			}
		}
	})

	t.Run("no cycles renders the green line", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil), RenderOptions{})
		if !strings.Contains(out, "## 🔄 Circular Dependencies\n\n🟢 No circular dependencies detected.\n") {
			t.Errorf("cycles section: output does not contain the no-cycles line")
		}
	})

	t.Run("zero-member cycle does not panic and renders no chain", func(t *testing.T) {
		t.Parallel()
		g := mustGraph(t, metrics.StatusComplete, nil, []metrics.Cycle{{}}, nil)
		out := RenderDetailed(g, RenderOptions{})
		if !strings.Contains(out, "🟢 No circular dependencies detected.") {
			t.Errorf("cycles section: a zero-member cycle should render the no-cycles line, got:\n%s", out)
		}
	})
}

// TestRenderDetailed_DuplicationTable pins the duplication table: one row
// per deduplicated block identity (the dupBlockKey semantics
// ComputeAggregates applies), with max-wins line counts and similarity, the
// cross-group occurrence count, rows ordered by file then line range, and a
// lead line whose overall percentage matches ComputeAggregates.
func TestRenderDetailed_DuplicationTable(t *testing.T) {
	t.Parallel()

	t.Run("deduplicated rows with occurrences and max-wins values", func(t *testing.T) {
		t.Parallel()
		// The x/a.go block appears in both groups: one row, two occurrences,
		// and the larger line count plus higher similarity win. A third group
		// references x/b.go with a higher similarity but a smaller line
		// count, pinning both max-wins directions. 44 deduplicated lines
		// over 440 total lines is 10.0%.
		dups := []metrics.Duplication{
			{
				ModulePath: "pkg/dup",
				Blocks: []metrics.DuplicateBlock{
					dupBlock("x/a.go", 10, 25, 16),
					dupBlock("x/b.go", 40, 55, 16),
				},
				Similarity: 0.9,
			},
			{
				ModulePath: "pkg/dup",
				Blocks: []metrics.DuplicateBlock{
					dupBlock("x/a.go", 10, 25, 20),
					dupBlock("x/c.go", 1, 9, 8),
				},
				Similarity: 0.75,
			},
			{
				ModulePath: "pkg/dup",
				Blocks: []metrics.DuplicateBlock{
					dupBlock("x/b.go", 40, 55, 10),
				},
				Similarity: 0.95,
			},
		}
		mods := []metrics.ModuleResult{dupMod("pkg/dup", 440, dups...)}
		g := mustGraph(t, metrics.StatusComplete, mods, nil, nil)
		out := RenderDetailed(g, RenderOptions{})

		wantLead := fmt.Sprintf("**3 duplicated blocks detected (overall %.1f%%):**", ComputeAggregates(g).DuplicationPercent)
		if !strings.Contains(out, wantLead) {
			t.Errorf("duplication lead line: output does not contain %q", wantLead)
		}
		if want := "(overall 10.0%)"; !strings.Contains(out, want) {
			t.Errorf("duplication lead line: output does not contain %q", want)
		}
		wantHeader := "| File | Lines | Similarity | Occurrences |\n" +
			"|------|-------|------------|-------------|\n"
		if !strings.Contains(out, wantHeader) {
			t.Errorf("duplication table header: output does not contain %q", wantHeader)
		}
		for _, want := range []string{
			"| `x/a.go:10-25` | 20 lines | 90% | 2 |",
			"| `x/b.go:40-55` | 16 lines | 95% | 2 |",
			"| `x/c.go:1-9` | 8 lines | 75% | 1 |",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("duplication row: output does not contain %q", want)
			}
		}
		// Dedup consistency with the engine: the table's row count is the
		// engine's deduplicated block count over the same module set.
		blocks, _, _ := duplicationFacts(mods)
		if got := strings.Count(out, " lines | "); got != blocks {
			t.Errorf("duplication table row count: got %v, want %v (duplicationFacts blocks)", got, blocks)
		}
	})

	t.Run("no duplicated blocks renders the empty line", func(t *testing.T) {
		t.Parallel()
		for _, g := range []*metrics.ModuleGraph{
			nil,
			mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
		} {
			out := RenderDetailed(g, RenderOptions{})
			if !strings.Contains(out, "## 📋 Code Duplication\n\nNo duplicated blocks detected.\n") {
				t.Errorf("empty duplication section: output does not contain the no-duplication line")
			}
		}
	})
}

// TestRenderDetailed_Scorecard pins the health scorecard: the five graded
// dimensions with their values, grades, and status emoji — reusing the task
// 1.2 classifier boundary pins (distance exactly 0.1 → B+, LCOM4 4 → C+, one
// cycle → F, duplication exactly 5% → C+, spread 1 → B) — and the overall
// status derived from the worst dimension grade.
func TestRenderDetailed_Scorecard(t *testing.T) {
	t.Parallel()

	wantHeader := "| Dimension | Value | Grade | Status |\n" +
		"|-----------|-------|-------|--------|\n"

	t.Run("mixed grades pin the spec's emoji scenario", func(t *testing.T) {
		t.Parallel()
		// Grades across the five dimensions: A (distance 0.05), B+ (max LCOM4
		// 2), F (one cycle), C+ (duplication exactly 5%: 50 deduplicated
		// lines over 1000 total), B (spread 1: one extreme-instability
		// module). Worst grade F → overall 🔴 Needs Attention.
		calm := mod("pkg/calm", 1, 1, 0.0, 0.90, 0.05, 1)
		calm.TotalLines = 500
		balanced := mod("pkg/balanced", 1, 1, 0.5, 0.90, 0.05, 2)
		balanced.TotalLines = 500
		balanced.Duplications = []metrics.Duplication{{
			ModulePath: "pkg/balanced",
			Blocks: []metrics.DuplicateBlock{
				dupBlock("d/a.go", 1, 50, 25),
				dupBlock("d/b.go", 1, 50, 25),
			},
			Similarity: 0.8,
		}}
		g := mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{calm, balanced},
			[]metrics.Cycle{{"pkg/balanced", "pkg/calm"}}, nil)
		out := RenderDetailed(g, RenderOptions{})
		for _, want := range []string{
			wantHeader,
			"| Avg Distance from Main Sequence | 0.05 | A | 🟢 |",
			"| Max LCOM4 | 2 | B+ | 🟢 |",
			"| Cycle Count | 1 | F | 🔴 |",
			"| Duplication % | 5.0% | C+ | 🟡 |",
			"| Instability Spread | 1 extremes | B | 🟡 |",
			"**Overall**: 🔴 Needs Attention",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("scorecard: output does not contain %q", want)
			}
		}
	})

	t.Run("distance boundary exactly 0.1 grades B+", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0.45, 0.1, 1),
			mod("pkg/b", 1, 1, 0.5, 0.45, 0.1, 1),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		if !strings.Contains(out, "| Avg Distance from Main Sequence | 0.10 | B+ | 🟢 |") {
			t.Errorf("scorecard distance boundary: output does not contain the B+ row")
		}
	})

	t.Run("lcom boundary exactly 4 grades C+", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{mod("pkg/a", 1, 1, 0.5, 0.45, 0.1, 4)}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		if !strings.Contains(out, "| Max LCOM4 | 4 | C+ | 🟡 |") {
			t.Errorf("scorecard LCOM boundary: output does not contain the C+ row")
		}
	})

	t.Run("worst grade B overall is mixed", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0.45, 0.25, 3),
			mod("pkg/b", 1, 1, 0.5, 0.45, 0.25, 2),
		}
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, mods, nil, nil), RenderOptions{})
		if !strings.Contains(out, "**Overall**: 🟡 Mixed") {
			t.Errorf("scorecard overall: output does not contain the mixed overall line")
		}
	})

	t.Run("empty graph grades all A and overall healthy", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(nil, RenderOptions{})
		for _, want := range []string{
			wantHeader,
			"| Avg Distance from Main Sequence | 0.00 | A | 🟢 |",
			"| Max LCOM4 | 0 | A | 🟢 |",
			"| Cycle Count | 0 | A | 🟢 |",
			"| Duplication % | 0.0% | A | 🟢 |",
			"| Instability Spread | 0 extremes | A | 🟢 |",
			"**Overall**: 🟢 Healthy",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("empty scorecard: output does not contain %q", want)
			}
		}
	})
}

// TestRenderDetailed_Warnings pins the warnings section: snapshot warnings
// reproduce with their module attribution, module-less warnings render bare,
// the section orders by module (module-less first) regardless of snapshot
// order, Zone of Uselessness packages render after the snapshot warnings in
// path order, and the empty case renders the no-warnings line.
func TestRenderDetailed_Warnings(t *testing.T) {
	t.Parallel()

	t.Run("warnings reproduce, sorted, then zone-of-uselessness packages", func(t *testing.T) {
		t.Parallel()
		warnings := []metrics.Warning{
			{Code: "c2", Message: "z note", Module: "pkg/z"},
			{Code: "c1", Message: "a note", Module: "pkg/a"},
			{Code: "global", Message: "global note"},
		}
		mods := []metrics.ModuleResult{
			mod("pkg/omega", 0, 1, 0.94, 0.90, 0.16, 1),
			mod("pkg/alpha", 1, 1, 0.5, 0, 0.25, 1),
		}
		mods[0].Zone = metrics.ZoneOfUselessness
		g := mustGraph(t, metrics.StatusPartial, mods, nil, warnings)
		out := RenderDetailed(g, RenderOptions{})

		wantLines := []string{
			"- global note",
			"- `pkg/a`: a note",
			"- `pkg/z`: z note",
			"- `pkg/omega` is in the Zone of Uselessness",
		}
		last := -1
		for _, want := range wantLines {
			if !strings.Contains(out, want+"\n") {
				t.Errorf("warnings section: output does not contain line %q", want)
				continue
			}
			if i := strings.Index(out, want+"\n"); i <= last {
				t.Errorf("warnings section order broken at %q (index %d, previous %d)", want, i, last)
			} else {
				last = i
			}
		}
	})

	t.Run("no warnings and no zone-of-uselessness packages", func(t *testing.T) {
		t.Parallel()
		out := RenderDetailed(mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil), RenderOptions{})
		if !strings.Contains(out, "## ⚠️ Warnings\n\nNo warnings.\n") {
			t.Errorf("warnings section: output does not contain the no-warnings line")
		}
	})
}

// TestRenderDetailed_Tiebreaks pins the canonical ordering for inputs where
// the primary sort key ties: duplicate module paths order by the rendered
// scalar fields, warnings with equal modules order by code then message,
// duplication rows with equal file and start line order by end line, and
// the grade-rank and overall-word fallbacks keep unknown grades worst and
// raw. These tie-breaks exist so no input permutation can change the
// rendered bytes; the classifiers never produce an unknown grade, so those
// fallbacks are pinned by direct call.
func TestRenderDetailed_Tiebreaks(t *testing.T) {
	t.Parallel()

	t.Run("duplicate module paths order by rendered fields", func(t *testing.T) {
		t.Parallel()
		lowCa := mod("pkg/dup", 1, 1, 0.5, 0, 0.25, 1)
		highCa := mod("pkg/dup", 2, 1, 0.5, 0, 0.25, 1)
		forward := mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{lowCa, highCa}, nil, nil)
		reverse := mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{highCa, lowCa}, nil, nil)
		if got, want := RenderDetailed(forward, RenderOptions{}), RenderDetailed(reverse, RenderOptions{}); got != want {
			t.Errorf("duplicate-path modules: input order changed the output:\nforward:\n%s\nreverse:\n%s", got, want)
		}
		out := RenderDetailed(forward, RenderOptions{})
		low := "| `pkg/dup` | 0.50 | 0.00 | 0.25 | 1 | 1 | 1 |"
		high := "| `pkg/dup` | 0.50 | 0.00 | 0.25 | 1 | 2 | 1 |"
		if l, h := strings.Index(out, low), strings.Index(out, high); l < 0 || h < 0 || l > h {
			t.Errorf("duplicate-path rows: low-Ca row at %d, high-Ca row at %d, want low first", l, h)
		}
	})

	t.Run("moduleLess tie-breaks on every rendered field", func(t *testing.T) {
		t.Parallel()
		base := mod("pkg/same", 1, 1, 0.5, 0.25, 0.25, 1)
		base.TotalLines = 100
		pathLower := base
		pathLower.Path = "pkg/a"
		pathHigher := base
		pathHigher.Path = "pkg/b"
		tests := []struct {
			name  string
			lower func(m metrics.ModuleResult) metrics.ModuleResult
		}{
			{"path", func(m metrics.ModuleResult) metrics.ModuleResult { return pathLower }},
			{"ca", func(m metrics.ModuleResult) metrics.ModuleResult { m.Ca--; return m }},
			{"ce", func(m metrics.ModuleResult) metrics.ModuleResult { m.Ce--; return m }},
			{"instability", func(m metrics.ModuleResult) metrics.ModuleResult { m.Instability -= 0.1; return m }},
			{"abstractness", func(m metrics.ModuleResult) metrics.ModuleResult { m.Abstractness -= 0.1; return m }},
			{"distance", func(m metrics.ModuleResult) metrics.ModuleResult { m.Distance -= 0.1; return m }},
			{"lcom", func(m metrics.ModuleResult) metrics.ModuleResult { m.LCOM--; return m }},
			{"zone", func(m metrics.ModuleResult) metrics.ModuleResult { m.Zone = metrics.ZoneMainSequence; return m }},
			{"total lines", func(m metrics.ModuleResult) metrics.ModuleResult { m.TotalLines--; return m }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				higher := pathHigher
				if tt.name != "path" {
					higher = base
				}
				if !moduleLess(tt.lower(base), higher) {
					t.Errorf("moduleLess(%s): lower field does not order first", tt.name)
				}
				if moduleLess(higher, tt.lower(base)) {
					t.Errorf("moduleLess(%s): ordering is symmetric, want strict", tt.name)
				}
			})
		}
	})

	t.Run("warnings with equal modules order by code then message", func(t *testing.T) {
		t.Parallel()
		warnings := []metrics.Warning{
			{Code: "b", Message: "second by code", Module: "pkg/a"},
			{Code: "a", Message: "first by code", Module: "pkg/a"},
			{Code: "same", Message: "zulu message", Module: "pkg/b"},
			{Code: "same", Message: "alpha message", Module: "pkg/b"},
		}
		g := mustGraph(t, metrics.StatusComplete, nil, nil, warnings)
		out := RenderDetailed(g, RenderOptions{})
		wantLines := []string{
			"- `pkg/a`: first by code",
			"- `pkg/a`: second by code",
			"- `pkg/b`: alpha message",
			"- `pkg/b`: zulu message",
		}
		last := -1
		for _, want := range wantLines {
			i := strings.Index(out, want+"\n")
			if i < 0 {
				t.Fatalf("warnings tiebreak: output does not contain line %q", want)
			}
			if i <= last {
				t.Errorf("warnings tiebreak order broken at %q (index %d, previous %d)", want, i, last)
			}
			last = i
		}
	})

	t.Run("duplication rows order by file, start line, then end line", func(t *testing.T) {
		t.Parallel()
		dups := []metrics.Duplication{{
			ModulePath: "pkg/dup",
			Blocks: []metrics.DuplicateBlock{
				dupBlock("x/a.go", 10, 30, 5),
				dupBlock("x/a.go", 10, 20, 5),
				dupBlock("x/a.go", 5, 9, 4),
			},
			Similarity: 0.9,
		}}
		g := mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{dupMod("pkg/dup", 100, dups...)}, nil, nil)
		out := RenderDetailed(g, RenderOptions{})
		wantRows := []string{
			"| `x/a.go:5-9` | 4 lines | 90% | 1 |",
			"| `x/a.go:10-20` | 5 lines | 90% | 1 |",
			"| `x/a.go:10-30` | 5 lines | 90% | 1 |",
		}
		last := -1
		for _, want := range wantRows {
			i := strings.Index(out, want)
			if i < 0 {
				t.Fatalf("duplication row tiebreak: output does not contain row %q", want)
			}
			if i <= last {
				t.Errorf("duplication row tiebreak order broken at %q (index %d, previous %d)", want, i, last)
			}
			last = i
		}
	})

	t.Run("unknown grade ranks worst and renders raw", func(t *testing.T) {
		t.Parallel()
		if got := gradeRank(Grade("X")); got != 6 {
			t.Errorf("gradeRank(X): got %v, want 6 (below F)", got)
		}
		if got := overallWord(Grade("X")); got != "X" {
			t.Errorf("overallWord(X): got %q, want %q", got, "X")
		}
	})
}

// TestRenderDetailed_ExactOutput pins the full byte format of the detailed
// report — every section heading, empty state, table separator, chain fence,
// and formatted value — for the empty module set and a rich graph. This is
// the format contract the CLI golden tests (task 3.3) and the regenerated
// example report (task 5.4) build on.
func TestRenderDetailed_ExactOutput(t *testing.T) {
	t.Parallel()

	legend := "📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md"

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		want string
	}{
		{
			name: "nil graph renders every section in its empty state",
			g:    nil,
			want: "# 🏗️ Architectural Health Report\n" +
				"\n" +
				"**Provenance**: vibe-check dev | 0 packages | 0 lines\n" +
				"\n" +
				legend + "\n" +
				"\n" +
				"All Clear | 0 packages | Avg Instability: 0.00 | Avg Distance: 0.00 | Max LCOM4: 0 | Cycles: 0\n" +
				"\n" +
				"## Prioritized Recommendations\n" +
				"\n" +
				"1. 🟢 Architecture is healthy — no significant issues detected; keep monitoring\n" +
				"\n" +
				"## 📊 Package Metrics\n" +
				"\n" +
				"No packages to analyze.\n" +
				"\n" +
				"## 🔗 Coupling Summary\n" +
				"\n" +
				"No packages to analyze.\n" +
				"\n" +
				"## 🧩 Cohesion Summary\n" +
				"\n" +
				"No packages to analyze.\n" +
				"\n" +
				"## 🔄 Circular Dependencies\n" +
				"\n" +
				"🟢 No circular dependencies detected.\n" +
				"\n" +
				"## 📋 Code Duplication\n" +
				"\n" +
				"No duplicated blocks detected.\n" +
				"\n" +
				"## 🏥 Health Scorecard\n" +
				"\n" +
				"| Dimension | Value | Grade | Status |\n" +
				"|-----------|-------|-------|--------|\n" +
				"| Avg Distance from Main Sequence | 0.00 | A | 🟢 |\n" +
				"| Max LCOM4 | 0 | A | 🟢 |\n" +
				"| Cycle Count | 0 | A | 🟢 |\n" +
				"| Duplication % | 0.0% | A | 🟢 |\n" +
				"| Instability Spread | 0 extremes | A | 🟢 |\n" +
				"\n" +
				"**Overall**: 🟢 Healthy\n" +
				"\n" +
				"## ⚠️ Warnings\n" +
				"\n" +
				"No warnings.\n",
		},
		{
			name: "rich graph renders every section populated",
			g:    detailedFixture(t),
			want: "# 🏗️ Architectural Health Report\n" +
				"\n" +
				"**Provenance**: vibe-check dev | 2 packages | 800 lines\n" +
				"\n" +
				legend + "\n" +
				"\n" +
				"Needs Attention | 2 packages | Avg Instability: 0.50 | Avg Distance: 0.32 | Max LCOM4: 2 | Cycles: 1\n" +
				"\n" +
				"## Prioritized Recommendations\n" +
				"\n" +
				"1. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing\n" +
				"2. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing\n" +
				"3. 🟢 Watch duplication in `x/a.go`, `x/b.go` — `2` blocks at `90`% similarity\n" +
				"\n" +
				"## 📊 Package Metrics\n" +
				"\n" +
				"| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce | Zone |\n" +
				"|---------|-------------|--------------|----------|-------|----|----|------|\n" +
				"| `pkg/alpha` | 0.06 | 0.50 | 0.44 | 1 | 12 | 0 | 🟢 Main Sequence |\n" +
				"| `pkg/omega` | 0.94 | 0.90 | 0.20 | 2 | 0 | 3 | 🔴 Zone of Uselessness |\n" +
				"\n" +
				"## 🔗 Coupling Summary\n" +
				"\n" +
				"- **Max Ca**: `pkg/alpha` (12 dependents)\n" +
				"- **Max Ce**: `pkg/omega` (3 dependencies)\n" +
				"- **Leaves (Ce=0)**: `pkg/alpha`\n" +
				"- **Heavily depended-on (Ca > 10)**: `pkg/alpha`\n" +
				"\n" +
				"## 🧩 Cohesion Summary\n" +
				"\n" +
				"- **Best**: `pkg/alpha` (LCOM4=1)\n" +
				"- **Worst**: `pkg/omega` (LCOM4=2)\n" +
				"- **Packages needing review (LCOM4 ≥ 4)**: none\n" +
				"\n" +
				"## 🔄 Circular Dependencies\n" +
				"\n" +
				"🔴 **1 cycle detected:**\n" +
				"\n" +
				"```\npkg/alpha → pkg/omega → pkg/alpha\n```\n" +
				"\n" +
				"## 📋 Code Duplication\n" +
				"\n" +
				"**2 duplicated blocks detected (overall 4.0%):**\n" +
				"\n" +
				"| File | Lines | Similarity | Occurrences |\n" +
				"|------|-------|------------|-------------|\n" +
				"| `x/a.go:10-25` | 16 lines | 90% | 1 |\n" +
				"| `x/b.go:40-55` | 16 lines | 90% | 1 |\n" +
				"\n" +
				"## 🏥 Health Scorecard\n" +
				"\n" +
				"| Dimension | Value | Grade | Status |\n" +
				"|-----------|-------|-------|--------|\n" +
				"| Avg Distance from Main Sequence | 0.32 | C+ | 🟡 |\n" +
				"| Max LCOM4 | 2 | B+ | 🟢 |\n" +
				"| Cycle Count | 1 | F | 🔴 |\n" +
				"| Duplication % | 4.0% | B | 🟡 |\n" +
				"| Instability Spread | 0 extremes | A | 🟢 |\n" +
				"\n" +
				"**Overall**: 🔴 Needs Attention\n" +
				"\n" +
				"## ⚠️ Warnings\n" +
				"\n" +
				"- `pkg/gone`: package pkg/gone: type information unavailable\n" +
				"- `pkg/omega` is in the Zone of Uselessness\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := RenderDetailed(tt.g, RenderOptions{}); got != tt.want {
				t.Errorf("RenderDetailed(%s):\ngot:\n%s\nwant:\n%s", tt.name, got, tt.want)
			}
		})
	}
}

// renderModes pairs each report mode's exported render entry point with its
// name, for the byte-stability tests that pin the same property on both.
var renderModes = []struct {
	name   string
	render func(*metrics.ModuleGraph, RenderOptions) string
}{
	{name: "summary", render: RenderSummary},
	{name: "detailed", render: RenderDetailed},
}

// graphPermutations lists the fixed array permutations the byte-stability
// tests apply to a graph: every renderer-consumed array is reordered and the
// rendered output must not change. The permutations are deterministic — no
// math/rand — so any failure reproduces exactly.
var graphPermutations = []struct {
	name string
	perm func(n, i int) int
}{
	{name: "reversed", perm: func(n, i int) int { return n - 1 - i }},
	{name: "rotated left by one", perm: func(n, i int) int { return (i + n - 1) % n }},
}

// permuteGraph returns a deep copy of g with every array the renderers
// consume reordered by perm, which maps each element's index in an n-element
// array to its index in the copy: the Modules, Cycles, and Warnings arrays,
// and within each module the Duplications array and each group's Blocks
// array. Cycle member sequences are deliberately not permuted — the [metrics.Cycle]
// contract requires lexicographically sorted members, so permuting members
// is contract-invalid input, not a different array order. The copy shares
// no slice or pointer with g. g must be non-nil: the permutation tests have
// no nil case, because a nil graph has no arrays to reorder.
func permuteGraph(g *metrics.ModuleGraph, perm func(n, i int) int) *metrics.ModuleGraph {
	out := *g
	if g.Provenance != nil {
		provenance := *g.Provenance
		out.Provenance = &provenance
	}
	out.Modules = permuteSlice(g.Modules, perm)
	out.Cycles = permuteSlice(g.Cycles, perm)
	out.Warnings = permuteSlice(g.Warnings, perm)
	for i := range out.Modules {
		out.Modules[i].Duplications = permuteSlice(out.Modules[i].Duplications, perm)
		for j := range out.Modules[i].Duplications {
			out.Modules[i].Duplications[j].Blocks = permuteSlice(out.Modules[i].Duplications[j].Blocks, perm)
		}
	}
	return &out
}

// permuteSlice returns a copy of s with each element moved from its current
// index to perm(len(s), index).
func permuteSlice[T any](s []T, perm func(n, i int) int) []T {
	out := make([]T, len(s))
	for i, v := range s {
		out[perm(len(s), i)] = v
	}
	return out
}

// stabilityFixture builds the rich fixture the byte-stability tests permute:
// five modules fed in scrambled path order, covering every renderer input
// surface — all four zones, two cycles of different lengths, module-scoped
// and global warnings in unsorted order, duplication groups on two modules
// sharing one block identity with conflicting line counts (max-wins), an
// instability extreme, and Ca and LCOM ties for the max/best scans — plus a
// provenance object whose generatedAt MUST NOT leak into the render. The
// fixture fires more recommendation hits than [RecommendationCap], so the
// surviving top five must themselves be permutation-stable.
func stabilityFixture(t *testing.T) *metrics.ModuleGraph {
	t.Helper()
	api := mod("pkg/api", 12, 25, 0.80, 0.10, 0.30, 2)
	api.TotalLines = 300
	core := mod("pkg/core", 12, 0, 0.04, 0.40, 0.20, 1)
	core.Zone = metrics.ZoneMainSequence
	core.TotalLines = 400
	legacy := mod("pkg/legacy", 3, 5, 0.60, 0.00, 0.55, 6)
	legacy.Zone = metrics.ZoneOfPain
	legacy.TotalLines = 250
	ui := mod("pkg/ui", 2, 21, 0.50, 0.20, 0.25, 1)
	ui.TotalLines = 200
	ui.Duplications = []metrics.Duplication{{
		ModulePath: "pkg/ui",
		Blocks: []metrics.DuplicateBlock{
			dupBlock("x/ui-a.go", 10, 25, 16),
			dupBlock("x/ui-b.go", 40, 55, 16),
		},
		Similarity: 0.85,
	}}
	util := mod("pkg/util", 0, 2, 1.00, 0.90, 0.20, 2)
	util.Zone = metrics.ZoneOfUselessness
	util.TotalLines = 150
	util.Duplications = []metrics.Duplication{{
		ModulePath: "pkg/util",
		Blocks: []metrics.DuplicateBlock{
			dupBlock("x/util.go", 1, 20, 12),
			// The same block identity pkg/ui reports with 16 lines: the
			// duplication table deduplicates and max-wins line counts, so
			// every ordering of the two reports must agree.
			dupBlock("x/ui-a.go", 10, 25, 20),
		},
		Similarity: 0.92,
	}}
	warnings := []metrics.Warning{
		{Code: "load-error", Message: "package pkg/gone: type information unavailable", Module: "pkg/gone"},
		{Code: "partial-build", Message: "analysis partially loaded; metrics may be incomplete"},
		{Code: "type-error", Message: "package pkg/legacy: stale type information", Module: "pkg/legacy"},
	}
	cycles := []metrics.Cycle{
		{"pkg/legacy", "pkg/ui", "pkg/util"},
		{"pkg/api", "pkg/core"},
	}
	g := mustGraph(t, metrics.StatusPartial,
		[]metrics.ModuleResult{util, legacy, api, ui, core}, // scrambled path order on purpose
		cycles, warnings)
	return withProvenance(t, g, "vibe-check", "1.2.3", "2026-01-15T10:30:00Z")
}

// recommendationTieFixture builds the recommendation-ordering fixture: five
// modules whose fired recommendations tie on extremity within each severity
// — the two cycle members tie at one participating cycle (red), the two
// over-coupled modules tie at Ce 21 (yellow), and the green duplication
// watch fills the fifth and final slot exactly at [RecommendationCap] — so
// only the engine's severity → extremity → path ordering can fix the
// rendered order. Modules are fed in reverse path order on purpose.
func recommendationTieFixture(t *testing.T) *metrics.ModuleGraph {
	t.Helper()
	cycleA := mod("pkg/cyc-a", 1, 1, 0.50, 0.10, 0.25, 1)
	cycleA.TotalLines = 200
	cycleB := mod("pkg/cyc-b", 1, 1, 0.50, 0.10, 0.25, 1)
	cycleB.TotalLines = 200
	hub := mod("pkg/hub", 1, 21, 0.50, 0.10, 0.25, 1)
	hub.TotalLines = 200
	zeta := mod("pkg/zeta", 1, 21, 0.50, 0.10, 0.25, 1)
	zeta.TotalLines = 200
	// 32 deduplicated lines over 1000 total lines is 3.2%: strictly between
	// 0% and 5%, so the green watch rule fires.
	dup := dupMod("pkg/dup", 200, metrics.Duplication{
		ModulePath: "pkg/dup",
		Blocks: []metrics.DuplicateBlock{
			dupBlock("x/a.go", 10, 25, 16),
			dupBlock("x/b.go", 40, 55, 16),
		},
		Similarity: 0.9,
	})
	return mustGraph(t, metrics.StatusComplete,
		[]metrics.ModuleResult{dup, zeta, hub, cycleB, cycleA}, // reverse path order on purpose
		[]metrics.Cycle{{"pkg/cyc-a", "pkg/cyc-b"}},
		nil)
}

// TestRender_ByteStableRepeatedRuns pins the byte-stability requirement's
// repeated-runs scenario at the renderer level: rendering the same graph
// twice — separated by a real time advance, because the report carries no
// generation timestamp, so a later wall clock cannot change a byte —
// produces identical output in both modes.
func TestRender_ByteStableRepeatedRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
		opts RenderOptions
	}{
		{name: "rich graph", g: stabilityFixture(t), opts: RenderOptions{}},
		{name: "recommendation ties", g: recommendationTieFixture(t), opts: RenderOptions{}},
		{name: "clean graph", g: mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil), opts: RenderOptions{}},
		{name: "violations graph", g: mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil), opts: RenderOptions{}},
		{name: "zero modules", g: mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{}, nil, nil), opts: RenderOptions{}},
		{name: "nil graph", g: nil, opts: RenderOptions{FallbackProducer: "x-renderer", FallbackVersion: "0.9.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range renderModes {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					first := mode.render(tt.g, tt.opts)
					time.Sleep(10 * time.Millisecond)
					second := mode.render(tt.g, tt.opts)
					if first != second {
						t.Errorf("%s mode: repeated runs of the same graph differ:\nfirst:\n%s\nsecond:\n%s", mode.name, first, second)
					}
				})
			}
		})
	}
}

// TestRender_ByteStableInputOrder pins the byte-stability requirement's
// input-order-independence scenario at the renderer level: a graph whose
// Modules, Cycles, Warnings, Duplications, and duplication Blocks arrays are
// reordered renders byte-identically in both modes. Modules render in path
// order regardless of input order (the fixture feeds them scrambled), and
// cycles, warnings, duplication rows, and recommendations render in their
// canonical orders, so no array permutation can change the output.
func TestRender_ByteStableInputOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
	}{
		{name: "rich graph", g: stabilityFixture(t)},
		{name: "clean graph", g: mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil)},
		{name: "violations graph", g: mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range renderModes {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()
					want := mode.render(tt.g, RenderOptions{})
					for _, p := range graphPermutations {
						if got := mode.render(permuteGraph(tt.g, p.perm), RenderOptions{}); got != want {
							t.Errorf("%s mode: %s arrays changed the output:\nwant:\n%s\ngot:\n%s", mode.name, p.name, want, got)
						}
					}
				})
			}
		})
	}
}

// TestRender_ByteStableRecommendationOrder pins the recommendation-ordering
// contract under input permutation: severity first (red before yellow
// before green), then rule-specific extremity descending, then path
// ascending as the tiebreak. The fixture ties extremities within the red
// and yellow tiers and fills the five-entry cap exactly, so the rendered
// order depends entirely on that comparator — and every array permutation
// must render the same five lines in the same order in both modes.
func TestRender_ByteStableRecommendationOrder(t *testing.T) {
	t.Parallel()

	g := recommendationTieFixture(t)

	// The engine's order pinned on the base render first, so a permutation
	// failure reads as a broken order rather than an opaque byte diff.
	wantLines := []string{
		"1. 🔴 Break cycle between `pkg/cyc-a`, `pkg/cyc-b` — circular dependencies prevent independent testing",
		"2. 🔴 Break cycle between `pkg/cyc-a`, `pkg/cyc-b` — circular dependencies prevent independent testing",
		"3. 🟡 Decouple `pkg/hub` — `21` efferent couplings; consider interface extraction",
		"4. 🟡 Decouple `pkg/zeta` — `21` efferent couplings; consider interface extraction",
		"5. 🟢 Watch duplication in `x/a.go`, `x/b.go` — `2` blocks at `90`% similarity",
	}
	out := RenderSummary(g, RenderOptions{})
	last := -1
	for _, want := range wantLines {
		i := strings.Index(out, want+"\n")
		if i < 0 {
			t.Fatalf("recommendation order: output does not contain line %q:\n%s", want, out)
		}
		if i <= last {
			t.Errorf("recommendation order broken at %q (index %d, previous %d)", want, i, last)
		}
		last = i
	}

	for _, mode := range renderModes {
		want := mode.render(g, RenderOptions{})
		for _, p := range graphPermutations {
			if got := mode.render(permuteGraph(g, p.perm), RenderOptions{}); got != want {
				t.Errorf("%s mode: %s arrays changed the output:\nwant:\n%s\ngot:\n%s", mode.name, p.name, want, got)
			}
		}
	}
}

// TestRenderDetailed_NoTimestamp pins the byte-stability rule that no
// generation timestamp appears anywhere in the detailed output either: the
// snapshot's provenance carries one, no RFC3339-shaped timestamp appears
// anywhere in the report, and the header line carries exactly the producer,
// version, package count, and total lines — no date or time field. The
// exact-output goldens pin the full report bytes; this test pins the
// detailed-mode header field list.
func TestRenderDetailed_NoTimestamp(t *testing.T) {
	t.Parallel()

	out := RenderDetailed(stabilityFixture(t), RenderOptions{})

	if strings.Contains(out, "2026-01-15T10:30:00Z") {
		t.Errorf("detailed output contains the snapshot's generatedAt timestamp")
	}
	rfc3339 := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	if match := rfc3339.FindString(out); match != "" {
		t.Errorf("detailed output contains an RFC3339-shaped timestamp %q", match)
	}
	wantHeader := "**Provenance**: vibe-check 1.2.3 | 5 packages | 1300 lines"
	// The report opens with the title heading, a blank line, then the
	// header line — so the header is the third line, and its exact bytes
	// pin the field list: producer, version, package count, total lines,
	// and nothing else.
	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("detailed output has no header line:\n%s", out)
	}
	if lines[2] != wantHeader {
		t.Errorf("header line: got %q, want %q", lines[2], wantHeader)
	}
}
