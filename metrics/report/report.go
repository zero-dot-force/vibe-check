package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Grade is a health scorecard grade on an A–F ordinal scale. The
// deterministic report grades five scorecard dimensions — average
// distance from main sequence, max LCOM4, circular dependency count,
// overall duplication percentage, and instability spread — each against
// the fixed threshold table defined by its Grade threshold constants
// below.
//
// The scale runs from GradeA (best) to GradeF (worst) and deliberately
// omits a D grade: the threshold tables grade C as the last passing band
// and drop straight to F. Values serialize as their letter strings
// ("A", "B+", "B", "C+", "C", "F").
//
// Grades are informational only — they advise, they never gate. The
// entropy-regression thresholds in [metrics.VerdictThresholds] (applied by
// [metrics.DecideVerdict]) remain the only enforcement thresholds, and these
// grade tables MUST NOT be wired into CI gating.
type Grade string

const (
	// GradeA is the best health grade: the dimension value sits in the
	// top band of its threshold table.
	GradeA Grade = "A"
	// GradeBPlus is the second-best health grade.
	GradeBPlus Grade = "B+"
	// GradeB is a middle health grade: acceptable, with room to improve.
	GradeB Grade = "B"
	// GradeCPlus is a lower-middle health grade.
	GradeCPlus Grade = "C+"
	// GradeC is the last passing health grade before failure.
	GradeC Grade = "C"
	// GradeF is the worst health grade: the dimension value sits in the
	// bottom band of its threshold table.
	GradeF Grade = "F"
)

// Status emoji for grade rendering, drawn from the report's fixed
// formatting vocabulary. They are unexported because [Grade.Emoji] is the
// public mapping; rendering code consumes grades through it.
const (
	emojiGreen  = "🟢"
	emojiYellow = "🟡"
	emojiRed    = "🔴"
)

// Emoji returns the status emoji for the grade: 🟢 for GradeA and
// GradeBPlus (healthy), 🟡 for GradeB and GradeCPlus (watch), and 🔴 for
// GradeC and GradeF (needs attention). The mapping is fixed by the report
// spec's grade-status-emoji requirement.
//
// An unrecognized grade returns the empty string. Grade classification
// functions only ever produce the six defined grades, so an empty result
// indicates a programming error elsewhere, not a reportable condition.
func (g Grade) Emoji() string {
	switch g {
	case GradeA, GradeBPlus:
		return emojiGreen
	case GradeB, GradeCPlus:
		return emojiYellow
	case GradeC, GradeF:
		return emojiRed
	default:
		return ""
	}
}

// Average-distance grade thresholds for the health scorecard. The graded
// value is the mean of all module [metrics.Distance] values: a dimensionless
// ratio in [0.0, 1.0], where 0.0 lies on the main sequence. Each constant
// is the inclusive minimum of its grade's band, so a value that lands
// exactly on a boundary belongs to the lower grade's band:
//
//	A:   avg < 0.1
//	B+:  0.1 <= avg < 0.2
//	B:   0.2 <= avg < 0.3
//	C+:  0.3 <= avg < 0.5
//	C:   0.5 <= avg < 0.7
//	F:   0.7 <= avg
//
// For example, an average distance of exactly 0.1 grades B+, because A
// requires strictly less than 0.1. These thresholds are informational
// scorecard grades, not gates; [metrics.VerdictThresholds] gates enforcement.
const (
	// DistanceGradeBPlus is the inclusive minimum of the B+ band: an
	// average distance of at least 0.1 and under 0.2 grades B+. Unit:
	// dimensionless ratio in [0.0, 1.0].
	DistanceGradeBPlus = 0.1
	// DistanceGradeB is the inclusive minimum of the B band: an average
	// distance of at least 0.2 and under 0.3 grades B. Unit:
	// dimensionless ratio in [0.0, 1.0].
	DistanceGradeB = 0.2
	// DistanceGradeCPlus is the inclusive minimum of the C+ band: an
	// average distance of at least 0.3 and under 0.5 grades C+. Unit:
	// dimensionless ratio in [0.0, 1.0].
	DistanceGradeCPlus = 0.3
	// DistanceGradeC is the inclusive minimum of the C band: an average
	// distance of at least 0.5 and under 0.7 grades C. Unit:
	// dimensionless ratio in [0.0, 1.0].
	DistanceGradeC = 0.5
	// DistanceGradeF is the inclusive minimum of the F band: an average
	// distance of at least 0.7 grades F. Unit: dimensionless ratio in
	// [0.0, 1.0].
	DistanceGradeF = 0.7
)

// Max-LCOM4 grade thresholds for the health scorecard. The graded value
// is the maximum [metrics.LCOM] (LCOM4 variant) across all modules: a
// non-negative integer counting connected components in the method-field
// graph, where 1 means fully cohesive and values above 1 mean the module
// can be split into that many independent pieces. Each constant is the
// inclusive minimum of its grade's band:
//
//	A:   max LCOM4 <= 1
//	B+:  2 <= max LCOM4 < 3
//	B:   3 <= max LCOM4 < 4
//	C+:  4 <= max LCOM4 < 5
//	C:   5 <= max LCOM4 < 6
//	F:   6 <= max LCOM4
//
// For example, a max LCOM4 of exactly 4 grades C+. A max LCOM4 of 0 (all
// modules trivially cohesive, with no methods) also grades A, since it
// sits below the B+ band. These thresholds are informational scorecard
// grades, not gates; [metrics.VerdictThresholds] gates enforcement.
const (
	// LCOMGradeBPlus is the inclusive minimum of the B+ band: a max LCOM4
	// of at least 2 and under 3 grades B+. Unit: non-negative integer
	// (connected components).
	LCOMGradeBPlus = 2
	// LCOMGradeB is the inclusive minimum of the B band: a max LCOM4 of
	// at least 3 and under 4 grades B. Unit: non-negative integer
	// (connected components).
	LCOMGradeB = 3
	// LCOMGradeCPlus is the inclusive minimum of the C+ band: a max LCOM4
	// of at least 4 and under 5 grades C+. Unit: non-negative integer
	// (connected components).
	LCOMGradeCPlus = 4
	// LCOMGradeC is the inclusive minimum of the C band: a max LCOM4 of
	// at least 5 and under 6 grades C. Unit: non-negative integer
	// (connected components).
	LCOMGradeC = 5
	// LCOMGradeF is the inclusive minimum of the F band: a max LCOM4 of
	// at least 6 grades F. Unit: non-negative integer (connected
	// components).
	LCOMGradeF = 6
)

