package coverratchet

import (
	"maps"
	"reflect"
	"testing"
)

func TestUpdate_Table(t *testing.T) {
	tests := []struct {
		name       string
		measured   map[string]Coverage
		baseline   Baseline
		wantStatus map[string]Status
		wantOK     bool
		wantFloors map[string]float64
	}{
		{
			name:     "raise to measured",
			measured: map[string]Coverage{"metrics": {Covered: 9, Total: 10}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusRaised,
			},
			wantOK:     true,
			wantFloors: map[string]float64{"metrics": 90.0},
		},
		{
			name:     "measurement at floor leaves entry unchanged",
			measured: map[string]Coverage{"metrics": {Covered: 829, Total: 1000}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusUnchanged,
			},
			wantOK:     true,
			wantFloors: map[string]float64{"metrics": 82.9},
		},
		{
			name:     "never lower below floor",
			measured: map[string]Coverage{"metrics": {Covered: 4, Total: 5}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"metrics": StatusBelowFloor,
			},
			wantOK: false,
		},
		{
			name: "refuse new package below entry bar",
			measured: map[string]Coverage{
				"internal/coverratchet": {Covered: 3, Total: 4},
				"metrics":               {Covered: 829, Total: 1000},
			},
			baseline: Baseline{Version: 1, Floors: map[string]float64{"metrics": 82.9}},
			wantStatus: map[string]Status{
				"internal/coverratchet": StatusRefused,
				"metrics":               StatusUnchanged,
			},
			wantOK: false,
		},
		{
			name:     "add new package at the entry bar",
			measured: map[string]Coverage{"internal/coverratchet": {Covered: 4, Total: 5}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{}},
			wantStatus: map[string]Status{
				"internal/coverratchet": StatusAdded,
			},
			wantOK:     true,
			wantFloors: map[string]float64{"internal/coverratchet": 80.0},
		},
		{
			name:     "add new package above the entry bar",
			measured: map[string]Coverage{"internal/coverratchet": {Covered: 9, Total: 10}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{}},
			wantStatus: map[string]Status{
				"internal/coverratchet": StatusAdded,
			},
			wantOK:     true,
			wantFloors: map[string]float64{"internal/coverratchet": 90.0},
		},
		{
			name: "update leaves other floors untouched",
			measured: map[string]Coverage{
				"metrics":            {Covered: 9, Total: 10},
				"internal/goadapter": {Covered: 4, Total: 5},
				"internal/scaffold":  {Covered: 879, Total: 1000},
			},
			baseline: Baseline{Version: 1, Floors: map[string]float64{
				"metrics":            82.9,
				"internal/goadapter": 80.0,
				"internal/scaffold":  87.9,
			}},
			wantStatus: map[string]Status{
				"metrics":            StatusRaised,
				"internal/goadapter": StatusUnchanged,
				"internal/scaffold":  StatusUnchanged,
			},
			wantOK: true,
			wantFloors: map[string]float64{
				"metrics":            90.0,
				"internal/goadapter": 80.0,
				"internal/scaffold":  87.9,
			},
		},
		{
			name: "mixed run is all-or-nothing",
			measured: map[string]Coverage{
				"metrics":            {Covered: 9, Total: 10},
				"internal/goadapter": {Covered: 7, Total: 10},
			},
			baseline: Baseline{Version: 1, Floors: map[string]float64{
				"metrics":            82.9,
				"internal/goadapter": 80.0,
			}},
			wantStatus: map[string]Status{
				"metrics":            StatusRaised,
				"internal/goadapter": StatusBelowFloor,
			},
			wantOK: false,
		},
		{
			name:     "stale floor blocks update",
			measured: map[string]Coverage{"metrics": {Covered: 9, Total: 10}},
			baseline: Baseline{Version: 1, Floors: map[string]float64{
				"metrics": 82.9,
				"gone":    50.0,
			}},
			wantStatus: map[string]Status{
				"metrics": StatusRaised,
				"gone":    StatusStaleFloor,
			},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := maps.Clone(tt.baseline.Floors)
			got, result := Update(tt.measured, tt.baseline)
			if result.OK() != tt.wantOK {
				t.Errorf("OK() = %v, want %v", result.OK(), tt.wantOK)
			}
			if len(result.Packages) != len(tt.wantStatus) {
				t.Fatalf("len(Packages) = %d, want %d; got %+v", len(result.Packages), len(tt.wantStatus), result.Packages)
			}
			for _, p := range result.Packages {
				want, ok := tt.wantStatus[p.Package]
				if !ok {
					t.Fatalf("unexpected package %q in result", p.Package)
				}
				if p.Status != want {
					t.Errorf("package %q status = %q, want %q", p.Package, p.Status, want)
				}
				if p.Status == StatusAdded || p.Status == StatusRaised {
					if p.Floor != p.Measured {
						t.Errorf("package %q row floor = %v, want the post-update floor %v", p.Package, p.Floor, p.Measured)
					}
				}
			}
			if !reflect.DeepEqual(tt.baseline.Floors, original) {
				t.Errorf("Update mutated the input baseline floors: got %v, want %v", tt.baseline.Floors, original)
			}
			if tt.wantOK {
				if got.Version != BaselineVersion {
					t.Errorf("updated baseline version = %d, want %d", got.Version, BaselineVersion)
				}
				if !reflect.DeepEqual(got.Floors, tt.wantFloors) {
					t.Errorf("updated floors = %v, want %v", got.Floors, tt.wantFloors)
				}
			} else if !reflect.DeepEqual(got.Floors, original) {
				t.Errorf("violating update returned changed floors: got %v, want the input %v", got.Floors, original)
			}
		})
	}
}
