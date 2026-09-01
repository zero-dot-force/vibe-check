package metrics

import (
	"reflect"
	"strings"
	"testing"
)

func TestDefaultVerdictThresholds_ProtectedDefaults(t *testing.T) {
	t.Parallel()

	got := DefaultVerdictThresholds()
	want := VerdictThresholds{MaxInstabilityDelta: 0.15, MaxDistanceDelta: 0.20, MaxLCOMDelta: 2}
	if got != want {
		t.Errorf("DefaultVerdictThresholds: got %+v, want %+v", got, want)
	}
}

func TestRoundTo4_HalfAwayFromZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "rounds_up_0_14999", in: 0.14999, want: 0.1500},
		{name: "rounds_down_0_14994", in: 0.14994, want: 0.1499},
		{name: "distance_0_19999", in: 0.19999, want: 0.2000},
		{name: "half_away_up", in: 0.00005, want: 0.0001},
		{name: "negative_half_away", in: -0.00005, want: -0.0001},
		{name: "exact_zero", in: 0.0, want: 0.0},
		{name: "negative_improvement", in: -0.20000000000000001, want: -0.2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := roundTo4(tt.in); got != tt.want {
				t.Errorf("roundTo4(%v): got %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestDecideVerdict_EndToEnd exercises the full ComputeDelta → DecideVerdict path
// over schema-valid fixtures, asserting both the verdict and the reporting-only
// entropy direction at every gate boundary (guarding the inclusive ≥ comparison
// and the 4-decimal round-half-away-from-zero rule).
func TestDecideVerdict_EndToEnd(t *testing.T) {
	t.Parallel()

	cyclePair := func(baseCycles, prCycles []Cycle) (*ModuleGraph, *ModuleGraph) {
		mods := []ModuleResult{mod("pkg/a", 1, 1, 0.0, 0.0, 0.0, 0), mod("pkg/b", 1, 1, 0.0, 0.0, 0.0, 0)}
		return mustGraph(t, StatusComplete, mods, baseCycles, nil),
			mustGraph(t, StatusComplete, mods, prCycles, nil)
	}
	baseNoCycle, prNewCycle := cyclePair(nil, []Cycle{{"pkg/a", "pkg/b"}})
	basePreCycle, prPreCycle := cyclePair([]Cycle{{"pkg/a", "pkg/b"}}, []Cycle{{"pkg/b", "pkg/a"}})

	tests := []struct {
		name          string
		base          *ModuleGraph
		pr            *ModuleGraph
		wantVerdict   Verdict
		wantDirection EntropyDirection
		wantReason    string
	}{
		{
			name:          "instability_below_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.14, 0, 0, 0)),
			wantVerdict:   VerdictComment,
			wantDirection: EntropyStable,
			wantReason:    "materiality",
		},
		{
			name:          "instability_at_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.15, 0, 0, 0)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "instability",
		},
		{
			name:          "instability_rounds_up_to_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.14999, 0, 0, 0)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "instability",
		},
		{
			name:          "instability_rounds_down_below_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.14994, 0, 0, 0)),
			wantVerdict:   VerdictComment,
			wantDirection: EntropyStable,
			wantReason:    "materiality",
		},
		{
			name:          "distance_below_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0, 0, 0.00, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0, 0, 0.19, 0)),
			wantVerdict:   VerdictComment,
			wantDirection: EntropyStable,
			wantReason:    "materiality",
		},
		{
			name:          "distance_at_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0, 0, 0.00, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0, 0, 0.20, 0)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "distance",
		},
		{
			name:          "distance_rounds_up_to_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0, 0, 0.00, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0, 0, 0.19999, 0)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "distance",
		},
		{
			name:          "lcom_below_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0, 0, 0, 1)),
			wantVerdict:   VerdictComment,
			wantDirection: EntropyStable,
			wantReason:    "materiality",
		},
		{
			name:          "lcom_at_threshold",
			base:          completeGraph(t, mod("m", 1, 1, 0, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0, 0, 0, 2)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "lcom",
		},
		{
			name:          "comment_band_small_instability",
			base:          completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.05, 0, 0, 0)),
			wantVerdict:   VerdictComment,
			wantDirection: EntropyStable,
			wantReason:    "materiality",
		},
		{
			name:          "improvement",
			base:          completeGraph(t, mod("m", 1, 1, 0.20, 0, 0.20, 2)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.00, 0, 0.00, 0)),
			wantVerdict:   VerdictApprove,
			wantDirection: EntropyImproving,
		},
		{
			name:          "no_change",
			base:          completeGraph(t, mod("m", 1, 1, 0.30, 0.10, 0.40, 1)),
			pr:            completeGraph(t, mod("m", 1, 1, 0.30, 0.10, 0.40, 1)),
			wantVerdict:   VerdictApprove,
			wantDirection: EntropyStable,
		},
		{
			name:          "empty_vs_empty",
			base:          completeGraph(t),
			pr:            completeGraph(t),
			wantVerdict:   VerdictApprove,
			wantDirection: EntropyStable,
		},
		{
			name:          "added_removed_do_not_gate",
			base:          completeGraph(t, mod("m/x", 1, 1, 0, 0, 0, 0)),
			pr:            completeGraph(t, mod("m/y", 1, 1, 0, 0, 0, 0)),
			wantVerdict:   VerdictApprove,
			wantDirection: EntropyStable,
		},
		{
			name:          "spec_instability_plus_0_15",
			base:          completeGraph(t, mod("svc", 1, 1, 0.40, 0, 0, 0)),
			pr:            completeGraph(t, mod("svc", 1, 1, 0.55, 0, 0, 0)),
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "instability",
		},
		{
			name:          "new_cycle_forces_request_changes",
			base:          baseNoCycle,
			pr:            prNewCycle,
			wantVerdict:   VerdictRequestChanges,
			wantDirection: EntropyDegrading,
			wantReason:    "new-cycle",
		},
		{
			name:          "pre_existing_cycle_approves",
			base:          basePreCycle,
			pr:            prPreCycle,
			wantVerdict:   VerdictApprove,
			wantDirection: EntropyStable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gd := ComputeDelta(tt.base, tt.pr)
			verdict, reasons := DecideVerdict(gd, DefaultVerdictThresholds())

			if verdict != tt.wantVerdict {
				t.Errorf("verdict: got %v, want %v (reasons: %v)", verdict, tt.wantVerdict, reasons)
			}
			if gd.Direction != tt.wantDirection {
				t.Errorf("direction: got %v, want %v", gd.Direction, tt.wantDirection)
			}
			if tt.wantReason != "" {
				found := false
				for _, r := range reasons {
					if strings.Contains(r, tt.wantReason) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("reasons %v do not contain %q", reasons, tt.wantReason)
				}
			}
			if tt.wantVerdict == VerdictApprove && len(reasons) != 0 {
				t.Errorf("APPROVE must carry no reasons, got %v", reasons)
			}
		})
	}
}