// Circular-dependency grade thresholds for the health scorecard. The
// graded value is the count of circular dependencies in the graph: a
// non-negative integer. The scale is binary — zero cycles grades A, and
// any cycle at all grades F — so the table has a single boundary:
//
//	A:   cycles == 0
//	F:   1 <= cycles
//
// These thresholds are informational scorecard grades, not gates;
// [metrics.VerdictThresholds] gates enforcement (a cycle newly introduced by a
// PR remains a REQUEST_CHANGES gate there).
const (
	// CycleGradeF is the inclusive minimum of the F band: a circular
	// dependency count of at least 1 grades F; only zero cycles grades A.
	// Unit: non-negative integer (detected dependency cycles).
	CycleGradeF = 1
)

// Duplication grade thresholds for the health scorecard. The graded value
// is the overall duplication percentage: deduplicated duplicated lines
// divided by the total lines summed across all modules, times 100 — a
// percentage in [0.0, 100.0]. Each constant is the inclusive minimum of
// its grade's band, with one irregularity: A is a point band at exactly
// zero, and the B+ band is open at zero (any non-zero duplication under
// 3%):
//
//	A:   dup == 0
//	B+:  0 < dup < 3
//	B:   3 <= dup < 5
//	C+:  5 <= dup < 10
//	C:   10 <= dup < 15
//	F:   15 <= dup
//
// For example, an overall duplication of exactly 5% grades C+, because B
// requires strictly less than 5%. There is no DuplicationGradeBPlus
// constant because the B+ band has no inclusive minimum — it begins
// strictly above zero. These thresholds are informational scorecard
// grades, not gates; [metrics.VerdictThresholds] gates enforcement.
const (
	// DuplicationGradeB is the inclusive minimum of the B band: an
	// overall duplication of at least 3% and under 5% grades B. Unit:
	// percent in [0.0, 100.0].
	DuplicationGradeB = 3.0
	// DuplicationGradeCPlus is the inclusive minimum of the C+ band: an
	// overall duplication of at least 5% and under 10% grades C+. Unit:
	// percent in [0.0, 100.0].
	DuplicationGradeCPlus = 5.0
	// DuplicationGradeC is the inclusive minimum of the C band: an
	// overall duplication of at least 10% and under 15% grades C. Unit:
	// percent in [0.0, 100.0].
	DuplicationGradeC = 10.0
	// DuplicationGradeF is the inclusive minimum of the F band: an
	// overall duplication of at least 15% grades F. Unit: percent in
	// [0.0, 100.0].
	DuplicationGradeF = 15.0
)

// Instability-spread grade thresholds for the health scorecard. The
// graded value is the instability spread: the count of modules at
// instability extremes — instability at or below 0.001, or at or above
// 0.999, per the aggregate computation — a non-negative integer in
// [0, module count]. Each constant is the inclusive minimum of its
// grade's band, with A as the point band at zero:
//
//	A:   spread == 0
//	B:   1 <= spread < 2
//	C+:  2 <= spread < 3
//	F:   3 <= spread
//
// These thresholds are informational scorecard grades, not gates;
// [metrics.VerdictThresholds] gates enforcement.
const (
	// SpreadGradeB is the inclusive minimum of the B band: an instability
	// spread of at least 1 and under 2 grades B. Unit: non-negative
	// integer (count of extreme-instability modules).
	SpreadGradeB = 1
	// SpreadGradeCPlus is the inclusive minimum of the C+ band: an
	// instability spread of at least 2 and under 3 grades C+. Unit:
	// non-negative integer (count of extreme-instability modules).
	SpreadGradeCPlus = 2
	// SpreadGradeF is the inclusive minimum of the F band: an instability
	// spread of at least 3 grades F. Unit: non-negative integer (count of
	// extreme-instability modules).
	SpreadGradeF = 3
)

// ClassifyDistance grades the health scorecard's average-distance
// dimension. The graded value avg is the mean of all module [metrics.Distance]
// values: a dimensionless ratio in [0.0, 1.0], where 0.0 lies on the main
// sequence. Classification cascades down the threshold table documented
// on the [DistanceGradeBPlus] through [DistanceGradeF] constants, each
// the inclusive minimum of its band, so a value landing exactly on a
// boundary belongs to the lower grade's band: an average distance of
// exactly 0.1 grades B+, because A requires strictly less than 0.1.
// Values below 0.1 grade A. The function is pure and deterministic; the
// grade is informational only and never gates — [metrics.DecideVerdict] remains
// the only enforcement path.
//
// The cascade's parallel structure with the other Classify* functions is
// deliberate: duplication sits below the 3+ extraction threshold, and each
// band's bounds are documented on its own constant.
func ClassifyDistance(avg metrics.Distance) Grade {
	switch {
	case avg >= DistanceGradeF:
		return GradeF
	case avg >= DistanceGradeC:
		return GradeC
	case avg >= DistanceGradeCPlus:
		return GradeCPlus
	case avg >= DistanceGradeB:
		return GradeB
	case avg >= DistanceGradeBPlus:
		return GradeBPlus
	default:
		return GradeA
	}
}

