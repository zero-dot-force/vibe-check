package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Status summary words for the deterministic report. [ReportStatus] returns
// exactly one of these two values, and the summary-mode status line begins
// with the returned word.
const (
	// StatusAllClear is the overall status word when no violation-severity
	// (red or yellow) recommendation fires.
	StatusAllClear = "All Clear"
	// StatusNeedsAttention is the overall status word when at least one
	// violation-severity (red or yellow) recommendation fires.
	StatusNeedsAttention = "Needs Attention"
)

// Render-time provenance fallback defaults. The report package cannot know
// which binary is rendering a report, so the report header's fallback
// producer identity is injected by the caller via [RenderOptions]; these
// constants are the last-resort defaults when neither the snapshot's
// provenance nor the options supply a value.
const (
	// DefaultRenderProducer is the producer name rendered in the report
	// header when the snapshot omits provenance and [RenderOptions] supplies
	// no fallback producer. It matches the CLI's producer name constant.
	DefaultRenderProducer = "vibe-check"
	// DefaultRenderVersion is the version rendered in the report header when
	// the snapshot omits provenance and [RenderOptions] supplies no fallback
	// version. It matches the CLI's ldflags version default.
	DefaultRenderVersion = "dev"
)

// metricsGuideURL is the absolute URL to the metrics guide, linked from the
// fixed legend at the top of every report. An absolute URL is contract: the
// legend must link the guide regardless of where the report is rendered
// (CLI stdout, CI step summary, or agent output).
const metricsGuideURL = "https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md"

// reportTitle is the first heading of every report: the 🏗️ title line of
// the report's fixed emoji vocabulary.
const reportTitle = "# 🏗️ Architectural Health Report"

// reportLegend is the fixed one-line metric legend rendered under the report
// header. It is contract, carried over verbatim from the vibe-check-reporter
// agent's legend template: it defines all six metric terms (Instability,
// Abstractness, Distance, LCOM4, Ca, Ce) and links the metrics guide by
// absolute URL. Rendering MUST NOT rewrite it.
const reportLegend = "📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: " + metricsGuideURL

// RenderOptions supplies the report renderer's provenance fallback: the
// producer identity stamped into the report header when the snapshot omits
// provenance. The report package cannot know the rendering binary's name or
// embedded version, so callers inject them here — the CLI passes its producer
// name and its ldflags-embedded version, keeping the rendered fallback equal
// to what `vibe-check analyze` stamps into the snapshots it emits.
//
// The zero value is valid: empty fields fall back to [DefaultRenderProducer]
// and [DefaultRenderVersion]. RenderOptions is a value struct passed per
// call — no package-level mutable state — so rendering stays a pure,
// deterministic function of (graph, options).
type RenderOptions struct {
	// FallbackProducer is the producer name rendered in the report header
	// when the snapshot omits provenance (or carries an empty producer
	// field). Empty falls back to [DefaultRenderProducer].
	FallbackProducer string
	// FallbackVersion is the version rendered in the report header when the
	// snapshot omits provenance (or carries an empty version field). Empty
	// falls back to [DefaultRenderVersion].
	FallbackVersion string
}

// Emoji returns the status emoji for the severity: 🔴 for [SeverityRed], 🟡
// for [SeverityYellow], and 🟢 for [SeverityGreen], per the report's fixed
// emoji vocabulary. Rendering maps the machine-readable severity strings onto
// these emoji; the JSON model carries the strings, markdown carries the emoji.
//
// An unrecognized severity returns the empty string. The rule engine only
// ever produces the three defined severities, so an empty result indicates a
// programming error elsewhere, not a reportable condition.
func (s RecommendationSeverity) Emoji() string {
	switch s {
	case SeverityRed:
		return emojiRed
	case SeverityYellow:
		return emojiYellow
	case SeverityGreen:
		return emojiGreen
	default:
		return ""
	}
}

