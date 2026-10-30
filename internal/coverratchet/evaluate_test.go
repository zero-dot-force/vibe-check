package coverratchet

import (
	"slices"
	"testing"
)

func TestEvaluate_Table(t *testing.T) {
	tests := []struct {
		name       string
		measured   map[string]Coverage
		baseline   Baseline
		wantStatus map[string]Status
		wantOK     bool
	}{
		{
			name:     "package below its floor fails",
			measured: map[string]Coverage{"metrics": {Covered: 400, Total: 500}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusBelowFloor,
			},
			wantOK: false,
		},
		{
			name:     "measurement exactly at the floor passes",
			measured: map[string]Coverage{"metrics": {Covered: 829, Total: 1000}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusPass,
			},
			wantOK: true,
		},
		{
			name:     "package above its floor passes",
			measured: map[string]Coverage{"metrics": {Covered: 9, Total: 10}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusPass,
			},
			wantOK: true,
		},
		{
			name: "unregistered package fails",
			measured: map[string]Coverage{
				"internal/coverratchet": {Covered: 9, Total: 10},
				"metrics":               {Covered: 829, Total: 1000},
			},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"internal/coverratchet": StatusUnregistered,
				"metrics":               StatusPass,
			},
			wantOK: false,
		},
		{
			name:     "stale floor fails instead of passing vacuously",
			measured: map[string]Coverage{},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusStaleFloor,
			},
			wantOK: false,
		},
		{
			name: "all registered packages pass completeness",
			measured: map[string]Coverage{
				"cmd/vibe-check":     {Covered: 927, Total: 1000},
				"internal/goadapter": {Covered: 9, Total: 10},
			},
			baseline: Baseline{Version: 1, Floors: map[string]float64{
				"cmd/vibe-check":     92.7,
				"internal/goadapter": 89.3,
			}},
			wantStatus: map[string]Status{
				"cmd/vibe-check":     StatusPass,
				"internal/goadapter": StatusPass,
			},
			wantOK: true,
		},
		{
			name:     "empty profile and empty baseline pass",
			measured: map[string]Coverage{},
			baseline: Baseline{Version: 1, Floors: map[string]float64{}},
			wantOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.measured, tt.baseline)
			if got.OK() != tt.wantOK {
				t.Errorf("OK() = %v, want %v", got.OK(), tt.wantOK)
			}
			if len(got.Packages) != len(tt.wantStatus) {
				t.Fatalf("len(Packages) = %d, want %d; got %+v", len(got.Packages), len(tt.wantStatus), got.Packages)
			}
			for _, p := range got.Packages {
				want, ok := tt.wantStatus[p.Package]
				if !ok {
					t.Fatalf("unexpected package %q in result", p.Package)
				}
				if p.Status != want {
					t.Errorf("package %q status = %q, want %q", p.Package, p.Status, want)
				}
			}
		})
	}
}

func TestEvaluate_ReportsMeasuredAndFloor(t *testing.T) {
	measured := map[string]Coverage{
		"metrics":               {Covered: 829, Total: 1000},
		"internal/coverratchet": {Covered: 9, Total: 10},
	}
	baseline := Baseline{Version: 1, Floors: map[string]float64{
		"metrics": 82.9,
		"gone":    50.0,
	}}
	got := Evaluate(measured, baseline)
	byPkg := make(map[string]PackageResult)
	for _, p := range got.Packages {
		byPkg[p.Package] = p
	}
	if p := byPkg["metrics"]; p.Measured != 82.9 || p.Floor != 82.9 {
		t.Errorf("metrics result = {Measured %v, Floor %v}, want {82.9, 82.9}", p.Measured, p.Floor)
	}
	if p := byPkg["internal/coverratchet"]; p.Measured != 90.0 || p.Floor != 0 {
		t.Errorf("internal/coverratchet result = {Measured %v, Floor %v}, want {90, 0}", p.Measured, p.Floor)
	}
	if p := byPkg["gone"]; p.Measured != 0 || p.Floor != 50.0 {
		t.Errorf("gone result = {Measured %v, Floor %v}, want {0, 50}", p.Measured, p.Floor)
	}
}

func TestEvaluate_SortedByPackagePath(t *testing.T) {
	measured := map[string]Coverage{
		"metrics":        {Covered: 9, Total: 10},
		"cmd/vibe-check": {Covered: 9, Total: 10},
	}
	baseline := Baseline{Version: 1, Floors: map[string]float64{
		"metrics":        80.0,
		"aardvark":       80.0,
		"cmd/vibe-check": 80.0,
	}}
	got := Evaluate(measured, baseline)
	var paths []string
	for _, p := range got.Packages {
		paths = append(paths, p.Package)
	}
	if !slices.IsSorted(paths) {
		t.Errorf("package paths not sorted: %v", paths)
	}
}