// ClassifyLCOM4 grades the health scorecard's max-LCOM4 dimension. The
// graded value maxLCOM4 is the maximum [metrics.LCOM] (LCOM4 variant) across all
// modules: a non-negative integer counting connected components in the
// method-field graph, where 1 means fully cohesive and 0 means no
// methods. Classification cascades down the threshold table documented
// on the [LCOMGradeBPlus] through [LCOMGradeF] constants, each the
// inclusive minimum of its band: a max LCOM4 of exactly 4 grades C+.
// Values below 2 (0 or 1) grade A. The function is pure and
// deterministic; the grade is informational only and never gates —
// [metrics.DecideVerdict] remains the only enforcement path.
//
// The cascade's parallel structure with the other Classify* functions is
// deliberate: duplication sits below the 3+ extraction threshold, and each
// band's bounds are documented on its own constant.
func ClassifyLCOM4(maxLCOM4 metrics.LCOM) Grade {
	switch {
	case maxLCOM4 >= LCOMGradeF:
		return GradeF
	case maxLCOM4 >= LCOMGradeC:
		return GradeC
	case maxLCOM4 >= LCOMGradeCPlus:
		return GradeCPlus
	case maxLCOM4 >= LCOMGradeB:
		return GradeB
	case maxLCOM4 >= LCOMGradeBPlus:
		return GradeBPlus
	default:
		return GradeA
	}
}

// ClassifyCycles grades the health scorecard's circular-dependency
// dimension. The graded value count is the number of circular
// dependencies in the graph: a non-negative integer. The scale is binary
// per the [CycleGradeF] boundary — zero cycles grades A, and any cycle at
// all (a count of 1 or more) grades F. The function is pure and
// deterministic; the grade is informational only and never gates — a
// cycle newly introduced by a PR remains a REQUEST_CHANGES gate in
// [metrics.DecideVerdict].
func ClassifyCycles(count int) Grade {
	if count >= CycleGradeF {
		return GradeF
	}
	return GradeA
}

// ClassifyDuplication grades the health scorecard's duplication
// dimension. The graded value pct is the overall duplication percentage:
// deduplicated duplicated lines divided by the total lines summed across
// all modules, times 100 — a percentage in [0.0, 100.0]. Classification
// cascades down the threshold table documented on the [DuplicationGradeB]
// through [DuplicationGradeF] constants, with the table's one
// irregularity: A is a point band at exactly 0%, and the B+ band is open
// at zero, so any non-zero duplication under 3% grades B+ (there is no
// DuplicationGradeBPlus constant). A percentage of exactly 5% grades C+,
// because B requires strictly less than 5%. The function is pure and
// deterministic; the grade is informational only and never gates —
// [metrics.DecideVerdict] remains the only enforcement path.
func ClassifyDuplication(pct float64) Grade {
	switch {
	case pct >= DuplicationGradeF:
		return GradeF
	case pct >= DuplicationGradeC:
		return GradeC
	case pct >= DuplicationGradeCPlus:
		return GradeCPlus
	case pct >= DuplicationGradeB:
		return GradeB
	case pct > 0:
		// A is the point band at exactly 0%; the B+ band is open at
		// zero, so any non-zero duplication under 3% grades B+.
		return GradeBPlus
	default:
		return GradeA
	}
}

// ClassifySpread grades the health scorecard's instability-spread
// dimension. The graded value count is the instability spread: the number
// of modules at instability extremes — instability at or below 0.001 or
// at or above 0.999 — a non-negative integer in [0, module count].
// Classification cascades down the threshold table documented on the
// [SpreadGradeB] through [SpreadGradeF] constants, with A as the point
// band at zero: a spread of exactly 1 grades B, 2 grades C+, and 3 or
// more grades F. The function is pure and deterministic; the grade is
// informational only and never gates — [metrics.DecideVerdict] remains the only
// enforcement path.
func ClassifySpread(count int) Grade {
	switch {
	case count >= SpreadGradeF:
		return GradeF
	case count >= SpreadGradeCPlus:
		return GradeCPlus
	case count >= SpreadGradeB:
		return GradeB
	default:
		return GradeA
	}
}

// Instability-spread extreme bounds for the aggregate computation. The
// instability spread counts modules at the extremes of the instability
// scale: a module whose [metrics.Instability] is at or below SpreadStableMax is a
// stable extreme (many dependents, few dependencies — costly to change),
// and a module whose instability is at or above SpreadUnstableMin is an
// unstable extreme (few dependents, many dependencies — isolated from the
// rest of the graph). The bounds are inclusive, so exact 0.0 and 1.0 values
// always count; the small tolerance also absorbs values such as
// Ce/(Ca+Ce) = 1000/1001 = 0.9990009… that are effectively, though not
// exactly, maximally unstable.
//
// The spread is an informational scorecard input — it feeds
// [ClassifySpread] and the report's aggregate output; it never gates.
// [metrics.VerdictThresholds] gates enforcement.
const (
	// SpreadStableMax is the inclusive maximum instability counted as a
	// stable extreme: a module with instability at or below 0.001 counts
	// toward the instability spread. Unit: dimensionless ratio in
	// [0.0, 1.0].
	SpreadStableMax = 0.001
	// SpreadUnstableMin is the inclusive minimum instability counted as an
	// unstable extreme: a module with instability at or above 0.999 counts
	// toward the instability spread. Unit: dimensionless ratio in
	// [0.0, 1.0].
	SpreadUnstableMin = 0.999
)