// HasViolations reports whether recs contains at least one
// violation-severity recommendation — red or yellow. This predicate drives
// both the report's status word ([ReportStatus]) and the report command's
// exit-code contract: exit 1 when any red or yellow recommendation fires,
// exit 0 otherwise. Green-only recommendations (watch rule, all-clear) are
// informational and never count as violations.
func HasViolations(recs []Recommendation) bool {
	for _, r := range recs {
		if r.Severity == SeverityRed || r.Severity == SeverityYellow {
			return true
		}
	}
	return false
}

// ReportStatus returns the overall status word for the fired
// recommendations: [StatusNeedsAttention] when any red or yellow
// recommendation fires, [StatusAllClear] otherwise (including an empty
// recommendation list). It is a pure function of the recommendation list, so
// callers that already computed recommendations — the renderers, the report
// command's exit path, the JSON model — derive the status without
// recomputing the rule engine.
func ReportStatus(recs []Recommendation) string { //nolint:revive // stutter accepted: name kept stable across the metrics-to-report move so callers change only their import path
	if HasViolations(recs) {
		return StatusNeedsAttention
	}
	return StatusAllClear
}

// RenderSummary renders the summary-mode markdown report for a ModuleGraph:
// the 🏗️ report header (title, producer name and version, package count,
// total analyzed lines — no timestamp), the 📖 fixed metric legend, the
// status summary line, and the prioritized recommendations. It is a PURE,
// DETERMINISTIC function of (graph, options): built with strings.Builder per
// design D6, no text/template, no I/O, no timestamps, and every value derived
// deterministically from the graph, so identical input renders
// byte-identical output.
//
// The producer name and version come from the snapshot's provenance when
// present; when the snapshot omits provenance, they fall back to the
// opts-injected rendering-binary identity (see [RenderOptions]). The
// snapshot's provenance.GeneratedAt is never rendered.
//
// A nil graph renders as the empty module-set report — "All Clear" over zero
// packages — mirroring the nil handling of [ComputeAggregates] and
// [ComputeRecommendations]. Rendering cannot fail, so no error is returned.
//
// The summary sections are written by unexported section helpers
// (writeReportHeader, writeLegend, writeStatusSummary, writeRecommendations)
// that the detailed-mode renderer composes alongside its own sections.
func RenderSummary(g *metrics.ModuleGraph, opts RenderOptions) string {
	recs := ComputeRecommendations(g)
	var b strings.Builder
	writeReportHeader(&b, g, opts)
	writeLegend(&b)
	writeStatusSummary(&b, g, recs)
	writeRecommendations(&b, recs)
	return finalizeReport(&b)
}

// writeReportHeader writes the report's opening section: the 🏗️ title
// heading and the header line carrying the producer name and version, the
// package count, and the total analyzed lines. Per the report spec's
// byte-stability requirement, no generation timestamp is rendered — the
// snapshot's provenance.GeneratedAt is deliberately dropped.
func writeReportHeader(b *strings.Builder, g *metrics.ModuleGraph, opts RenderOptions) {
	producer, version := reportProducer(g, opts)
	packages, totalLines := graphStats(g)
	b.WriteString(reportTitle)
	b.WriteString("\n\n")
	fmt.Fprintf(b, "**Provenance**: %s %s | %s | %s\n\n",
		producer, version, countWord(packages, "package"), countWord(totalLines, "line"))
}

// writeLegend writes the fixed one-line metric legend. The legend text is
// contract (see [reportLegend]); this helper exists so every report mode
// renders it identically.
func writeLegend(b *strings.Builder) {
	b.WriteString(reportLegend)
	b.WriteString("\n\n")
}

// writeStatusSummary writes the status summary line: the overall status word
// followed by the package count, average instability, average distance, max
// LCOM4, and cycle count. The line begins with the status word itself — no
// emoji prefix — per the report spec's status-summary scenario. Ratio metrics
// render with %.2f and counts as integers, per the report's fixed formatting
// vocabulary.
func writeStatusSummary(b *strings.Builder, g *metrics.ModuleGraph, recs []Recommendation) {
	agg := ComputeAggregates(g)
	packages, _ := graphStats(g)
	fmt.Fprintf(b, "%s | %s | Avg Instability: %.2f | Avg Distance: %.2f | Max LCOM4: %d | Cycles: %d\n\n",
		ReportStatus(recs), countWord(packages, "package"),
		agg.AverageInstability, agg.AverageDistance, agg.MaxLCOM4, agg.CycleCount)
}

