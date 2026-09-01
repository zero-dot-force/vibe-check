package metrics

import (
	"encoding/json"
	"reflect"
	"testing"
)

// mod builds a ModuleResult with explicit metric values. Fixtures set the metric
// fields directly to exercise delta boundaries; they are NOT required to satisfy
// the D = |A + I - 1| identity, which Validate does not enforce — Validate checks
// only that ratio metrics fall in [0.0, 1.0] and that counts are non-negative.
func mod(path string, ca, ce int, inst, abst, dist float64, lcom int) ModuleResult {
	return ModuleResult{
		Module: Module{
			Path:          path,
			Name:          path,
			Ca:            ca,
			Ce:            ce,
			ExportedTypes: 1,
			AbstractTypes: 0,
		},
		Instability:  Instability(inst),
		Abstractness: Abstractness(abst),
		Distance:     Distance(dist),
		LCOM:         LCOM(lcom),
		Zone:         ZoneNormal,
	}
}

// mustGraph builds a schema-valid ModuleGraph fixture and fails the test if it
// does not pass metrics.Validate, so every fixture is validated at construction.
func mustGraph(t *testing.T, status Status, mods []ModuleResult, cycles []Cycle, warnings []Warning) *ModuleGraph {
	t.Helper()
	if mods == nil {
		mods = []ModuleResult{}
	}
	if cycles == nil {
		cycles = []Cycle{}
	}
	if warnings == nil {
		warnings = []Warning{}
	}
	g := &ModuleGraph{
		SchemaVersion: "1.1",
		Language:      "go",
		Modules:       mods,
		Cycles:        cycles,
		Warnings:      warnings,
		Status:        status,
	}
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := Validate(data); err != nil {
		t.Fatalf("fixture failed schema validation: %v", err)
	}
	return g
}

// completeGraph builds a schema-valid, complete-status graph from the given
// modules with no cycles and no warnings.
func completeGraph(t *testing.T, mods ...ModuleResult) *ModuleGraph {
	t.Helper()
	return mustGraph(t, StatusComplete, mods, nil, nil)
}

func TestComputeDelta_MatchedModuleDeltas(t *testing.T) {
	t.Parallel()

	base := completeGraph(t,
		mod("m/a", 0, 1, 0.40, 0.10, 0.30, 2),
		mod("m/b", 2, 0, 0.00, 0.50, 0.50, 1),
	)
	pr := completeGraph(t,
		mod("m/a", 1, 2, 0.55, 0.20, 0.45, 4),
		mod("m/b", 2, 0, 0.00, 0.50, 0.50, 1),
	)

	gd := ComputeDelta(base, pr)

	if got, want := len(gd.Modules), 2; got != want {
		t.Fatalf("len(Modules): got %v, want %v", got, want)
	}
	// Stable sort by Path.
	if got, want := gd.Modules[0].Path, "m/a"; got != want {
		t.Errorf("Modules[0].Path: got %v, want %v", got, want)
	}
	if got, want := gd.Modules[1].Path, "m/b"; got != want {
		t.Errorf("Modules[1].Path: got %v, want %v", got, want)
	}

	a := gd.Modules[0]
	if got, want := a.Ca, 1; got != want {
		t.Errorf("m/a ΔCa: got %v, want %v", got, want)
	}
	if got, want := a.Ce, 1; got != want {
		t.Errorf("m/a ΔCe: got %v, want %v", got, want)
	}
	if got, want := roundTo4(a.Instability), 0.15; got != want {
		t.Errorf("m/a ΔInstability: got %v, want %v", got, want)
	}
	if got, want := roundTo4(a.Abstractness), 0.10; got != want {
		t.Errorf("m/a ΔAbstractness: got %v, want %v", got, want)
	}
	if got, want := roundTo4(a.Distance), 0.15; got != want {
		t.Errorf("m/a ΔDistance: got %v, want %v", got, want)
	}
	if got, want := a.LCOM, 2; got != want {
		t.Errorf("m/a ΔLCOM: got %v, want %v", got, want)
	}

	if got, want := len(gd.Added), 0; got != want {
		t.Errorf("len(Added): got %v, want %v", got, want)
	}
	if got, want := len(gd.Removed), 0; got != want {
		t.Errorf("len(Removed): got %v, want %v", got, want)
	}
}