// Aggregates holds the deterministic aggregate values the report computes
// over a [metrics.ModuleGraph]: the inputs to the status summary line and to the
// health scorecard's five graded dimensions. [ComputeAggregates] produces
// it as a pure function, and every field is defined for the empty module
// set, so callers never divide by zero or special-case emptiness.
//
// Field types match the grade classifiers so each dimension value flows to
// its classifier without conversion: AverageDistance feeds
// [ClassifyDistance], MaxLCOM4 feeds [ClassifyLCOM4], CycleCount feeds
// [ClassifyCycles], InstabilitySpread feeds [ClassifySpread], and
// DuplicationPercent feeds [ClassifyDuplication].
type Aggregates struct {
	// AverageInstability is the arithmetic mean of all module
	// [metrics.Instability] values. Unit: dimensionless ratio in [0.0, 1.0],
	// where 0.0 is maximally stable. Defined as 0.0 when the graph has
	// no modules.
	AverageInstability metrics.Instability `json:"averageInstability"`
	// AverageDistance is the arithmetic mean of all module [metrics.Distance]
	// values. Unit: dimensionless ratio in [0.0, 1.0], where 0.0 lies on
	// the main sequence. Defined as 0.0 when the graph has no modules.
	AverageDistance metrics.Distance `json:"averageDistance"`
	// MaxLCOM4 is the maximum [metrics.LCOM] (LCOM4 variant) across all modules.
	// Unit: non-negative integer (connected components), where 1 means
	// fully cohesive. Defined as 0 when the graph has no modules.
	MaxLCOM4 metrics.LCOM `json:"maxLCOM4"`
	// CycleCount is the number of circular dependencies carried by the
	// graph's Cycles field. Unit: non-negative integer (detected
	// dependency cycles).
	CycleCount int `json:"cycleCount"`
	// InstabilitySpread is the count of modules at instability extremes:
	// instability at or below [SpreadStableMax] (0.001) or at or above
	// [SpreadUnstableMin] (0.999). Unit: non-negative integer in
	// [0, module count]. Defined as 0 when the graph has no modules.
	InstabilitySpread int `json:"instabilitySpread"`
	// DuplicationPercent is the overall duplication percentage: the sum
	// of deduplicated duplication block line counts, divided by the
	// total lines summed across all modules, times 100. Unit: percent,
	// nominally in [0.0, 100.0] — the value is a pure ratio of reported
	// line counts, so a malformed snapshot with inconsistent TotalLines
	// can exceed 100.0. Defined as 0.0 when the summed total lines are
	// zero.
	DuplicationPercent float64 `json:"duplicationPercent"`
}

// ComputeAggregates computes the report's [Aggregates] over a
// [metrics.ModuleGraph]. It is a PURE, DETERMINISTIC function: no I/O, no globals,
// no error return, and identical inputs always yield identical output.
//
// Behavior:
//   - A nil graph returns the zero Aggregates (every field zero), never
//     panicking — mirroring [metrics.ComputeDelta]'s nil handling.
//   - When the graph has no modules, every module-derived aggregate is
//     defined as its zero value (averages 0.0, max LCOM4 0, spread 0,
//     duplication 0.0%) rather than dividing by zero; CycleCount still
//     mirrors the Cycles field length.
//   - Averages sum module values in ascending order rather than in Modules
//     slice order: floating-point addition is not associative, and a
//     canonical summation order keeps the means — and therefore the report
//     output — unchanged when the modules array is reordered, as the
//     report's byte-stability requirement demands.
//   - Duplication blocks are deduplicated by file and line range before
//     summing, so a block reported by several duplication groups or
//     modules counts once; see [duplicationPercent] for the dedup key.
func ComputeAggregates(g *metrics.ModuleGraph) Aggregates {
	if g == nil {
		return Aggregates{}
	}

	instabilities := make([]float64, 0, len(g.Modules))
	distances := make([]float64, 0, len(g.Modules))
	var maxLCOM4 metrics.LCOM
	spread := 0
	for _, m := range g.Modules {
		instabilities = append(instabilities, float64(m.Instability))
		distances = append(distances, float64(m.Distance))
		if m.LCOM > maxLCOM4 {
			maxLCOM4 = m.LCOM
		}
		if m.Instability <= SpreadStableMax || m.Instability >= SpreadUnstableMin {
			spread++
		}
	}

	return Aggregates{
		AverageInstability: metrics.Instability(sortedMean(instabilities)),
		AverageDistance:    metrics.Distance(sortedMean(distances)),
		MaxLCOM4:           maxLCOM4,
		CycleCount:         len(g.Cycles),
		InstabilitySpread:  spread,
		DuplicationPercent: duplicationPercent(g.Modules),
	}
}

// sortedMean returns the arithmetic mean of values, or 0.0 for an empty
// slice. The values are summed in ascending order rather than slice order
// because floating-point addition is not associative: a canonical order
// makes the mean a function of the value multiset alone, so reordering the
// Modules slice of a graph cannot change the result.
func sortedMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	return sum / float64(len(sorted))
}

// dupBlockKey is the deduplication identity of a [metrics.DuplicateBlock]: its
// file path plus its 1-indexed inclusive line range. ModulePath is
// deliberately excluded — file paths are relative to the project root, so
// the same block reported by two duplication groups or by two modules is
// one block counted once. LineCount is likewise excluded because it is
// derived from the same file and range.
type dupBlockKey struct {
	file      string
	startLine int
	endLine   int
}

// duplicationPercent computes the overall duplication percentage over a
// module set: the sum of [DuplicateBlock.LineCount] across all duplication
// blocks in all modules, deduplicated by [dupBlockKey], divided by the
// [metrics.ModuleResult.TotalLines] summed across all modules, times 100. It
// returns 0.0 when the summed total lines are zero (an empty module set,
// or line counts that could not be determined) rather than dividing by
// zero.
//
// When the same block identity is reported with conflicting LineCount
// values — malformed input; well-formed graphs report one count per file
// and range — the largest reported value is used, keeping the result
// independent of the order modules and duplications appear in the graph.
func duplicationPercent(mods []metrics.ModuleResult) float64 {
	totalLines := 0
	blockLines := make(map[dupBlockKey]int)
	for _, m := range mods {
		totalLines += m.TotalLines
		for _, d := range m.Duplications {
			for _, b := range d.Blocks {
				k := dupBlockKey{file: b.File, startLine: b.StartLine, endLine: b.EndLine}
				if b.LineCount > blockLines[k] {
					blockLines[k] = b.LineCount
				}
			}
		}
	}
	if totalLines == 0 {
		return 0
	}
	dupLines := 0
	for _, lines := range blockLines {
		dupLines += lines
	}
	return float64(dupLines) / float64(totalLines) * 100
}

