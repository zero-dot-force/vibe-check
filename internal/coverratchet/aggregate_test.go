package coverratchet

import (
	"math"
	"path/filepath"
	"testing"

	"golang.org/x/tools/cover"
)

func TestCoverage_Percent_Truncation(t *testing.T) {
	tests := []struct {
		name    string
		covered int
		total   int
		want    float64
	}{
		{name: "89.99 truncates to 89.9", covered: 8999, total: 10000, want: 89.9},
		{name: "exact one-decimal value stays put", covered: 899, total: 1000, want: 89.9},
		{name: "33.33 truncates to 33.3", covered: 1, total: 3, want: 33.3},
		{name: "66.66 truncates to 66.6", covered: 2, total: 3, want: 66.6},
		{name: "sub-decimal coverage truncates to 0.0", covered: 1, total: 10000, want: 0.0},
		{name: "full coverage", covered: 5, total: 5, want: 100.0},
		{name: "no coverage", covered: 0, total: 5, want: 0.0},
		{name: "zero total reports 0 rather than panicking", covered: 0, total: 0, want: 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Coverage{Covered: tt.covered, Total: tt.total}.Percent()
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAggregate_AtomicNonZeroCounts(t *testing.T) {
	profiles, err := LoadProfiles(filepath.Join("testdata", "atomic.out"))
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}
	agg, err := Aggregate(profiles)
	if err != nil {
		t.Fatalf("Aggregate error: %v", err)
	}
	c, ok := agg["metrics"]
	if !ok {
		t.Fatalf("Aggregate missing package %q; got %v", "metrics", agg)
	}
	// Blocks: NumStmt 2 with Count 7 (covered), NumStmt 3 with Count 0
	// (uncovered), NumStmt 2 with Count 4 (covered). In mode: atomic any
	// non-zero count means executed, so covered = 2+2 = 4, total = 7.
	if c.Covered != 4 {
		t.Errorf("Covered = %d, want 4 (non-zero counts count as executed)", c.Covered)
	}
	if c.Total != 7 {
		t.Errorf("Total = %d, want 7", c.Total)
	}
	if want := 57.1; math.Abs(c.Percent()-want) > 1e-9 {
		t.Errorf("Percent() = %v, want %v", c.Percent(), want)
	}
}

func TestAggregate_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		profiles []*cover.Profile
		want     map[string]Coverage
	}{
		{
			name: "package whose blocks are all uncovered",
			profiles: []*cover.Profile{{
				FileName: "github.com/zero-dot-force/vibe-check/metrics/zone.go",
				Mode:     "set",
				Blocks: []cover.ProfileBlock{
					{NumStmt: 3, Count: 0},
					{NumStmt: 4, Count: 0},
				},
			}},
			want: map[string]Coverage{"metrics": {Covered: 0, Total: 7}},
		},
		{
			name: "package with only zero-statement blocks is excluded",
			profiles: []*cover.Profile{{
				FileName: "github.com/zero-dot-force/vibe-check/internal/empty/empty.go",
				Mode:     "set",
				Blocks: []cover.ProfileBlock{
					{NumStmt: 0, Count: 0},
					{NumStmt: 0, Count: 5},
				},
			}},
			want: map[string]Coverage{},
		},
		{
			name: "two files in the same package aggregate together",
			profiles: []*cover.Profile{
				{
					FileName: "github.com/zero-dot-force/vibe-check/metrics/one.go",
					Mode:     "set",
					Blocks:   []cover.ProfileBlock{{NumStmt: 2, Count: 1}},
				},
				{
					FileName: "github.com/zero-dot-force/vibe-check/metrics/two.go",
					Mode:     "set",
					Blocks:   []cover.ProfileBlock{{NumStmt: 3, Count: 0}, {NumStmt: 5, Count: 1}},
				},
			},
			want: map[string]Coverage{"metrics": {Covered: 7, Total: 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Aggregate(tt.profiles)
			if err != nil {
				t.Fatalf("Aggregate error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len(Aggregate) = %d, want %d; got %v", len(got), len(tt.want), got)
			}
			for key, wantC := range tt.want {
				gotC, ok := got[key]
				if !ok {
					t.Fatalf("Aggregate missing package %q; got %v", key, got)
				}
				if gotC != wantC {
					t.Errorf("Aggregate[%q] = %+v, want %+v", key, gotC, wantC)
				}
			}
		})
	}
}

func TestAggregate_UnprefixablePath(t *testing.T) {
	profiles := []*cover.Profile{
		{
			FileName: "example.com/other/pkg/tool.go",
			Mode:     "set",
			Blocks:   []cover.ProfileBlock{{NumStmt: 5, Count: 1}},
		},
	}
	if _, err := Aggregate(profiles); err == nil {
		t.Errorf("Aggregate on profile outside the module = nil error, want error")
	}
}
