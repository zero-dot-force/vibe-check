package report

import "github.com/zero-dot-force/vibe-check/metrics"

// ReportGrades holds the health scorecard's per-dimension grades: the
// [Grade] each of the five graded dimensions of a report classifies into.
// Field names and JSON keys mirror the graded fields of [Aggregates]
// one-for-one, so a consumer of the JSON model can join
// grades.<dimension> onto aggregates.<dimension> mechanically.
//
// Like every grade surface, these are informational only — they advise,
// they never gate. The entropy-regression thresholds in [metrics.VerdictThresholds]
// (applied by [metrics.DecideVerdict]) remain the only enforcement thresholds.
type ReportGrades struct { //nolint:revive // stutter accepted: name kept stable across the metrics-to-report move so callers change only their import path
	// AverageDistance grades the arithmetic mean of module distances from
	// the main sequence: [ClassifyDistance] over
	// [Aggregates].AverageDistance.
	AverageDistance Grade `json:"averageDistance"`
	// MaxLCOM4 grades the maximum module LCOM4: [ClassifyLCOM4] over
	// [Aggregates].MaxLCOM4.
	MaxLCOM4 Grade `json:"maxLCOM4"`
	// CycleCount grades the circular dependency count: [ClassifyCycles]
	// over [Aggregates].CycleCount.
	CycleCount Grade `json:"cycleCount"`
	// DuplicationPercent grades the overall duplication percentage:
	// [ClassifyDuplication] over [Aggregates].DuplicationPercent.
	DuplicationPercent Grade `json:"duplicationPercent"`
	// InstabilitySpread grades the count of extreme-instability modules:
	// [ClassifySpread] over [Aggregates].InstabilitySpread.
	InstabilitySpread Grade `json:"instabilitySpread"`
}

// ReportModel is the computed report model: the report's full data content —
// aggregates, per-dimension grades, overall status, and fired
// recommendations — assembled by [BuildReportModel] before any rendering.
// `vibe-check report --json` serializes it as a single JSON object so
// machine consumers (drift tracking, trend agents) get a stable contract
// that does not depend on parsing markdown.
//
// The struct is marshaled directly — never through a map — so field
// declaration order determines JSON key order and every key is camelCase,
// making serialization byte-stable across runs for identical input.
// Recommendation severity serializes as the machine-readable strings
// "red", "yellow", and "green" ([RecommendationSeverity]), never emoji;
// grades serialize as their letter strings ("A" … "F").
type ReportModel struct { //nolint:revive // stutter accepted: name kept stable across the metrics-to-report move so callers change only their import path
	// Aggregates carries the aggregate metric values the status summary
	// line and the scorecard dimensions derive from.
	Aggregates Aggregates `json:"aggregates"`
	// Grades carries the health scorecard's per-dimension grades.
	Grades ReportGrades `json:"grades"`
	// Status is the overall status word: [StatusAllClear] when no
	// violation-severity recommendation fires, [StatusNeedsAttention]
	// otherwise.
	Status string `json:"status"`
	// Recommendations lists the fired recommendations — ordered and capped
	// by [ComputeRecommendations] — each carrying its machine-readable
	// severity and rule.
	Recommendations []Recommendation `json:"recommendations"`
}

// BuildReportModel assembles the [ReportModel] for a snapshot from its
// already-computed recommendations: aggregates via [ComputeAggregates],
// per-dimension grades via the Classify* functions over those aggregates,
// and the overall status via [ReportStatus]. It is a PURE, DETERMINISTIC
// function: no I/O, no globals, no error return, and identical inputs
// always yield identical output.
//
// Callers pass the recommendations [ComputeRecommendations] produced for
// the same graph — the reuse pattern [ReportStatus] documents — so the
// model, the exit code, and any rendered report agree without evaluating
// the rule engine twice. A nil graph computes zero aggregates (every
// dimension grades A); the recommendations and status derive from recs
// exactly as passed.
func BuildReportModel(g *metrics.ModuleGraph, recs []Recommendation) ReportModel {
	agg := ComputeAggregates(g)
	return ReportModel{
		Aggregates: agg,
		Grades: ReportGrades{
			AverageDistance:    ClassifyDistance(agg.AverageDistance),
			MaxLCOM4:           ClassifyLCOM4(agg.MaxLCOM4),
			CycleCount:         ClassifyCycles(agg.CycleCount),
			DuplicationPercent: ClassifyDuplication(agg.DuplicationPercent),
			InstabilitySpread:  ClassifySpread(agg.InstabilitySpread),
		},
		Status:          ReportStatus(recs),
		Recommendations: recs,
	}
}
