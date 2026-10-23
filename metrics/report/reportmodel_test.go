package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// TestBuildReportModel pins the model builder's wiring: aggregates come
// from [ComputeAggregates], each dimension's grade is the literal its
// aggregate value classifies into, status is the [ReportStatus] word for
// the passed recommendations, and the recommendations pass through
// unchanged. The nil graph yields the empty-graph model: zero aggregates,
// every dimension graded A, the all-clear status.
func TestBuildReportModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		graph      *metrics.ModuleGraph
		wantGrades ReportGrades
		wantStatus string
		wantRecs   int
	}{
		{
			name:  "clean graph",
			graph: mustGraph(t, metrics.StatusComplete, cleanMods(t), nil, nil),
			wantGrades: ReportGrades{
				AverageDistance:    GradeB, // mean distance 0.25
				MaxLCOM4:           GradeBPlus,
				CycleCount:         GradeA,
				DuplicationPercent: GradeA,
				InstabilitySpread:  GradeA,
			},
			wantStatus: StatusAllClear,
			wantRecs:   1, // the green all-clear
		},
		{
			name:  "violations graph",
			graph: mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil),
			wantGrades: ReportGrades{
				AverageDistance:    GradeB, // mean distance 0.25
				MaxLCOM4:           GradeC, // monolith LCOM4 5
				CycleCount:         GradeA,
				DuplicationPercent: GradeA,
				InstabilitySpread:  GradeA,
			},
			wantStatus: StatusNeedsAttention,
			wantRecs:   2, // red lcom4-split, yellow ce-decouple
		},
		{
			name:  "nil graph",
			graph: nil,
			wantGrades: ReportGrades{ // zero aggregates grade every dimension A
				AverageDistance:    GradeA,
				MaxLCOM4:           GradeA,
				CycleCount:         GradeA,
				DuplicationPercent: GradeA,
				InstabilitySpread:  GradeA,
			},
			wantStatus: StatusAllClear,
			wantRecs:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recs := ComputeRecommendations(tt.graph)
			model := BuildReportModel(tt.graph, recs)

			if want := ComputeAggregates(tt.graph); model.Aggregates != want {
				t.Errorf("Aggregates: got %+v, want %+v", model.Aggregates, want)
			}
			if model.Grades != tt.wantGrades {
				t.Errorf("Grades: got %+v, want %+v", model.Grades, tt.wantGrades)
			}
			if model.Status != tt.wantStatus {
				t.Errorf("Status: got %q, want %q", model.Status, tt.wantStatus)
			}
			if len(model.Recommendations) != tt.wantRecs {
				t.Fatalf("Recommendations: got %d entries, want %d", len(model.Recommendations), tt.wantRecs)
			}
			for i, r := range model.Recommendations {
				if r != recs[i] {
					t.Errorf("Recommendations[%d]: got %+v, want the passed recommendation %+v", i, r, recs[i])
				}
			}
		})
	}
}

// TestReportModel_JSONShape pins the model's serialization contract: the
// marshaled payload is exactly the four top-level keys in declaration order
// (aggregates, grades, status, recommendations), every key is camelCase,
// severity serializes as the machine-readable strings red/yellow/green
// rather than emoji, and grades serialize as their letter strings.
func TestReportModel_JSONShape(t *testing.T) {
	t.Parallel()

	t.Run("exact bytes", func(t *testing.T) {
		t.Parallel()

		// A single module whose means are exact in binary floating point,
		// so the marshaled numbers pin exactly: instability 0.5, distance
		// 0.25, LCOM4 1 — no rule fires, so the single green all-clear is
		// the model's only recommendation.
		g := mustGraph(t, metrics.StatusComplete,
			[]metrics.ModuleResult{mod("ex/a", 1, 1, 0.5, 0.25, 0.25, 1)}, nil, nil)
		data, err := json.MarshalIndent(BuildReportModel(g, ComputeRecommendations(g)), "", "  ")
		if err != nil {
			t.Fatalf("marshal model: %v", err)
		}
		want := `{
  "aggregates": {
    "averageInstability": 0.5,
    "averageDistance": 0.25,
    "maxLCOM4": 1,
    "cycleCount": 0,
    "instabilitySpread": 0,
    "duplicationPercent": 0
  },
  "grades": {
    "averageDistance": "B",
    "maxLCOM4": "A",
    "cycleCount": "A",
    "duplicationPercent": "A",
    "instabilitySpread": "A"
  },
  "status": "All Clear",
  "recommendations": [
    {
      "severity": "green",
      "rule": "all-clear",
      "modulePath": "",
      "message": "Architecture is healthy — no significant issues detected; keep monitoring"
    }
  ]
}`
		if got := string(data); got != want {
			t.Errorf("marshaled model:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("machine-readable severity", func(t *testing.T) {
		t.Parallel()

		// The violations fixture fires one red and one yellow
		// recommendation, so its payload must carry both machine-readable
		// severity strings and none of the report's severity emoji.
		g := mustGraph(t, metrics.StatusComplete, violationMods(t), nil, nil)
		data, err := json.MarshalIndent(BuildReportModel(g, ComputeRecommendations(g)), "", "  ")
		if err != nil {
			t.Fatalf("marshal model: %v", err)
		}
		got := string(data)
		for _, want := range []string{`"severity": "red"`, `"severity": "yellow"`} {
			if !strings.Contains(got, want) {
				t.Errorf("payload missing machine-readable severity %s:\n%s", want, got)
			}
		}
		for _, emoji := range []string{"🔴", "🟡", "🟢"} {
			if strings.Contains(got, emoji) {
				t.Errorf("payload must not carry severity emoji %s:\n%s", emoji, got)
			}
		}
	})
}

// marshalTestReport marshals the report model of g (with the graph's own
// computed recommendations) with the same two-space indentation the report
// command uses, so order-independence comparisons run against the exact
// bytes consumers see.
func marshalTestReport(t *testing.T, g *metrics.ModuleGraph) string {
	t.Helper()
	data, err := json.MarshalIndent(BuildReportModel(g, ComputeRecommendations(g)), "", "  ")
	if err != nil {
		t.Fatalf("marshal model: %v", err)
	}
	return string(data)
}

// TestBuildReportModel_InputOrderIndependence pins the model-level face of
// the byte-stability requirement: reordering every array of the snapshot
// the model consumes cannot change the marshaled payload, because the
// aggregates use canonical summation, the recommendations are
// order-independent, and the grades derive from the aggregates.
func TestBuildReportModel_InputOrderIndependence(t *testing.T) {
	t.Parallel()

	g := detailedFixture(t)
	want := marshalTestReport(t, g)
	for _, p := range graphPermutations {
		if got := marshalTestReport(t, permuteGraph(g, p.perm)); got != want {
			t.Errorf("%s arrays changed the marshaled model:\nwant:\n%s\ngot:\n%s", p.name, want, got)
		}
	}
}