// RecommendationSeverity is the severity tier of a [Recommendation]
// produced by the report's rule engine. Values are the machine-readable
// strings the report spec fixes for JSON serialization — "red",
// "yellow", and "green" — rather than emoji; rendering maps them onto
// the report's fixed emoji vocabulary (🔴 🟡 🟢).
//
// Red and yellow are violation severities: the report's status summary
// reads "Needs Attention" and the report command exits 1 when any of
// them fires. Green is informational: a green-only report reads
// "All Clear" and exits 0.
type RecommendationSeverity string

const (
	// SeverityRed marks a violation-severity recommendation from the
	// rule table's red tier: LCOM4 split, distance refactor, cycle
	// break, or duplication eliminate.
	SeverityRed RecommendationSeverity = "red"
	// SeverityYellow marks a violation-severity recommendation from the
	// rule table's yellow tier: instability isolate, instability
	// protect, Ce decouple, or abstractness add-interfaces.
	SeverityYellow RecommendationSeverity = "yellow"
	// SeverityGreen marks an informational recommendation: the
	// duplication watch rule or the all-clear fallback.
	SeverityGreen RecommendationSeverity = "green"
)

// Rule identifies which of the report's fixed recommendation rules
// fired. Values are stable, kebab-case machine strings carried by the
// report's --json output and used as an ordering tiebreak; they are
// contract, and renaming one changes both.
type Rule string

const (
	// RuleLCOM4Split is the red rule for a module whose LCOM4 is at
	// least [RuleLCOM4Min]: the module should be split.
	RuleLCOM4Split Rule = "lcom4-split"
	// RuleDistanceRefactor is the red rule for a module whose [metrics.Distance]
	// from the main sequence is at least [RuleDistanceMin].
	RuleDistanceRefactor Rule = "distance-refactor"
	// RuleCycleBreak is the red rule for a module that participates in
	// any circular dependency.
	RuleCycleBreak Rule = "cycle-break"
	// RuleDuplicationEliminate is the red, graph-level rule for an
	// overall duplication percentage of at least [RuleDuplicationMin].
	RuleDuplicationEliminate Rule = "duplication-eliminate"
	// RuleInstabilityIsolate is the yellow rule for a module whose
	// [metrics.Instability] is at least [RuleInstabilityHighMin].
	RuleInstabilityIsolate Rule = "instability-isolate"
	// RuleInstabilityProtect is the yellow rule for a module whose
	// instability is at most [RuleInstabilityLowMax].
	RuleInstabilityProtect Rule = "instability-protect"
	// RuleCeDecouple is the yellow rule for a module whose efferent
	// coupling is strictly greater than [RuleCeMin].
	RuleCeDecouple Rule = "ce-decouple"
	// RuleAbstractnessInterfaces is the yellow rule for a fully concrete
	// module (abstractness exactly [RuleAbstractnessMax]) with more than
	// [RuleExportedTypesMin] exported types.
	RuleAbstractnessInterfaces Rule = "abstractness-interfaces"
	// RuleDuplicationWatch is the green, graph-level rule for an overall
	// duplication percentage strictly between 0% and [RuleDuplicationMin].
	RuleDuplicationWatch Rule = "duplication-watch"
	// RuleAllClear is the green fallback emitted only when no rule of
	// any severity fires; it is then the single recommendation.
	RuleAllClear Rule = "all-clear"
)

// Recommendation rule thresholds. These bound the ten recommendation
// rules documented on [ComputeRecommendations]. Like the grade tables
// above, they are informational — they advise in reports and never gate
// enforcement, which remains [metrics.VerdictThresholds] territory. They are
// deliberately separate constants from the grade tables even where the
// values coincide ([RuleLCOM4Min] equals [LCOMGradeCPlus],
// [RuleDistanceMin] equals [DistanceGradeC], and [RuleDuplicationMin]
// equals [DuplicationGradeCPlus]): the rule table and the scorecard are
// different contracts that happen to agree today, and sharing constants
// would couple them silently.
//
// Bounds are inclusive unless marked otherwise; the two exclusive
// bounds ([RuleCeMin] and [RuleExportedTypesMin]) fire only strictly
// above their value, per the report spec's "Ce > 20" and "more than 10
// exported types" wording.
const (
	// RuleLCOM4Min is the inclusive minimum LCOM4 that fires the
	// lcom4-split rule. Unit: non-negative integer (connected
	// components).
	RuleLCOM4Min = 4
	// RuleDistanceMin is the inclusive minimum [metrics.Distance] that fires the
	// distance-refactor rule. Unit: dimensionless ratio in [0.0, 1.0].
	RuleDistanceMin = 0.5
	// RuleDuplicationMin is the inclusive minimum overall duplication
	// percentage that fires the duplication-eliminate rule; the green
	// duplication-watch rule fires strictly below it and strictly above
	// zero. Unit: percent in [0.0, 100.0].
	RuleDuplicationMin = 5.0
	// RuleInstabilityHighMin is the inclusive minimum [metrics.Instability] that
	// fires the instability-isolate rule. Unit: dimensionless ratio in
	// [0.0, 1.0].
	RuleInstabilityHighMin = 0.95
	// RuleInstabilityLowMax is the inclusive maximum instability that
	// fires the instability-protect rule. Unit: dimensionless ratio in
	// [0.0, 1.0].
	RuleInstabilityLowMax = 0.05
	// RuleCeMin is the EXCLUSIVE lower bound on efferent coupling for
	// the ce-decouple rule: it fires only for Ce strictly greater than
	// 20, so a module at exactly 20 does not fire. Unit: non-negative
	// integer (module dependencies).
	RuleCeMin = 20
	// RuleExportedTypesMin is the EXCLUSIVE lower bound on exported
	// types for the abstractness-interfaces rule, which additionally
	// requires abstractness of exactly zero. Unit: non-negative integer
	// (exported types).
	RuleExportedTypesMin = 10
	// RuleAbstractnessMax is the abstractness value the
	// abstractness-interfaces rule requires: exactly zero, fully
	// concrete. For schema-valid modules (ratio in [0.0, 1.0]) it is
	// both the minimum and the maximum of the firing set. Unit:
	// dimensionless ratio in [0.0, 1.0].
	RuleAbstractnessMax = 0.0
)

