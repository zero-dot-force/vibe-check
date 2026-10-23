package report

import (
	"strings"
	"testing"
	"time"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// --- Diff markdown rendering ---------------------------------------------------

// declinedDiffFixture returns a base/PR pair where ex/a crosses all three
// metric gates (ΔInstability +0.30, ΔDistance +0.25, ΔLCOM +2) and a new
// ex/b→ex/c cycle appears, plus one removed and one added package:
// REQUEST_CHANGES, degrading, four reasons, one new cycle. It mirrors the
// committed CLI fixture pair cmd/vibe-check/testdata/diff/declined-{base,pr}.json
// so this engine exact-output test and the CLI golden pin the same graphs.
func declinedDiffFixture(t *testing.T) (base, pr *metrics.ModuleGraph) {
	t.Helper()
	base = mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{
		mod("ex/a", 1, 1, 0.30, 0.10, 0.20, 1),
		mod("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
		mod("ex/c", 0, 1, 0.50, 0.00, 0.30, 1),
		mod("ex/d", 1, 0, 0.40, 0.10, 0.30, 1),
	}, nil, nil)
	pr = mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{
		mod("ex/a", 1, 1, 0.60, 0.10, 0.45, 3),
		mod("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
		mod("ex/c", 0, 1, 0.50, 0.00, 0.30, 1),
		mod("ex/e", 1, 1, 0.50, 0.10, 0.40, 1),
	}, []metrics.Cycle{{"ex/c", "ex/b"}}, nil)
	return base, pr
}

// improvedDiffFixture returns a base/PR pair where ex/a improves on every
// gated metric and the base cycle is resolved: APPROVE, improving, no
// reasons, one resolved cycle. It mirrors the committed CLI fixture pair
// cmd/vibe-check/testdata/diff/improved-{base,pr}.json.
func improvedDiffFixture(t *testing.T) (base, pr *metrics.ModuleGraph) {
	t.Helper()
	base = mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{
		mod("ex/a", 1, 2, 0.50, 0.10, 0.40, 3),
		mod("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
	}, []metrics.Cycle{{"ex/b", "ex/a"}}, nil)
	pr = mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{
		mod("ex/a", 1, 1, 0.30, 0.10, 0.20, 1),
		mod("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
	}, nil, nil)
	return base, pr
}

// mixedDiffFixture returns a base/PR pair where ex/a instability rises by
// 0.05 — a material shift below every REQUEST_CHANGES gate: COMMENT,
// stable, the single materiality reason. It mirrors the committed CLI
// fixture pair cmd/vibe-check/testdata/diff/mixed-{base,pr}.json.
func mixedDiffFixture(t *testing.T) (base, pr *metrics.ModuleGraph) {
	t.Helper()
	base = mustGraph(t, metrics.StatusComplete,
		[]metrics.ModuleResult{mod("ex/a", 1, 1, 0.50, 0.10, 0.30, 1)}, nil, nil)
	pr = mustGraph(t, metrics.StatusComplete,
		[]metrics.ModuleResult{mod("ex/a", 1, 1, 0.55, 0.10, 0.30, 1)}, nil, nil)
	return base, pr
}

// TestRenderDiffMarkdown_ExactOutput pins the diff markdown renderer's exact
// bytes for every verdict scenario: the title, the verdict/direction line,
// the reasons section (verbatim DecideVerdict strings or the none line), the
// per-module delta table in path order with the 4-decimal ratio vocabulary,
// and the new/resolved cycle sections with closed chains. The first three
// rows mirror the committed CLI fixtures, so this test and the CLI goldens
// cross-validate the same graphs through both surfaces.
func TestRenderDiffMarkdown_ExactOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t *testing.T) (base, pr *metrics.ModuleGraph)
		want  string
	}{
		{
			name:  "request_changes",
			build: declinedDiffFixture,
			want: "# 🏗️ Structural Entropy Diff\n\n" +
				"**Verdict**: REQUEST_CHANGES | **Entropy direction**: degrading\n\n" +
				"## ⚠️ Verdict Reasons\n\n" +
				"- new-cycle: ex/b ex/c\n" +
				"- instability: module \"ex/a\" delta 0.3000 >= threshold 0.1500\n" +
				"- distance: module \"ex/a\" delta 0.2500 >= threshold 0.2000\n" +
				"- lcom: module \"ex/a\" delta 2 >= threshold 2\n" +
				"\n" +
				"## 📊 Per-module Deltas (PR minus base)\n\n" +
				"| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce |\n" +
				"|---------|-------------|--------------|----------|-------|----|----|\n" +
				"| `ex/a` | 0.3000 | 0.0000 | 0.2500 | 2 | 0 | 0 |\n" +
				"| `ex/b` | 0.0000 | 0.0000 | 0.0000 | 0 | 0 | 0 |\n" +
				"| `ex/c` | 0.0000 | 0.0000 | 0.0000 | 0 | 0 | 0 |\n" +
				"\n" +
				"## 🔄 New Cycles\n\n" +
				"🔴 **1 cycle detected:**\n\n" +
				"```\n" +
				"ex/b → ex/c → ex/b\n" +
				"```\n" +
				"\n" +
				"## 🔄 Resolved Cycles\n\n" +
				"None.\n",
		},
		{
			name:  "approve_improving",
			build: improvedDiffFixture,
			want: "# 🏗️ Structural Entropy Diff\n\n" +
				"**Verdict**: APPROVE | **Entropy direction**: improving\n\n" +
				"## ⚠️ Verdict Reasons\n\n" +
				"None.\n" +
				"\n" +
				"## 📊 Per-module Deltas (PR minus base)\n\n" +
				"| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce |\n" +
				"|---------|-------------|--------------|----------|-------|----|----|\n" +
				"| `ex/a` | -0.2000 | 0.0000 | -0.2000 | -2 | 0 | -1 |\n" +
				"| `ex/b` | 0.0000 | 0.0000 | 0.0000 | 0 | 0 | 0 |\n" +
				"\n" +
				"## 🔄 New Cycles\n\n" +
				"🟢 No new circular dependencies.\n" +
				"\n" +
				"## 🔄 Resolved Cycles\n\n" +
				"🟢 **1 cycle resolved:**\n\n" +
				"```\n" +
				"ex/a → ex/b → ex/a\n" +
				"```\n",
		},
		{
			name:  "comment_stable",
			build: mixedDiffFixture,
			want: "# 🏗️ Structural Entropy Diff\n\n" +
				"**Verdict**: COMMENT | **Entropy direction**: stable\n\n" +
				"## ⚠️ Verdict Reasons\n\n" +
				"- materiality: non-zero metric shifts below REQUEST_CHANGES thresholds\n" +
				"\n" +
				"## 📊 Per-module Deltas (PR minus base)\n\n" +
				"| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce |\n" +
				"|---------|-------------|--------------|----------|-------|----|----|\n" +
				"| `ex/a` | 0.0500 | 0.0000 | 0.0000 | 0 | 0 | 0 |\n" +
				"\n" +
				"## 🔄 New Cycles\n\n" +
				"🟢 No new circular dependencies.\n" +
				"\n" +
				"## 🔄 Resolved Cycles\n\n" +
				"None.\n",
		},
		{
			name: "approve_no_shared_modules",
			build: func(t *testing.T) (base, pr *metrics.ModuleGraph) {
				t.Helper()
				return completeGraph(t, mod("ex/a", 1, 1, 0.50, 0.10, 0.40, 1)),
					completeGraph(t, mod("ex/b", 1, 1, 0.50, 0.10, 0.40, 1))
			},
			want: "# 🏗️ Structural Entropy Diff\n\n" +
				"**Verdict**: APPROVE | **Entropy direction**: stable\n\n" +
				"## ⚠️ Verdict Reasons\n\n" +
				"None.\n" +
				"\n" +
				"## 📊 Per-module Deltas (PR minus base)\n\n" +
				"No shared packages.\n" +
				"\n" +
				"## 🔄 New Cycles\n\n" +
				"🟢 No new circular dependencies.\n" +
				"\n" +
				"## 🔄 Resolved Cycles\n\n" +
				"None.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base, pr := tt.build(t)
			delta := metrics.ComputeDelta(base, pr)
			verdict, reasons := metrics.DecideVerdict(delta, metrics.DefaultVerdictThresholds())
			if got := RenderDiffMarkdown(delta, verdict, reasons); got != tt.want {
				t.Errorf("RenderDiffMarkdown: got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// TestRenderDiffMarkdown_ByteStableRepeatedRenders pins the byte-stability
// requirement for the diff markdown: rendering the same delta twice — with
// real time advancing between the renders — produces byte-identical output,
// because the renderer derives every byte from the delta and never reads a
// clock.
func TestRenderDiffMarkdown_ByteStableRepeatedRenders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t *testing.T) (base, pr *metrics.ModuleGraph)
	}{
		{name: "declined", build: declinedDiffFixture},
		{name: "improved", build: improvedDiffFixture},
		{name: "mixed", build: mixedDiffFixture},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base, pr := tt.build(t)
			delta := metrics.ComputeDelta(base, pr)
			verdict, reasons := metrics.DecideVerdict(delta, metrics.DefaultVerdictThresholds())
			first := RenderDiffMarkdown(delta, verdict, reasons)
			time.Sleep(10 * time.Millisecond)
			second := RenderDiffMarkdown(delta, verdict, reasons)
			if first != second {
				t.Errorf("two renders of the same delta differ:\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

// TestRenderDiffMarkdown_InputOrderIndependent pins the renderer's defensive
// canonical ordering: permuting the delta's Modules slice and its NewCycles
// and ResolvedCycles lists — the arrays a hand-built GraphDelta could carry
// in any order — does not change a single byte. Cycle member sequences are
// deliberately NOT permuted: the Cycle contract requires sorted members, so
// permuting them is contract-invalid input.
func TestRenderDiffMarkdown_InputOrderIndependent(t *testing.T) {
	t.Parallel()

	delta := metrics.GraphDelta{
		Modules: []metrics.Delta{
			{Path: "ex/c", Instability: 0.01},
			{Path: "ex/a", Distance: -0.02, LCOM: 1},
			{Path: "ex/b", Ca: 2, Ce: -1},
		},
		NewCycles:      []metrics.Cycle{{"ex/a", "ex/b"}, {"ex/c", "ex/d"}},
		ResolvedCycles: []metrics.Cycle{{"ex/e", "ex/f"}, {"ex/g", "ex/h"}},
		Direction:      metrics.EntropyDegrading,
	}
	permuted := metrics.GraphDelta{
		Modules: []metrics.Delta{
			delta.Modules[2],
			delta.Modules[0],
			delta.Modules[1],
		},
		NewCycles:      []metrics.Cycle{delta.NewCycles[1], delta.NewCycles[0]},
		ResolvedCycles: []metrics.Cycle{delta.ResolvedCycles[1], delta.ResolvedCycles[0]},
		Direction:      metrics.EntropyDegrading,
	}

	got := RenderDiffMarkdown(delta, metrics.VerdictRequestChanges, nil)
	want := RenderDiffMarkdown(permuted, metrics.VerdictRequestChanges, nil)
	if got != want {
		t.Errorf("output depends on input slice order:\noriginal:\n%s\npermuted:\n%s", got, want)
	}

	// Non-vacuous setup and canonical order: both cycle sections carry two
	// chains, and the module rows render in path order.
	if !strings.Contains(got, "2 cycles detected") || !strings.Contains(got, "2 cycles resolved") {
		t.Errorf("fixture must carry two new and two resolved cycles for the test to be non-vacuous, got:\n%s", got)
	}
	at := strings.Index(got, "`ex/a`")
	for _, path := range []string{"`ex/b`", "`ex/c`"} {
		next := strings.Index(got, path)
		if next < 0 || next < at {
			t.Errorf("module row %s must follow the preceding row in path order (at %d, got %d):\n%s", path, at, next, got)
		}
		at = next
	}
}

// TestDeltaLess_TiebreaksOnEveryRenderedField pins the delta comparator's
// strict total order: path orders first, then every rendered scalar field
// as a tiebreak, mirroring the moduleLess tiebreak test over the report's
// canonical module order. Duplicate-path deltas are malformed input —
// metrics.ComputeDelta keys by path so it never emits them — but the
// comparator must still order them strictly so rendering cannot depend on
// the order the delta's slice carries.
func TestDeltaLess_TiebreaksOnEveryRenderedField(t *testing.T) {
	t.Parallel()

	base := metrics.Delta{Path: "pkg/same", Ca: 2, Ce: 2, Instability: 0.5, Abstractness: 0.25, Distance: 0.25, LCOM: 2}
	pathLower := base
	pathLower.Path = "pkg/a"
	pathHigher := base
	pathHigher.Path = "pkg/b"
	tests := []struct {
		name  string
		lower func(d metrics.Delta) metrics.Delta
	}{
		{"path", func(d metrics.Delta) metrics.Delta { return pathLower }},
		{"ca", func(d metrics.Delta) metrics.Delta { d.Ca--; return d }},
		{"ce", func(d metrics.Delta) metrics.Delta { d.Ce--; return d }},
		{"instability", func(d metrics.Delta) metrics.Delta { d.Instability -= 0.1; return d }},
		{"abstractness", func(d metrics.Delta) metrics.Delta { d.Abstractness -= 0.1; return d }},
		{"distance", func(d metrics.Delta) metrics.Delta { d.Distance -= 0.1; return d }},
		{"lcom", func(d metrics.Delta) metrics.Delta { d.LCOM--; return d }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			higher := pathHigher
			if tt.name != "path" {
				higher = base
			}
			if !deltaLess(tt.lower(base), higher) {
				t.Errorf("deltaLess(%s): lower field does not order first", tt.name)
			}
			if deltaLess(higher, tt.lower(base)) {
				t.Errorf("deltaLess(%s): ordering is symmetric, want strict", tt.name)
			}
		})
	}
}