func TestComputeDelta_AddedAndRemovedClassified(t *testing.T) {
	t.Parallel()

	base := completeGraph(t,
		mod("keep", 1, 1, 0.0, 0.0, 0.0, 0),
		mod("zzz", 1, 1, 0.0, 0.0, 0.0, 0),
		mod("gone", 1, 1, 0.0, 0.0, 0.0, 0),
	)
	pr := completeGraph(t,
		mod("keep", 1, 1, 0.0, 0.0, 0.0, 0),
		mod("new", 1, 1, 0.0, 0.0, 0.0, 0),
		mod("aaa", 1, 1, 0.0, 0.0, 0.0, 0),
	)

	gd := ComputeDelta(base, pr)

	if got, want := len(gd.Modules), 1; got != want {
		t.Fatalf("len(Modules): got %v, want %v", got, want)
	}
	if got, want := gd.Modules[0].Path, "keep"; got != want {
		t.Errorf("Modules[0].Path: got %v, want %v", got, want)
	}
	if got, want := gd.Added, []string{"aaa", "new"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Added: got %v, want %v", got, want)
	}
	if got, want := gd.Removed, []string{"gone", "zzz"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Removed: got %v, want %v", got, want)
	}
}

func TestComputeDelta_CycleClassification(t *testing.T) {
	t.Parallel()

	// base cycles use unsorted members to prove normalization; {a,b} is
	// pre-existing (in both), {e,f} is resolved (base-only), {c,d} is new (PR-only).
	base := mustGraph(t, StatusComplete,
		[]ModuleResult{mod("a", 1, 1, 0, 0, 0, 0)},
		[]Cycle{{"b", "a"}, {"e", "f"}}, nil)
	pr := mustGraph(t, StatusComplete,
		[]ModuleResult{mod("a", 1, 1, 0, 0, 0, 0)},
		[]Cycle{{"a", "b"}, {"d", "c"}}, nil)

	gd := ComputeDelta(base, pr)

	if got, want := gd.NewCycles, []Cycle{{"c", "d"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("NewCycles: got %v, want %v", got, want)
	}
	if got, want := gd.ResolvedCycles, []Cycle{{"e", "f"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("ResolvedCycles: got %v, want %v", got, want)
	}
}

func TestComputeDelta_CycleRotationInvariant(t *testing.T) {
	t.Parallel()

	base := mustGraph(t, StatusComplete, nil, []Cycle{{"a", "b", "c"}}, nil)
	pr := mustGraph(t, StatusComplete, nil, []Cycle{{"b", "c", "a"}}, nil)

	gd := ComputeDelta(base, pr)

	if got, want := len(gd.NewCycles), 0; got != want {
		t.Errorf("NewCycles (rotation should not be new): got %v, want %v", gd.NewCycles, want)
	}
	if got, want := len(gd.ResolvedCycles), 0; got != want {
		t.Errorf("ResolvedCycles (rotation should not be resolved): got %v, want %v", gd.ResolvedCycles, want)
	}
}

func TestComputeDelta_CyclesSortedByKey(t *testing.T) {
	t.Parallel()

	base := mustGraph(t, StatusComplete, nil, nil, nil)
	pr := mustGraph(t, StatusComplete, nil,
		[]Cycle{{"x", "y"}, {"m", "n"}, {"a", "b"}}, nil)

	gd := ComputeDelta(base, pr)

	want := []Cycle{{"a", "b"}, {"m", "n"}, {"x", "y"}}
	if got := gd.NewCycles; !reflect.DeepEqual(got, want) {
		t.Errorf("NewCycles sort order: got %v, want %v", got, want)
	}
}

func TestComputeDelta_UnreliableInputs(t *testing.T) {
	t.Parallel()

	oneMod := []ModuleResult{mod("x", 1, 1, 0.0, 0.0, 0.0, 0)}
	complete := mustGraph(t, StatusComplete, oneMod, nil, nil)

	tests := []struct {
		name           string
		base           *ModuleGraph
		pr             *ModuleGraph
		wantUnreliable bool
	}{
		{
			name:           "both_complete",
			base:           complete,
			pr:             complete,
			wantUnreliable: false,
		},
		{
			// Status alone marks this unreliable: no load-error warning is present.
			name:           "base_status_partial",
			base:           mustGraph(t, StatusPartial, oneMod, nil, nil),
			pr:             complete,
			wantUnreliable: true,
		},
		{
			name:           "pr_status_partial",
			base:           complete,
			pr:             mustGraph(t, StatusPartial, oneMod, nil, nil),
			wantUnreliable: true,
		},
		{
			name:           "pr_status_error",
			base:           complete,
			pr:             mustGraph(t, StatusError, oneMod, nil, nil),
			wantUnreliable: true,
		},
		{
			// The load-error warning alone marks this unreliable: Status is complete.
			name: "base_load_error_warning_status_complete",
			base: mustGraph(t, StatusComplete, oneMod, nil,
				[]Warning{{Code: "load-error", Message: "package x: type info unavailable", Module: "x"}}),
			pr:             complete,
			wantUnreliable: true,
		},
		{
			// A non-load-error warning does NOT by itself mark the graph unreliable.
			name: "non_load_error_warning_is_reliable",
			base: mustGraph(t, StatusComplete, oneMod, nil,
				[]Warning{{Code: "dynamic-imports", Message: "cannot resolve dynamic import"}}),
			pr:             complete,
			wantUnreliable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gd := ComputeDelta(tt.base, tt.pr)
			if got, want := gd.Unreliable, tt.wantUnreliable; got != want {
				t.Errorf("Unreliable: got %v, want %v", got, want)
			}
			if tt.wantUnreliable {
				if got, want := gd.Direction, EntropyStable; got != want {
					t.Errorf("Direction when unreliable: got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestComputeDelta_UnreliableSuppressesAddedRemoved(t *testing.T) {
	t.Parallel()

	// base has {shared, gone}; PR has {shared, added}. Were the measurement
	// reliable, "added"/"gone" would populate Added/Removed. A partial PR must
	// suppress both, while still computing the shared module's delta.
	base := completeGraph(t,
		mod("shared", 1, 1, 0.50, 0.0, 0.0, 0),
		mod("gone", 1, 1, 0.0, 0.0, 0.0, 0),
	)
	pr := mustGraph(t, StatusPartial,
		[]ModuleResult{
			mod("shared", 1, 1, 0.70, 0.0, 0.0, 0),
			mod("added", 1, 1, 0.0, 0.0, 0.0, 0),
		}, nil,
		[]Warning{{Code: "load-error", Message: "package added: build failed", Module: "added"}})

	gd := ComputeDelta(base, pr)

	if !gd.Unreliable {
		t.Fatal("Unreliable: got false, want true")
	}
	if got, want := len(gd.Added), 0; got != want {
		t.Errorf("Added suppressed: got %v, want empty", gd.Added)
	}
	if got, want := len(gd.Removed), 0; got != want {
		t.Errorf("Removed suppressed: got %v, want empty", gd.Removed)
	}
	if gd.Added == nil || gd.Removed == nil {
		t.Error("suppressed Added/Removed must be non-nil empty slices")
	}
	// Modules are still computed even when unreliable.
	if got, want := len(gd.Modules), 1; got != want {
		t.Fatalf("len(Modules): got %v, want %v", got, want)
	}
	if got, want := gd.Modules[0].Path, "shared"; got != want {
		t.Errorf("Modules[0].Path: got %v, want %v", got, want)
	}
	if got, want := gd.Direction, EntropyStable; got != want {
		t.Errorf("Direction: got %v, want %v", got, want)
	}
}

func TestComputeDelta_NilInputsAreUnreliable(t *testing.T) {
	t.Parallel()

	pr := completeGraph(t, mod("x", 1, 1, 0.0, 0.0, 0.0, 0))

	tests := []struct {
		name string
		base *ModuleGraph
		pr   *ModuleGraph
	}{
		{name: "nil_base", base: nil, pr: pr},
		{name: "nil_pr", base: pr, pr: nil},
		{name: "both_nil", base: nil, pr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gd := ComputeDelta(tt.base, tt.pr)
			if !gd.Unreliable {
				t.Error("Unreliable: got false, want true")
			}
			if got, want := gd.Direction, EntropyStable; got != want {
				t.Errorf("Direction: got %v, want %v", got, want)
			}
			// All slices must be non-nil empty (safe to range and marshal).
			if gd.Modules == nil || gd.Added == nil || gd.Removed == nil ||
				gd.NewCycles == nil || gd.ResolvedCycles == nil {
				t.Errorf("nil-input result must have non-nil empty slices: %+v", gd)
			}
			if len(gd.Modules)+len(gd.Added)+len(gd.Removed)+len(gd.NewCycles)+len(gd.ResolvedCycles) != 0 {
				t.Errorf("nil-input result must have empty slices: %+v", gd)
			}
		})
	}
}

func TestComputeDelta_EmptyVersusEmpty(t *testing.T) {
	t.Parallel()

	base := completeGraph(t)
	pr := completeGraph(t)

	gd := ComputeDelta(base, pr)

	if gd.Unreliable {
		t.Error("Unreliable: got true, want false")
	}
	if got, want := gd.Direction, EntropyStable; got != want {
		t.Errorf("Direction: got %v, want %v", got, want)
	}
	if len(gd.Modules)+len(gd.Added)+len(gd.Removed)+len(gd.NewCycles)+len(gd.ResolvedCycles) != 0 {
		t.Errorf("empty-vs-empty must produce empty deltas: %+v", gd)
	}
}

func TestComputeDelta_DirectionClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base *ModuleGraph
		pr   *ModuleGraph
		want EntropyDirection
	}{
		{
			name: "degrading_instability_gate",
			base: completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:   completeGraph(t, mod("m", 1, 1, 0.15, 0, 0, 0)),
			want: EntropyDegrading,
		},
		{
			name: "degrading_new_cycle",
			base: mustGraph(t, StatusComplete, []ModuleResult{mod("m", 1, 1, 0, 0, 0, 0)}, nil, nil),
			pr:   mustGraph(t, StatusComplete, []ModuleResult{mod("m", 1, 1, 0, 0, 0, 0)}, []Cycle{{"m", "n"}}, nil),
			want: EntropyDegrading,
		},
		{
			name: "improving_instability",
			base: completeGraph(t, mod("m", 1, 1, 0.20, 0, 0, 0)),
			pr:   completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			want: EntropyImproving,
		},
		{
			name: "improving_resolved_cycle",
			base: mustGraph(t, StatusComplete, []ModuleResult{mod("m", 1, 1, 0, 0, 0, 0)}, []Cycle{{"m", "n"}}, nil),
			pr:   mustGraph(t, StatusComplete, []ModuleResult{mod("m", 1, 1, 0, 0, 0, 0)}, nil, nil),
			want: EntropyImproving,
		},
		{
			name: "stable_small_regression",
			base: completeGraph(t, mod("m", 1, 1, 0.00, 0, 0, 0)),
			pr:   completeGraph(t, mod("m", 1, 1, 0.05, 0, 0, 0)),
			want: EntropyStable,
		},
		{
			name: "stable_no_change",
			base: completeGraph(t, mod("m", 1, 1, 0.30, 0, 0, 0)),
			pr:   completeGraph(t, mod("m", 1, 1, 0.30, 0, 0, 0)),
			want: EntropyStable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gd := ComputeDelta(tt.base, tt.pr)
			if got := gd.Direction; got != tt.want {
				t.Errorf("Direction: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeDelta_Deterministic(t *testing.T) {
	t.Parallel()

	base := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 1, 2, 0.40, 0.10, 0.30, 2),
			mod("m/gone", 0, 0, 0.0, 0.0, 0.0, 0),
		},
		[]Cycle{{"c", "b", "a"}}, nil)
	pr := mustGraph(t, StatusComplete,
		[]ModuleResult{
			mod("m/a", 2, 3, 0.55, 0.20, 0.45, 4),
			mod("m/new", 0, 0, 0.0, 0.0, 0.0, 0),
		},
		[]Cycle{{"d", "e"}}, nil)

	first := ComputeDelta(base, pr)
	second := ComputeDelta(base, pr)

	if !reflect.DeepEqual(first, second) {
		t.Errorf("ComputeDelta not deterministic:\n got %+v\nwant %+v", second, first)
	}
}