// RecommendationCap is the maximum number of recommendations
// [ComputeRecommendations] returns: after ordering, at most the five
// highest-priority recommendations survive. Unit: count of
// recommendations.
const RecommendationCap = 5

// Recommendation is one templated recommendation produced by the
// report's rule engine ([ComputeRecommendations]). Every field is a
// string, so the struct is comparable and tests pin it exactly.
//
// Message texts are contract — they reuse the vibe-check-reporter
// agent's recommendation templates verbatim — and rendering MUST NOT
// rewrite them.
type Recommendation struct {
	// Severity is the recommendation's severity tier: [SeverityRed],
	// [SeverityYellow], or [SeverityGreen]. Red and yellow are violation
	// severities; green is informational.
	Severity RecommendationSeverity `json:"severity"`
	// Rule identifies which fixed rule fired, e.g. [RuleLCOM4Split].
	Rule Rule `json:"rule"`
	// ModulePath is the path of the module a module-scoped rule names.
	// It is empty for the graph-scoped rules ([RuleDuplicationEliminate],
	// [RuleDuplicationWatch], [RuleAllClear]), whose messages name files
	// or nothing rather than a module.
	ModulePath string `json:"modulePath"`
	// Message is the human-readable, templated recommendation text.
	// Placeholders substitute as markdown code spans (backticked), with
	// %.2f ratios, %.0f similarity percents, and %d counts.
	Message string `json:"message"`
}

// ComputeRecommendations evaluates the report's fixed recommendation
// rule table over a [metrics.ModuleGraph] and returns the prioritized
// recommendations. It is a PURE, DETERMINISTIC function: no I/O, no
// globals, no error return, and identical inputs always yield identical
// output. The result is never empty — a graph where no rule fires
// yields the single green all-clear.
//
// The rule table has ten rules (thresholds are the [Rule...] constants;
// 🔴 and 🟡 hits are violation severities, 🟢 hits are informational):
//
//   - 🔴 [RuleLCOM4Split]: a module whose LCOM4 is at least
//     [RuleLCOM4Min] (per-module).
//   - 🔴 [RuleDistanceRefactor]: a module whose [metrics.Distance] is at least
//     [RuleDistanceMin] (per-module).
//   - 🔴 [RuleCycleBreak]: a module that participates in any circular
//     dependency (per-module; every member of every cycle fires).
//   - 🔴 [RuleDuplicationEliminate]: overall duplication
//     ([Aggregates].DuplicationPercent) of at least [RuleDuplicationMin]
//     (graph-level).
//   - 🟡 [RuleInstabilityIsolate]: a module whose [metrics.Instability] is at
//     least [RuleInstabilityHighMin] (per-module).
//   - 🟡 [RuleInstabilityProtect]: a module whose instability is at
//     most [RuleInstabilityLowMax] (per-module).
//   - 🟡 [RuleCeDecouple]: a module whose Ce is strictly greater than
//     [RuleCeMin] (per-module).
//   - 🟡 [RuleAbstractnessInterfaces]: a module whose abstractness is
//     exactly [RuleAbstractnessMax] with more than [RuleExportedTypesMin]
//     exported types (per-module; the exclusive exported-types bound
//     guards the zero-types case).
//   - 🟢 [RuleDuplicationWatch]: overall duplication strictly between
//     0% and [RuleDuplicationMin] (graph-level).
//   - 🟢 [RuleAllClear]: emitted only when no rule of any severity
//     fires — exactly one green recommendation, and nothing else.
//
// # Ordering contract
//
// Byte-stable report output depends on this exact order.
// Recommendations are sorted by:
//
//  1. Severity: red before yellow before green.
//  2. Rule-specific extremity, descending (most extreme first).
//  3. Module path, ascending; the empty path of graph-level rules sorts
//     first.
//  4. Rule identifier, ascending — disambiguates two rules firing for
//     the same module with equal extremity.
//  5. Message text, ascending — the final guard that makes the order
//     strict even for malformed graphs that repeat a module path: two
//     hits equal under every key are identical recommendations, so
//     [sort.Slice] stability cannot affect the output.
//
// The extremity measure is rule-specific — how far past its threshold a
// hit sits, with one count-based exception:
//
//   - [RuleLCOM4Split]: LCOM4 minus [RuleLCOM4Min] (connected
//     components past the split threshold).
//   - [RuleDistanceRefactor]: [metrics.Distance] minus [RuleDistanceMin].
//   - [RuleCycleBreak]: the number of cycles the module participates in
//     (more cycles, more entrenched).
//   - [RuleDuplicationEliminate]: DuplicationPercent minus
//     [RuleDuplicationMin] (percentage points).
//   - [RuleInstabilityIsolate]: [metrics.Instability] minus
//     [RuleInstabilityHighMin].
//   - [RuleInstabilityProtect]: [RuleInstabilityLowMax] minus
//     [metrics.Instability] (distance below the stable bound).
//   - [RuleCeDecouple]: Ce minus [RuleCeMin] (couplings past the bound).
//   - [RuleAbstractnessInterfaces]: ExportedTypes minus
//     [RuleExportedTypesMin] (types past the bound).
//   - [RuleDuplicationWatch]: the duplication percentage itself — the
//     closer to the eliminate bound, the earlier it renders.
//   - [RuleAllClear]: zero; it never competes because it renders alone.
//
// Within one severity, extremity values of different rules are compared
// numerically even though their units differ (components, ratios,
// counts, percentage points). That cross-rule comparison is heuristic,
// but it is fixed here as contract: changing any extremity definition
// changes report ordering.
//
// After ordering, at most [RecommendationCap] recommendations are
// returned, selected by this order.
//
// # Further behavior
//
//   - A nil graph yields the all-clear: no modules means no rule fires,
//     mirroring [ComputeAggregates]' nil handling.
//   - Input order independence: reordering the Modules or Cycles arrays
//     cannot change the result. Averages come from [ComputeAggregates]
//     (canonical summation); cycle messages name each module's
//     lexicographically smallest containing cycle; duplication facts
//     are set-valued and sorted; and the sort keys above form a strict
//     total order.
//   - Messages are contract: they reuse the vibe-check-reporter agent's
//     recommendation templates verbatim, substituting markdown code
//     spans for paths, values, zones, and file lists.
func ComputeRecommendations(g *metrics.ModuleGraph) []Recommendation {
	if g == nil {
		return []Recommendation{allClearRecommendation()}
	}
	agg := ComputeAggregates(g)

	var hits []ruleHit
	for _, m := range g.Modules {
		hits = append(hits, moduleRuleHits(m)...)
	}
	hits = append(hits, cycleRuleHits(g.Cycles)...)
	hits = append(hits, duplicationRuleHits(g.Modules, agg.DuplicationPercent)...)

	if len(hits) == 0 {
		return []Recommendation{allClearRecommendation()}
	}
	sort.Slice(hits, func(i, j int) bool { return hitLess(hits[i], hits[j]) })
	if len(hits) > RecommendationCap {
		hits = hits[:RecommendationCap]
	}
	recs := make([]Recommendation, len(hits))
	for i, hit := range hits {
		recs[i] = hit.rec
	}
	return recs
}