// writeRecommendations writes the prioritized recommendations section: the
// fired recommendations of the rule engine, already ordered and capped by
// [ComputeRecommendations], rendered as a numbered list with the severity
// emoji. Messages are contract and are rendered verbatim — the renderer adds
// only the list numbering, the emoji, and the separating space. recs is never
// empty because the rule engine falls back to the single green all-clear.
func writeRecommendations(b *strings.Builder, recs []Recommendation) {
	b.WriteString("## Prioritized Recommendations\n\n")
	for i, r := range recs {
		fmt.Fprintf(b, "%d. %s %s\n", i+1, r.Severity.Emoji(), r.Message)
	}
	b.WriteString("\n")
}

// finalizeReport returns the built report with trailing blank lines trimmed
// and exactly one final newline appended. Every section writer terminates its
// section with a blank separator line, so composition is order-free; this
// final pass keeps the report's last byte well-formed no matter which
// sections were written.
func finalizeReport(b *strings.Builder) string {
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// RenderDetailed renders the detailed-mode markdown report for a ModuleGraph:
// the complete summary-mode report — header, legend, status summary line, and
// prioritized recommendations, written by the same section writers
// [RenderSummary] uses so the detailed output begins with the summary output
// byte-for-byte — followed by seven additional sections in fixed order:
//
//   - 📊 Package Metrics: one row per module in path order, with full-word
//     column headers and an emoji-labeled zone column.
//   - 🔗 Coupling Summary: the max-Ca and max-Ce modules, the leaf modules
//     (Ce == 0), and the heavily-depended-on modules (Ca > 10).
//   - 🧩 Cohesion Summary: the best and worst LCOM4 modules, plus the
//     modules needing review (LCOM4 >= 4).
//   - 🔄 Circular Dependencies: every cycle as one closed chain per line,
//     members joined with →.
//   - 📋 Code Duplication: one row per deduplicated duplication block.
//   - 🏥 Health Scorecard: the five graded dimensions and the overall
//     status derived from the worst grade.
//   - ⚠️ Warnings: the snapshot's analysis warnings and the packages
//     classified in the Zone of Uselessness.
//
// It is a PURE, DETERMINISTIC function of (graph, options): built with
// strings.Builder per design D6, no text/template, no I/O, no timestamps.
// Modules render in path order regardless of the Modules array's order, and
// cycles, warnings, and duplication blocks render in canonical order too,
// so identical input renders byte-identical output.
//
// A nil graph renders as the empty module-set report: the summary mode's
// all-clear path followed by every detailed section in its empty state —
// the scorecard grades every dimension A over zero values. Rendering cannot
// fail, so no error is returned.
func RenderDetailed(g *metrics.ModuleGraph, opts RenderOptions) string {
	recs := ComputeRecommendations(g)
	var b strings.Builder
	writeReportHeader(&b, g, opts)
	writeLegend(&b)
	writeStatusSummary(&b, g, recs)
	writeRecommendations(&b, recs)
	writePackageMetrics(&b, g)
	writeCouplingSummary(&b, g)
	writeCohesionSummary(&b, g)
	writeCyclesSection(&b, g)
	writeDuplicationTable(&b, g)
	writeScorecard(&b, g)
	writeWarningsSection(&b, g)
	return finalizeReport(&b)
}

// heavilyDependedCaMin is the EXCLUSIVE lower bound on afferent coupling for
// the coupling summary's heavily-depended-on list: a module counts as
// heavily depended on when its Ca is strictly greater than 10, so a module
// at exactly 10 does not. The bound mirrors the example report's "Ca > 10"
// list and is a rendering choice, deliberately separate from the rule
// table's [RuleCeMin] even though both are exclusive bounds. Unit:
// non-negative integer (module dependents).
const heavilyDependedCaMin = 10

// writePackageMetrics writes the 📊 package metrics table: one row per module
// in the report's canonical path order, with full-word column headers and a
// zone column carrying the emoji-prefixed human label (see [zoneLabel]).
// Ratio metrics render with %.2f and counts as integers, per the report's
// fixed formatting vocabulary. A graph with no modules renders the heading
// with a no-packages note instead of a header-only table.
func writePackageMetrics(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## 📊 Package Metrics\n\n")
	mods := sortedModules(g)
	if len(mods) == 0 {
		b.WriteString("No packages to analyze.\n\n")
		return
	}
	writeTableHeader(b, "Package", "Instability", "Abstractness", "Distance", "LCOM4", "Ca", "Ce", "Zone")
	for _, m := range mods {
		fmt.Fprintf(b, "| `%s` | %.2f | %.2f | %.2f | %d | %d | %d | %s |\n",
			m.Path, m.Instability, m.Abstractness, m.Distance, m.LCOM, m.Ca, m.Ce, zoneLabel(m.Zone))
	}
	b.WriteString("\n")
}

// writeCouplingSummary writes the 🔗 coupling summary: the modules carrying
// the maximum afferent and efferent coupling, the leaf modules (Ce == 0,
// depending on nothing), and the heavily-depended-on modules (Ca strictly
// greater than [heavilyDependedCaMin]). Max Ca and max Ce each name the
// first module in path order among those tied at the maximum — the full
// per-module values already sit in the package metrics table above. Empty
// lists render "none"; a graph with no modules renders the heading with a
// no-packages note.
func writeCouplingSummary(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## 🔗 Coupling Summary\n\n")
	mods := sortedModules(g)
	if len(mods) == 0 {
		b.WriteString("No packages to analyze.\n\n")
		return
	}
	maxCa, maxCe := mods[0], mods[0]
	for _, m := range mods[1:] {
		if m.Ca > maxCa.Ca {
			maxCa = m
		}
		if m.Ce > maxCe.Ce {
			maxCe = m
		}
	}
	var leaves, heavy []string
	for _, m := range mods {
		if m.Ce == 0 {
			leaves = append(leaves, m.Path)
		}
		if m.Ca > heavilyDependedCaMin {
			heavy = append(heavy, m.Path)
		}
	}
	fmt.Fprintf(b, "- **Max Ca**: `%s` (%d dependents)\n", maxCa.Path, maxCa.Ca)
	fmt.Fprintf(b, "- **Max Ce**: `%s` (%d dependencies)\n", maxCe.Path, maxCe.Ce)
	fmt.Fprintf(b, "- **Leaves (Ce=0)**: %s\n", backtickListOrNone(leaves))
	fmt.Fprintf(b, "- **Heavily depended-on (Ca > %d)**: %s\n", heavilyDependedCaMin, backtickListOrNone(heavy))
	b.WriteString("\n")
}

// writeCohesionSummary writes the 🧩 cohesion summary: the best module
// (lowest LCOM4) and the worst module (highest LCOM4) — each the first
// module in path order among ties — and the modules needing review (LCOM4
// at least [RuleLCOM4Min], the same split threshold the recommendation
// engine applies), listed with their values in path order. Empty lists
// render "none"; a graph with no modules renders the heading with a
// no-packages note.
func writeCohesionSummary(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## 🧩 Cohesion Summary\n\n")
	mods := sortedModules(g)
	if len(mods) == 0 {
		b.WriteString("No packages to analyze.\n\n")
		return
	}
	best, worst := mods[0], mods[0]
	for _, m := range mods[1:] {
		if m.LCOM < best.LCOM {
			best = m
		}
		if m.LCOM > worst.LCOM {
			worst = m
		}
	}
	var review []string
	for _, m := range mods {
		if m.LCOM >= RuleLCOM4Min {
			review = append(review, fmt.Sprintf("`%s` (%d)", m.Path, m.LCOM))
		}
	}
	fmt.Fprintf(b, "- **Best**: `%s` (LCOM4=%d)\n", best.Path, best.LCOM)
	fmt.Fprintf(b, "- **Worst**: `%s` (LCOM4=%d)\n", worst.Path, worst.LCOM)
	fmt.Fprintf(b, "- **Packages needing review (LCOM4 ≥ %d)**: %s\n", RuleLCOM4Min, orNone(review))
	b.WriteString("\n")
}

// writeCyclesSection writes the 🔄 circular dependency chains: every cycle
// rendered as one chain per line — members joined with "→" and the loop
// closed by repeating the first member — inside a fenced code block, chains
// ordered by member sequence (see [sortedCycles]). A graph with no
// renderable cycles — including the degenerate zero-member cycles the
// schema permits but that name no members — renders the green no-cycles
// line instead of the block.
func writeCyclesSection(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## 🔄 Circular Dependencies\n\n")
	cycles := sortedCycles(g)
	if len(cycles) == 0 {
		b.WriteString("🟢 No circular dependencies detected.\n\n")
		return
	}
	fmt.Fprintf(b, "🔴 **%s detected:**\n\n", countWord(len(cycles), "cycle"))
	b.WriteString("```\n")
	for _, c := range cycles {
		b.WriteString(cycleChain(c))
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
}

// writeDuplicationTable writes the 📋 duplication table: one row per
// deduplicated duplication block — identity is file plus line range, the
// same [dupBlockKey] semantics [ComputeAggregates] applies — carrying the
// block's significant line count (largest reported on conflicting counts,
// matching [duplicationPercent]), its highest referencing group's
// similarity, and its occurrence count across duplication groups. Rows
// order by file, then line range. The lead line carries the deduplicated
// block count and the overall duplication percentage from
// [ComputeAggregates], tying the table to the scorecard's duplication
// dimension. A graph with no duplicated blocks renders the no-duplication
// line instead of the table.
func writeDuplicationTable(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## 📋 Code Duplication\n\n")
	var mods []metrics.ModuleResult
	if g != nil {
		mods = g.Modules
	}
	rows := duplicationRows(mods)
	if len(rows) == 0 {
		b.WriteString("No duplicated blocks detected.\n\n")
		return
	}
	fmt.Fprintf(b, "**%s detected (overall %.1f%%):**\n\n",
		countWord(len(rows), "duplicated block"), ComputeAggregates(g).DuplicationPercent)
	writeTableHeader(b, "File", "Lines", "Similarity", "Occurrences")
	for _, r := range rows {
		fmt.Fprintf(b, "| `%s:%d-%d` | %s | %.0f%% | %d |\n",
			r.file, r.startLine, r.endLine, countWord(r.lines, "line"), r.similarity*100, r.occurrences)
	}
	b.WriteString("\n")
}

// writeScorecard writes the 🏥 health scorecard: the five graded dimensions —
// average distance from the main sequence, max LCOM4, cycle count, overall
// duplication percentage, and instability spread — each with its value, its
// [Grade] from the classification functions, and the grade's status emoji,
// followed by the overall status line derived from the worst dimension
// grade. Grades are informational and never gate; [metrics.VerdictThresholds]
// remain the only enforcement path. A graph with no modules grades every
// dimension A — zero values sit in every top band — mirroring the summary
// mode's all-clear path.
func writeScorecard(b *strings.Builder, g *metrics.ModuleGraph) {
	agg := ComputeAggregates(g)
	b.WriteString("## 🏥 Health Scorecard\n\n")
	writeTableHeader(b, "Dimension", "Value", "Grade", "Status")
	dimensions := []struct {
		label string
		value string
		grade Grade
	}{
		{"Avg Distance from Main Sequence", fmt.Sprintf("%.2f", agg.AverageDistance), ClassifyDistance(agg.AverageDistance)},
		{"Max LCOM4", fmt.Sprintf("%d", agg.MaxLCOM4), ClassifyLCOM4(agg.MaxLCOM4)},
		{"Cycle Count", fmt.Sprintf("%d", agg.CycleCount), ClassifyCycles(agg.CycleCount)},
		{"Duplication %", fmt.Sprintf("%.1f%%", agg.DuplicationPercent), ClassifyDuplication(agg.DuplicationPercent)},
		{"Instability Spread", fmt.Sprintf("%d extremes", agg.InstabilitySpread), ClassifySpread(agg.InstabilitySpread)},
	}
	worst := GradeA
	for _, d := range dimensions {
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n", d.label, d.value, d.grade, d.grade.Emoji())
		if gradeRank(d.grade) > gradeRank(worst) {
			worst = d.grade
		}
	}
	fmt.Fprintf(b, "\n**Overall**: %s %s\n\n", worst.Emoji(), overallWord(worst))
}

// writeWarningsSection writes the ⚠️ warnings section: the analysis warnings
// carried by the snapshot — reproduced with their module attribution when
// the warning names one — followed by the packages classified in the Zone
// of Uselessness, in path order. Warnings order by module, then code, then
// message, so the section cannot depend on the order warnings appear in the
// snapshot; module-less (global) warnings sort first. A snapshot with no
// warnings and no Zone of Uselessness packages renders the no-warnings
// line.
func writeWarningsSection(b *strings.Builder, g *metrics.ModuleGraph) {
	b.WriteString("## ⚠️ Warnings\n\n")
	var lines []string
	if g != nil {
		warnings := append([]metrics.Warning(nil), g.Warnings...)
		sort.Slice(warnings, func(i, j int) bool { return warningLess(warnings[i], warnings[j]) })
		for _, w := range warnings {
			if w.Module != "" {
				lines = append(lines, fmt.Sprintf("- `%s`: %s", w.Module, w.Message))
			} else {
				lines = append(lines, "- "+w.Message)
			}
		}
	}
	for _, m := range sortedModules(g) {
		if m.Zone == metrics.ZoneOfUselessness {
			lines = append(lines, fmt.Sprintf("- `%s` is in the Zone of Uselessness", m.Path))
		}
	}
	if len(lines) == 0 {
		b.WriteString("No warnings.\n\n")
		return
	}
	for _, line := range lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// writeTableHeader writes a markdown table's header row and separator row.
// The separator's dashes span each header cell's rendered width so the table
// reads aligned in monospace viewers; markdown itself only requires at least
// one dash per column. Headers are ASCII by construction (the report's
// fixed column names), so len is the rendered width.
func writeTableHeader(b *strings.Builder, headers ...string) {
	b.WriteString("|")
	for _, h := range headers {
		fmt.Fprintf(b, " %s |", h)
	}
	b.WriteString("\n|")
	for _, h := range headers {
		b.WriteString(strings.Repeat("-", len(h)+2))
		b.WriteString("|")
	}
	b.WriteString("\n")
}

// zoneLabel returns the emoji-prefixed human label the package metrics
// table's Zone column renders for a module's [metrics.Zone] classification:
// 🟢 Main Sequence, 🟡 Balanced, 🔴 Zone of Pain, and 🔴 Zone of Uselessness,
// matching the example report's zone vocabulary. An unrecognized zone value —
// malformed input; adapters produce only the four defined zones, and the
// schema enforces the enum — renders as the raw machine string, so the table
// never silently drops or rewrites unknown data.
func zoneLabel(z metrics.Zone) string {
	switch z {
	case metrics.ZoneMainSequence:
		return "🟢 Main Sequence"
	case metrics.ZoneNormal:
		return "🟡 Balanced"
	case metrics.ZoneOfPain:
		return "🔴 Zone of Pain"
	case metrics.ZoneOfUselessness:
		return "🔴 Zone of Uselessness"
	default:
		return string(z)
	}
}

// sortedModules returns the graph's modules in the report's canonical module
// order: path ascending. A nil graph returns nil. The comparator is total —
// duplicate module paths (malformed input; the path is the module identity)
// tie-break on the scalar fields the report renders, and modules equal under
// all of them render identical rows — so no permutation of the Modules array
// can change any section's bytes. Duplication data needs no tiebreak: every
// section aggregates it with set semantics (dedup by [dupBlockKey],
// max-wins, counting), which is order-independent.
func sortedModules(g *metrics.ModuleGraph) []metrics.ModuleResult {
	if g == nil || len(g.Modules) == 0 {
		return nil
	}
	mods := append([]metrics.ModuleResult(nil), g.Modules...)
	sort.Slice(mods, func(i, j int) bool { return moduleLess(mods[i], mods[j]) })
	return mods
}

// moduleLess orders two modules for the report's canonical module order:
// path ascending, then the rendered scalar fields as a tiebreak for
// malformed duplicate-path graphs. See [sortedModules].
func moduleLess(a, b metrics.ModuleResult) bool {
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	if a.Ca != b.Ca {
		return a.Ca < b.Ca
	}
	if a.Ce != b.Ce {
		return a.Ce < b.Ce
	}
	if a.Instability != b.Instability {
		return a.Instability < b.Instability
	}
	if a.Abstractness != b.Abstractness {
		return a.Abstractness < b.Abstractness
	}
	if a.Distance != b.Distance {
		return a.Distance < b.Distance
	}
	if a.LCOM != b.LCOM {
		return a.LCOM < b.LCOM
	}
	if a.Zone != b.Zone {
		return a.Zone < b.Zone
	}
	return a.TotalLines < b.TotalLines
}

// sortedCycles returns the graph's cycles in the report's canonical cycle
// order: lexicographic by member sequence, the same order [cycleRuleHits]
// uses to pick representative cycles. Zero-member cycles — schema-valid but
// degenerate, naming no members to render — are dropped, so the returned
// set is exactly the renderable chains. The sort is defensive: the [metrics.Cycle]
// contract already sorts cycles, but rendering MUST NOT depend on the order
// cycles appear in the snapshot.
func sortedCycles(g *metrics.ModuleGraph) []metrics.Cycle {
	if g == nil {
		return nil
	}
	return sortedCycleList(g.Cycles)
}

// sortedCycleList returns the renderable cycles of a cycle list in the
// canonical cycle order: lexicographic by member sequence (see
// [cycleLess]). Zero-member cycles — schema-valid but degenerate, naming
// no members to render — are dropped, so the returned set is exactly the
// renderable chains. The sort is defensive: the [metrics.Cycle] contract already
// sorts cycles, but rendering MUST NOT depend on the order cycles appear
// in the input. The diff markdown renderer shares it for its new and
// resolved cycle sections.
func sortedCycleList(cycles []metrics.Cycle) []metrics.Cycle {
	var out []metrics.Cycle
	for _, c := range cycles {
		if len(c) > 0 {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return cycleLess(out[i], out[j]) })
	return out
}

// cycleChain renders one [metrics.Cycle] as a closed dependency chain: the members —
// the sorted membership set, per the Cycle contract — joined with "→", with
// the first member repeated at the end to close the loop (pkg/a → pkg/b →
// pkg/a). The caller guarantees a non-empty cycle (see [sortedCycles]).
func cycleChain(c metrics.Cycle) string {
	return strings.Join(append(append([]string(nil), c...), c[0]), " → ")
}

// duplicationRow aggregates every reference the snapshot makes to one
// deduplicated duplication block: its identity (file plus 1-indexed
// inclusive line range), the largest reported significant line count, the
// highest similarity among the duplication groups referencing it, and how
// many group references it received.
type duplicationRow struct {
	file        string
	startLine   int
	endLine     int
	lines       int
	similarity  float64
	occurrences int
}

// duplicationRows aggregates a module set's duplication blocks into one
// [duplicationRow] per deduplicated block identity — file plus line range,
// the [dupBlockKey] semantics [ComputeAggregates] applies — so a block
// reported by several duplication groups or modules counts once. Max-wins
// aggregation (largest line count, highest similarity) and reference
// counting are commutative, so the result is independent of the order
// modules and duplications appear in. Rows order by file, then line range,
// making the table a pure function of the snapshot's block multiset.
func duplicationRows(mods []metrics.ModuleResult) []duplicationRow {
	rows := make(map[dupBlockKey]*duplicationRow)
	for _, m := range mods {
		for _, d := range m.Duplications {
			for _, blk := range d.Blocks {
				k := dupBlockKey{file: blk.File, startLine: blk.StartLine, endLine: blk.EndLine}
				row := rows[k]
				if row == nil {
					rows[k] = &duplicationRow{
						file:        blk.File,
						startLine:   blk.StartLine,
						endLine:     blk.EndLine,
						lines:       blk.LineCount,
						similarity:  d.Similarity,
						occurrences: 1,
					}
					continue
				}
				row.occurrences++
				if blk.LineCount > row.lines {
					row.lines = blk.LineCount
				}
				if d.Similarity > row.similarity {
					row.similarity = d.Similarity
				}
			}
		}
	}
	out := make([]duplicationRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].file != out[j].file {
			return out[i].file < out[j].file
		}
		if out[i].startLine != out[j].startLine {
			return out[i].startLine < out[j].startLine
		}
		return out[i].endLine < out[j].endLine
	})
	return out
}

// gradeRank maps a grade to its ordinal position on the A–F scale: A best
// (0) through F worst (5). Unknown grades rank below F — worst — so an
// unrecognized value can never silently improve the overall status,
// mirroring [severityRank]'s treatment of unknown severities. The
// classifiers only produce the six defined grades, so the default case
// exists only to keep the ranking total.
func gradeRank(g Grade) int {
	switch g {
	case GradeA:
		return 0
	case GradeBPlus:
		return 1
	case GradeB:
		return 2
	case GradeCPlus:
		return 3
	case GradeC:
		return 4
	case GradeF:
		return 5
	default:
		return 6
	}
}

// overallWord maps the worst dimension grade to the scorecard's overall
// status word: Healthy for the green bands (A, B+), Mixed for the yellow
// bands (B, C+), and Needs Attention for the red bands (C, F) — reusing the
// summary status line's vocabulary for the worst case. An unrecognized
// grade falls back to its raw string.
func overallWord(g Grade) string {
	switch g {
	case GradeA, GradeBPlus:
		return "Healthy"
	case GradeB, GradeCPlus:
		return "Mixed"
	case GradeC, GradeF:
		return StatusNeedsAttention
	default:
		return string(g)
	}
}

// warningLess orders two warnings for the warnings section: module path
// ascending (the empty path of module-less, global warnings first), then
// code, then message — a total order, so the section cannot depend on the
// order warnings appear in the snapshot.
func warningLess(a, b metrics.Warning) bool {
	if a.Module != b.Module {
		return a.Module < b.Module
	}
	if a.Code != b.Code {
		return a.Code < b.Code
	}
	return a.Message < b.Message
}

// backtickListOrNone renders paths as backticked, comma-separated markdown
// code spans, or "none" for an empty list — the agent template's
// list-or-none contract.
func backtickListOrNone(paths []string) string {
	if len(paths) == 0 {
		return "none"
	}
	return backtickList(paths)
}

// orNone renders pre-formatted list items joined with ", ", or "none" for
// an empty list.
func orNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}