func TestDecideVerdict_UnreliableForcesComment(t *testing.T) {
	t.Parallel()

	// Direct zero-plus-flag path.
	verdict, reasons := DecideVerdict(GraphDelta{Unreliable: true}, DefaultVerdictThresholds())
	if verdict != VerdictComment {
		t.Errorf("verdict: got %v, want %v", verdict, VerdictComment)
	}
	if len(reasons) != 1 || !strings.Contains(reasons[0], "partial-build") {
		t.Errorf("reasons: got %v, want a single partial-build reason", reasons)
	}

	// End-to-end path: a partial input degrades even a would-be REQUEST_CHANGES
	// regression down to COMMENT (never APPROVE).
	base := completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0))
	pr := mustGraph(t, StatusPartial,
		[]ModuleResult{mod("m", 1, 1, 0.90, 0, 0, 0)}, nil,
		[]Warning{{Code: "load-error", Message: "boom", Module: "m"}})

	gd := ComputeDelta(base, pr)
	verdict, reasons = DecideVerdict(gd, DefaultVerdictThresholds())
	if verdict != VerdictComment {
		t.Errorf("partial-build verdict: got %v, want %v", verdict, VerdictComment)
	}
	if len(reasons) != 1 || !strings.Contains(reasons[0], "partial-build") {
		t.Errorf("partial-build reasons: got %v, want a single partial-build reason", reasons)
	}
}

func TestDecideVerdict_ZeroValueApproves(t *testing.T) {
	t.Parallel()

	verdict, reasons := DecideVerdict(GraphDelta{}, DefaultVerdictThresholds())
	if verdict != VerdictApprove {
		t.Errorf("verdict: got %v, want %v", verdict, VerdictApprove)
	}
	if len(reasons) != 0 {
		t.Errorf("reasons: got %v, want none", reasons)
	}
}