// ruleHit pairs a [Recommendation] with its rule-specific extremity —
// the sort key that orders recommendations within one severity. The
// extremity is internal: the report's JSON model carries severity and
// rule, not the ordering metric.
type ruleHit struct {
	rec       Recommendation
	extremity float64
}

// hitLess reports whether hit a orders before hit b, implementing the
// ordering contract documented on [ComputeRecommendations]: severity
// rank ascending (red, yellow, green), extremity descending, module
// path ascending, rule identifier ascending, message ascending. Two
// hits that compare equal under all five keys are identical
// recommendations, so [sort.Slice] stability cannot affect the output.
func hitLess(a, b ruleHit) bool {
	if a.rec.Severity != b.rec.Severity {
		return severityRank(a.rec.Severity) < severityRank(b.rec.Severity)
	}
	if a.extremity != b.extremity {
		return a.extremity > b.extremity
	}
	if a.rec.ModulePath != b.rec.ModulePath {
		return a.rec.ModulePath < b.rec.ModulePath
	}
	if a.rec.Rule != b.rec.Rule {
		return a.rec.Rule < b.rec.Rule
	}
	return a.rec.Message < b.rec.Message
}

// severityRank maps a severity to its ordering rank: red 0, yellow 1,
// green 2. Unknown severities rank last; the rule engine never produces
// them, so the case exists only to keep the ordering total.
func severityRank(s RecommendationSeverity) int {
	switch s {
	case SeverityRed:
		return 0
	case SeverityYellow:
		return 1
	case SeverityGreen:
		return 2
	default:
		return 3
	}
}

// moduleRuleHits evaluates the six module-scoped rules against one
// [metrics.ModuleResult] and returns the hits that fired (nil when none do).
// Extremity definitions per rule are documented on
// [ComputeRecommendations]; messages are the fixed templates.
func moduleRuleHits(m metrics.ModuleResult) []ruleHit {
	var hits []ruleHit
	if m.LCOM >= RuleLCOM4Min {
		hits = append(hits, newRuleHit(SeverityRed, RuleLCOM4Split, m.Path,
			fmt.Sprintf("Split `%s` — LCOM4 of `%d` suggests multiple unrelated responsibilities", m.Path, m.LCOM),
			float64(m.LCOM-RuleLCOM4Min)))
	}
	if m.Distance >= RuleDistanceMin {
		hits = append(hits, newRuleHit(SeverityRed, RuleDistanceRefactor, m.Path,
			fmt.Sprintf("Refactor `%s` — Distance of `%.2f` from Main Sequence places it in `%s`", m.Path, m.Distance, m.Zone),
			float64(m.Distance-RuleDistanceMin)))
	}
	if m.Instability >= RuleInstabilityHighMin {
		hits = append(hits, newRuleHit(SeverityYellow, RuleInstabilityIsolate, m.Path,
			fmt.Sprintf("Isolate `%s` — Instability of `%.2f`; depends on everything, nothing depends on it; appropriate for CLI layers", m.Path, m.Instability),
			float64(m.Instability-RuleInstabilityHighMin)))
	}
	if m.Instability <= RuleInstabilityLowMax {
		hits = append(hits, newRuleHit(SeverityYellow, RuleInstabilityProtect, m.Path,
			fmt.Sprintf("Protect `%s` — Instability of `%.2f` makes it maximally stable; changes ripple widely", m.Path, m.Instability),
			float64(RuleInstabilityLowMax-m.Instability)))
	}
	if m.Ce > RuleCeMin {
		hits = append(hits, newRuleHit(SeverityYellow, RuleCeDecouple, m.Path,
			fmt.Sprintf("Decouple `%s` — `%d` efferent couplings; consider interface extraction", m.Path, m.Ce),
			float64(m.Ce-RuleCeMin)))
	}
	// The exported-types bound is exclusive and the abstractness bound
	// is exactly zero, so a fully concrete module with few exported
	// types — including the zero-types case — does not fire.
	if m.Abstractness == RuleAbstractnessMax && m.ExportedTypes > RuleExportedTypesMin {
		hits = append(hits, newRuleHit(SeverityYellow, RuleAbstractnessInterfaces, m.Path,
			fmt.Sprintf("Add interfaces to `%s` — 0 abstractness with `%d` exported types leaves no room for abstraction", m.Path, m.ExportedTypes),
			float64(m.ExportedTypes-RuleExportedTypesMin)))
	}
	return hits
}