// reportProducer resolves the producer name and version for the report
// header. Resolution is per-field and total: a field from the snapshot's
// provenance wins when non-empty; otherwise the RenderOptions-injected
// rendering-binary identity applies; otherwise the package defaults. No
// input yields an empty string, so the header line is always well-formed —
// even for graphs whose provenance object carries empty fields.
func reportProducer(g *metrics.ModuleGraph, opts RenderOptions) (producer, version string) {
	producer, version = opts.FallbackProducer, opts.FallbackVersion
	if g != nil && g.Provenance != nil {
		if g.Provenance.Producer != "" {
			producer = g.Provenance.Producer
		}
		if g.Provenance.Version != "" {
			version = g.Provenance.Version
		}
	}
	if producer == "" {
		producer = DefaultRenderProducer
	}
	if version == "" {
		version = DefaultRenderVersion
	}
	return producer, version
}

// graphStats returns the module count and the total analyzed lines, summed
// across all modules' TotalLines. A nil graph counts as zero of each, so
// section writers stay nil-safe without each repeating the guard.
func graphStats(g *metrics.ModuleGraph) (modules, totalLines int) {
	if g == nil {
		return 0, 0
	}
	for _, m := range g.Modules {
		totalLines += m.TotalLines
	}
	return len(g.Modules), totalLines
}

// countWord renders a count with a singular or plural noun: "1 package",
// "8 packages", "0 lines", "300 lines". Both nouns the report renders
// (package, line) pluralize regularly with a trailing "s", so no irregular
// forms are needed.
func countWord(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