func TestDecideVerdict_NilComputeDeltaYieldsComment(t *testing.T) {
	t.Parallel()

	pr := completeGraph(t, mod("m", 1, 1, 0.0, 0.0, 0.0, 0))

	tests := []struct {
		name string
		base *ModuleGraph
		pr   *ModuleGraph
	}{
		{name: "nil_base", base: nil, pr: pr},
		{name: "nil_pr", base: pr, pr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gd := ComputeDelta(tt.base, tt.pr)
			verdict, reasons := DecideVerdict(gd, DefaultVerdictThresholds())
			if verdict != VerdictComment {
				t.Errorf("verdict: got %v, want %v", verdict, VerdictComment)
			}
			if len(reasons) != 1 || !strings.Contains(reasons[0], "partial-build") {
				t.Errorf("reasons: got %v, want a single partial-build reason", reasons)
			}
		})
	}
}

func TestDecideVerdict_ReasonsStableSorted(t *testing.T) {
	t.Parallel()

	// Two modules cross gates and a new cycle is introduced. Reasons MUST be
	// emitted deterministically: new-cycle first, then modules by path, and
	// within a module ordered instability, distance, LCOM.
	base := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 1, 1, 0.00, 0, 0.00, 0),
			mod("m/b", 1, 1, 0.00, 0, 0.00, 0),
		}, nil, nil)
	pr := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 1, 1, 0.20, 0, 0.00, 0),
			mod("m/b", 1, 1, 0.00, 0, 0.30, 3),
		},
		[]Cycle{{"z", "y"}}, nil)

	gd := ComputeDelta(base, pr)
	verdict, reasons := DecideVerdict(gd, DefaultVerdictThresholds())

	if verdict != VerdictRequestChanges {
		t.Fatalf("verdict: got %v, want %v", verdict, VerdictRequestChanges)
	}
	want := []string{
		"new-cycle: y z",
		`instability: module "m/a" delta 0.2000 >= threshold 0.1500`,
		`distance: module "m/b" delta 0.3000 >= threshold 0.2000`,
		`lcom: module "m/b" delta 3 >= threshold 2`,
	}
	if !reflect.DeepEqual(reasons, want) {
		t.Errorf("reasons:\n got %#v\nwant %#v", reasons, want)
	}
}

func TestDecideVerdict_Deterministic(t *testing.T) {
	t.Parallel()

	base := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 1, 1, 0.10, 0, 0.10, 1),
			mod("m/b", 1, 1, 0.00, 0, 0.00, 0),
		},
		[]Cycle{{"p", "q"}}, nil)
	pr := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 2, 2, 0.40, 0, 0.35, 4),
			mod("m/b", 1, 1, 0.00, 0, 0.00, 0),
		},
		[]Cycle{{"r", "s"}}, nil)

	gd1 := ComputeDelta(base, pr)
	v1, r1 := DecideVerdict(gd1, DefaultVerdictThresholds())
	gd2 := ComputeDelta(base, pr)
	v2, r2 := DecideVerdict(gd2, DefaultVerdictThresholds())

	if v1 != v2 {
		t.Errorf("verdict not deterministic: got %v and %v", v1, v2)
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Errorf("reasons not deterministic:\n got %#v\nand %#v", r1, r2)
	}
	if !reflect.DeepEqual(gd1, gd2) {
		t.Errorf("delta not deterministic:\n got %+v\nand %+v", gd1, gd2)
	}
}

func TestDecideVerdict_TightenedThreshold(t *testing.T) {
	t.Parallel()

	// A ΔInstability of 0.10 passes the default 0.15 gate (COMMENT) but a
	// tightened 0.10 gate turns it into REQUEST_CHANGES, proving the thresholds
	// argument is honored.
	base := completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0))
	pr := completeGraph(t, mod("m", 1, 1, 0.10, 0, 0, 0))
	gd := ComputeDelta(base, pr)

	if verdict, _ := DecideVerdict(gd, DefaultVerdictThresholds()); verdict != VerdictComment {
		t.Errorf("default threshold verdict: got %v, want %v", verdict, VerdictComment)
	}

	tightened := VerdictThresholds{MaxInstabilityDelta: 0.10, MaxDistanceDelta: 0.20, MaxLCOMDelta: 2}
	verdict, reasons := DecideVerdict(gd, tightened)
	if verdict != VerdictRequestChanges {
		t.Errorf("tightened threshold verdict: got %v, want %v", verdict, VerdictRequestChanges)
	}
	found := false
	for _, r := range reasons {
		if strings.Contains(r, "instability") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("tightened reasons %v do not name the instability gate", reasons)
	}
}