// cycleRuleHits evaluates the cycle-membership rule: every module that
// participates in any circular dependency fires one red hit. The
// message names the module's representative cycle — the
// lexicographically smallest cycle containing it, by member sequence —
// so the choice cannot depend on the order cycles appear in the graph.
// Extremity is the number of cycles the module participates in.
func cycleRuleHits(cycles []metrics.Cycle) []ruleHit {
	if len(cycles) == 0 {
		return nil
	}
	ordered := append([]metrics.Cycle(nil), cycles...)
	sort.Slice(ordered, func(i, j int) bool { return cycleLess(ordered[i], ordered[j]) })

	participation := make(map[string]int)
	representative := make(map[string]metrics.Cycle)
	for _, c := range ordered {
		for _, path := range c {
			participation[path]++
			if _, ok := representative[path]; !ok {
				representative[path] = c
			}
		}
	}

	paths := make([]string, 0, len(participation))
	for path := range participation {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var hits []ruleHit
	for _, path := range paths {
		hits = append(hits, newRuleHit(SeverityRed, RuleCycleBreak, path,
			fmt.Sprintf("Break cycle between %s — circular dependencies prevent independent testing", backtickList(representative[path])),
			float64(participation[path])))
	}
	return hits
}

// cycleLess orders cycles lexicographically by member sequence, with
// the shorter prefix ordering first when one cycle's members are a
// prefix of another's. It relies on each cycle's members being sorted,
// which the [metrics.Cycle] contract guarantees.
func cycleLess(a, b metrics.Cycle) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}

// duplicationRuleHits evaluates the two graph-scoped duplication rules
// against the overall duplication percentage computed by
// [ComputeAggregates]. At or above [RuleDuplicationMin] the red
// eliminate rule fires; strictly between zero and the bound the green
// watch rule fires; at exactly zero neither fires. Extremity is
// percentage points past the bound for the eliminate rule and the
// percentage itself for the watch rule.
func duplicationRuleHits(mods []metrics.ModuleResult, pct float64) []ruleHit {
	if pct <= 0 {
		return nil
	}
	blocks, files, maxSimilarity := duplicationFacts(mods)
	if pct >= RuleDuplicationMin {
		return []ruleHit{newRuleHit(SeverityRed, RuleDuplicationEliminate, "",
			fmt.Sprintf("Eliminate duplicate blocks — `%d` blocks at `%.0f`%% similarity in %s", blocks, maxSimilarity*100, backtickList(files)),
			pct-RuleDuplicationMin)}
	}
	return []ruleHit{newRuleHit(SeverityGreen, RuleDuplicationWatch, "",
		fmt.Sprintf("Watch duplication in %s — `%d` blocks at `%.0f`%% similarity", backtickList(files), blocks, maxSimilarity*100),
		pct)}
}

// duplicationFacts collects the graph-level duplication facts the two
// duplication rules cite in their messages: the number of deduplicated
// blocks (by [dupBlockKey] identity, the same dedup [duplicationPercent]
// applies), the sorted list of unique file paths containing those
// blocks, and the highest [Duplication.Similarity] among all
// duplication groups. Set and maximum semantics make every value
// independent of the order modules and duplications appear in.
func duplicationFacts(mods []metrics.ModuleResult) (blocks int, files []string, maxSimilarity float64) {
	seen := make(map[dupBlockKey]bool)
	fileSet := make(map[string]bool)
	for _, m := range mods {
		for _, d := range m.Duplications {
			if d.Similarity > maxSimilarity {
				maxSimilarity = d.Similarity
			}
			for _, b := range d.Blocks {
				k := dupBlockKey{file: b.File, startLine: b.StartLine, endLine: b.EndLine}
				if seen[k] {
					continue
				}
				seen[k] = true
				blocks++
				fileSet[b.File] = true
			}
		}
	}
	files = make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}
	sort.Strings(files)
	return blocks, files, maxSimilarity
}

// allClearRecommendation returns the single green all-clear
// recommendation, emitted only when no rule of any severity fires.
func allClearRecommendation() Recommendation {
	return Recommendation{
		Severity:   SeverityGreen,
		Rule:       RuleAllClear,
		ModulePath: "",
		Message:    "Architecture is healthy — no significant issues detected; keep monitoring",
	}
}

// newRuleHit builds a [ruleHit] from its recommendation parts.
func newRuleHit(severity RecommendationSeverity, rule Rule, modulePath, message string, extremity float64) ruleHit {
	return ruleHit{
		rec: Recommendation{
			Severity:   severity,
			Rule:       rule,
			ModulePath: modulePath,
			Message:    message,
		},
		extremity: extremity,
	}
}

// backtickList renders each item as a markdown code span and joins the
// spans with ", ". The recommendation templates' plural placeholders —
// cycle membership and duplication file lists — substitute this way.
func backtickList(items []string) string {
	spans := make([]string, len(items))
	for i, item := range items {
		spans[i] = "`" + item + "`"
	}
	return strings.Join(spans, ", ")
}
