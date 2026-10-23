package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// diffTitle is the first heading of every diff markdown report: the 🏗️
// title line of the report's fixed emoji vocabulary, naming the base-vs-PR
// comparison the report renders.
const diffTitle = "# 🏗️ Structural Entropy Diff"

// RenderDiffMarkdown renders the base-vs-PR structural-entropy diff as a
// CI-publishable markdown report: the 🏗️ title heading, the verdict and
// entropy-direction line, the ⚠️ triggering reasons, the 📊 per-module
// metric delta table (PR minus base), and the 🔄 new and resolved cycle
// chains. verdict and reasons are the already-computed outputs of
// [metrics.DecideVerdict]; rendering is display-only and never feeds verdict or
// gate computation, so the markdown view cannot alter exit semantics.
//
// It is a PURE, DETERMINISTIC function of (delta, verdict, reasons): built
// with strings.Builder per design D6, no text/template, no I/O, no
// timestamps. Deltas render in the canonical path order and cycles in the
// canonical member-sequence order regardless of the order the delta's
// slices carry, and reasons render verbatim in the stable order
// [metrics.DecideVerdict] returns (new-cycle reasons first, then per-module
// reasons by path) — the same order the default table consumes — so
// identical input renders byte-identical output.
//
// Ratio deltas render with the diff's fixed 4-decimal precision (%.4f,
// matching the verdict reasons' threshold strings and the default table),
// normalized so a delta of exactly zero renders "0.0000" rather than
// "-0.0000"; coupling and LCOM4 deltas render as signed integers. A delta
// with no shared modules renders the no-shared-packages note instead of a
// header-only table, and empty reason or cycle sections render explicit
// none lines, keeping the section structure fixed. Rendering cannot fail,
// so no error is returned; the report ends with exactly one final newline.
func RenderDiffMarkdown(d metrics.GraphDelta, verdict metrics.Verdict, reasons []string) string {
	var b strings.Builder
	b.WriteString(diffTitle)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "**Verdict**: %s | **Entropy direction**: %s\n\n", verdict, d.Direction)
	writeDiffReasons(&b, reasons)
	writeDiffModuleDeltas(&b, d.Modules)
	writeDiffCycleSection(&b, "## 🔄 New Cycles", "🟢 No new circular dependencies.", emojiRed, "detected", d.NewCycles)
	writeDiffCycleSection(&b, "## 🔄 Resolved Cycles", "None.", emojiGreen, "resolved", d.ResolvedCycles)
	return finalizeReport(&b)
}

// writeDiffReasons writes the ⚠️ verdict reasons section: one bullet per
// triggering reason, rendered verbatim in the stable order [metrics.DecideVerdict]
// returns, or the explicit none line when the verdict fired no gates.
// Rendering the none line — rather than omitting the section — keeps the
// report's section structure fixed for golden tests and markdown consumers.
func writeDiffReasons(b *strings.Builder, reasons []string) {
	b.WriteString("## ⚠️ Verdict Reasons\n\n")
	if len(reasons) == 0 {
		b.WriteString("None.\n\n")
		return
	}
	for _, r := range reasons {
		fmt.Fprintf(b, "- %s\n", r)
	}
	b.WriteString("\n")
}

// writeDiffModuleDeltas writes the 📊 per-module delta table: one row per
// module present in both graphs (PR value minus base value) in the
// renderer's canonical path order, with the report's full-word column
// headers and the same header/separator writer ([writeTableHeader]) the
// package metrics table uses. Ratio deltas render with %.4f and are
// normalized through [NormFloat]; coupling and LCOM4 deltas render as
// signed integers. A delta with no shared modules renders the
// no-shared-packages note instead of a header-only table.
func writeDiffModuleDeltas(b *strings.Builder, deltas []metrics.Delta) {
	b.WriteString("## 📊 Per-module Deltas (PR minus base)\n\n")
	sorted := sortedDeltas(deltas)
	if len(sorted) == 0 {
		b.WriteString("No shared packages.\n\n")
		return
	}
	writeTableHeader(b, "Package", "Instability", "Abstractness", "Distance", "LCOM4", "Ca", "Ce")
	for _, m := range sorted {
		fmt.Fprintf(b, "| `%s` | %.4f | %.4f | %.4f | %d | %d | %d |\n",
			m.Path,
			NormFloat(m.Instability), NormFloat(m.Abstractness), NormFloat(m.Distance),
			m.LCOM, m.Ca, m.Ce)
	}
	b.WriteString("\n")
}

// writeDiffCycleSection writes one 🔄 cycle section of the diff report:
// the heading, then either the given none line (empty list) or the
// marker-prefixed count line and a fenced code block of closed chains —
// the same chain rendering ([cycleChain]) the report's circular
// dependencies section uses. verb forms the count line ("1 cycle
// detected", "1 cycle resolved"); marker is the count line's status emoji.
// Cycles render in the canonical member-sequence order (see
// [sortedCycleList]), so the section cannot depend on the order the
// delta's slice carries.
func writeDiffCycleSection(b *strings.Builder, heading, noneLine, marker, verb string, cycles []metrics.Cycle) {
	b.WriteString(heading)
	b.WriteString("\n\n")
	sorted := sortedCycleList(cycles)
	if len(sorted) == 0 {
		b.WriteString(noneLine)
		b.WriteString("\n\n")
		return
	}
	fmt.Fprintf(b, "%s **%s %s:**\n\n", marker, countWord(len(sorted), "cycle"), verb)
	b.WriteString("```\n")
	for _, c := range sorted {
		b.WriteString(cycleChain(c))
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
}

// sortedDeltas returns the per-module deltas in the renderer's canonical
// order: path ascending. The comparator is total — duplicate paths
// (malformed input; [metrics.ComputeDelta] keys by path so it never emits them)
// tie-break on the scalar fields the table renders, and deltas equal under
// all of them render identical rows — so no permutation of the input slice
// can change the section's bytes.
func sortedDeltas(deltas []metrics.Delta) []metrics.Delta {
	if len(deltas) == 0 {
		return nil
	}
	out := append([]metrics.Delta(nil), deltas...)
	sort.Slice(out, func(i, j int) bool { return deltaLess(out[i], out[j]) })
	return out
}

// deltaLess orders two per-module deltas for the renderer's canonical
// order: path ascending, then the rendered scalar fields as a tiebreak.
// See [sortedDeltas].
func deltaLess(a, b metrics.Delta) bool {
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
	return a.LCOM < b.LCOM
}

// NormFloat maps negative zero to positive zero so a delta of exactly zero
// always renders as "0.0000" rather than "-0.0000", keeping delta tables
// stable. Returns the input unchanged for all other values. It is the
// single shared definition of the normalization: this package's diff
// markdown renderer and the CLI's human-readable delta table both route
// through it, so the two surfaces cannot drift on zero-delta rendering.
func NormFloat(f float64) float64 {
	if f == 0 {
		return 0
	}
	return f
}
