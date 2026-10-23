package report

import (
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// TestGradeEmoji_Mapping pins the report spec's grade-status-emoji
// requirement: A and B+ render green, B and C+ render yellow, C and F
// render red. Grades off the scale have no emoji.
func TestGradeEmoji_Mapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		grade Grade
		want  string
	}{
		{name: "A is green", grade: GradeA, want: "🟢"},
		{name: "B+ is green", grade: GradeBPlus, want: "🟢"},
		{name: "B is yellow", grade: GradeB, want: "🟡"},
		{name: "C+ is yellow", grade: GradeCPlus, want: "🟡"},
		{name: "C is red", grade: GradeC, want: "🔴"},
		{name: "F is red", grade: GradeF, want: "🔴"},
		{name: "D is absent from the scale and has no emoji", grade: Grade("D"), want: ""},
		{name: "zero value has no emoji", grade: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.grade.Emoji(); got != tt.want {
				t.Errorf("Grade(%q).Emoji() = %q, want %q", tt.grade, got, tt.want)
			}
		})
	}
}

// TestClassifyDistance_Boundaries pins the report spec's health-scorecard
// grading requirement for average distance: every band boundary exactly,
// values just below and above each boundary, and the 0.0/1.0 extremes.
// The spec's distance-boundary scenario — exactly 0.1 grades B+, because
// A requires strictly less than 0.1 — is the "0.1 boundary grades B+"
// case.
func TestClassifyDistance_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		avg  metrics.Distance
		want Grade
	}{
		{name: "0.0 extreme grades A", avg: 0, want: GradeA},
		{name: "0.05 mid-band grades A", avg: 0.05, want: GradeA},
		{name: "0.0999 just below B+ boundary grades A", avg: 0.0999, want: GradeA},
		{name: "0.1 boundary grades B+", avg: 0.1, want: GradeBPlus},
		{name: "0.1001 just above B+ boundary grades B+", avg: 0.1001, want: GradeBPlus},
		{name: "0.1999 just below B boundary grades B+", avg: 0.1999, want: GradeBPlus},
		{name: "0.2 boundary grades B", avg: 0.2, want: GradeB},
		{name: "0.2001 just above B boundary grades B", avg: 0.2001, want: GradeB},
		{name: "0.2999 just below C+ boundary grades B", avg: 0.2999, want: GradeB},
		{name: "0.3 boundary grades C+", avg: 0.3, want: GradeCPlus},
		{name: "0.3001 just above C+ boundary grades C+", avg: 0.3001, want: GradeCPlus},
		{name: "0.4999 just below C boundary grades C+", avg: 0.4999, want: GradeCPlus},
		{name: "0.5 boundary grades C", avg: 0.5, want: GradeC},
		{name: "0.5001 just above C boundary grades C", avg: 0.5001, want: GradeC},
		{name: "0.6999 just below F boundary grades C", avg: 0.6999, want: GradeC},
		{name: "0.7 boundary grades F", avg: 0.7, want: GradeF},
		{name: "0.7001 just above F boundary grades F", avg: 0.7001, want: GradeF},
		{name: "1.0 extreme grades F", avg: 1.0, want: GradeF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyDistance(tt.avg); got != tt.want {
				t.Errorf("ClassifyDistance(%v): got %v, want %v", tt.avg, got, tt.want)
			}
		})
	}
}

// TestClassifyLCOM4_Boundaries pins the report spec's health-scorecard
// grading requirement for max LCOM4. LCOM4 is an integer, so each band
// below F contains exactly one value and the boundary table is
// exhaustive. The spec's LCOM-boundary scenario — max LCOM4 of exactly 4
// grades C+ — is the "4 grades C+" case; values 0 and 1 both sit in the
// A band below the B+ minimum of 2.
func TestClassifyLCOM4_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		maxLCOM metrics.LCOM
		want    Grade
	}{
		{name: "0 extreme grades A", maxLCOM: 0, want: GradeA},
		{name: "1 fully cohesive grades A", maxLCOM: 1, want: GradeA},
		{name: "2 boundary grades B+", maxLCOM: 2, want: GradeBPlus},
		{name: "3 boundary grades B", maxLCOM: 3, want: GradeB},
		{name: "4 boundary grades C+", maxLCOM: 4, want: GradeCPlus},
		{name: "5 boundary grades C", maxLCOM: 5, want: GradeC},
		{name: "6 boundary grades F", maxLCOM: 6, want: GradeF},
		{name: "7 above F boundary grades F", maxLCOM: 7, want: GradeF},
		{name: "12 extreme grades F", maxLCOM: 12, want: GradeF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyLCOM4(tt.maxLCOM); got != tt.want {
				t.Errorf("ClassifyLCOM4(%v): got %v, want %v", tt.maxLCOM, got, tt.want)
			}
		})
	}
}

// TestClassifyCycles_Boundaries pins the report spec's any-cycle-fails
// scenario: zero cycles grades A, one cycle grades F, and additional
// cycles stay F.
func TestClassifyCycles_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  Grade
	}{
		{name: "0 cycles grades A", count: 0, want: GradeA},
		{name: "1 cycle grades F", count: 1, want: GradeF},
		{name: "2 cycles grade F", count: 2, want: GradeF},
		{name: "5 cycles grade F", count: 5, want: GradeF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyCycles(tt.count); got != tt.want {
				t.Errorf("ClassifyCycles(%v): got %v, want %v", tt.count, got, tt.want)
			}
		})
	}
}

// TestClassifyDuplication_Boundaries pins the report spec's
// health-scorecard grading requirement for duplication percentage: the
// A-band point case at exactly 0%, every band boundary exactly, values
// just below and above each boundary, and the 100% extreme. The spec's
// duplication-boundary scenario — exactly 5% grades C+, because B
// requires strictly less than 5% — is the "5% boundary grades C+" case.
func TestClassifyDuplication_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pct  float64
		want Grade
	}{
		{name: "0% point band grades A", pct: 0, want: GradeA},
		{name: "0.0001% just above zero grades B+", pct: 0.0001, want: GradeBPlus},
		{name: "1.5% mid-band grades B+", pct: 1.5, want: GradeBPlus},
		{name: "2.9999% just below B boundary grades B+", pct: 2.9999, want: GradeBPlus},
		{name: "3% boundary grades B", pct: 3, want: GradeB},
		{name: "3.0001% just above B boundary grades B", pct: 3.0001, want: GradeB},
		{name: "4.9999% just below C+ boundary grades B", pct: 4.9999, want: GradeB},
		{name: "5% boundary grades C+", pct: 5, want: GradeCPlus},
		{name: "5.0001% just above C+ boundary grades C+", pct: 5.0001, want: GradeCPlus},
		{name: "9.9999% just below C boundary grades C+", pct: 9.9999, want: GradeCPlus},
		{name: "10% boundary grades C", pct: 10, want: GradeC},
		{name: "10.0001% just above C boundary grades C", pct: 10.0001, want: GradeC},
		{name: "14.9999% just below F boundary grades C", pct: 14.9999, want: GradeC},
		{name: "15% boundary grades F", pct: 15, want: GradeF},
		{name: "15.0001% just above F boundary grades F", pct: 15.0001, want: GradeF},
		{name: "100% extreme grades F", pct: 100, want: GradeF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyDuplication(tt.pct); got != tt.want {
				t.Errorf("ClassifyDuplication(%v): got %v, want %v", tt.pct, got, tt.want)
			}
		})
	}
}

// TestClassifySpread_Boundaries pins the report spec's health-scorecard
// grading requirement for instability spread. Spread is an integer count,
// so each band contains exactly one value: 0 grades A (the point band),
// exactly 1 grades B, exactly 2 grades C+, and 3 or more grade F.
func TestClassifySpread_Boundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  Grade
	}{
		{name: "0 spread grades A", count: 0, want: GradeA},
		{name: "1 spread grades B", count: 1, want: GradeB},
		{name: "2 spread grades C+", count: 2, want: GradeCPlus},
		{name: "3 spread grades F", count: 3, want: GradeF},
		{name: "4 spread grades F", count: 4, want: GradeF},
		{name: "10 spread extreme grades F", count: 10, want: GradeF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifySpread(tt.count); got != tt.want {
				t.Errorf("ClassifySpread(%v): got %v, want %v", tt.count, got, tt.want)
			}
		})
	}
}

// dupMod builds a ModuleResult carrying duplication data and a total line
// count, reusing mod for the metric fields (instability 0.5, distance
// 0.25, LCOM4 1 — exact binary values so multi-module means stay exact).
// Every Duplication fixture it carries must be schema-shaped: modulePath,
// blocks, and similarity, with file/startLine/endLine/lineCount per block.
func dupMod(path string, totalLines int, dups ...metrics.Duplication) metrics.ModuleResult {
	m := mod(path, 1, 1, 0.5, 0, 0.25, 1)
	m.TotalLines = totalLines
	m.Duplications = dups
	return m
}

// dupBlock builds a schema-shaped DuplicateBlock fixture: a file path
// relative to the project root plus 1-indexed inclusive line bounds and a
// significant-line count.
func dupBlock(file string, startLine, endLine, lineCount int) metrics.DuplicateBlock {
	return metrics.DuplicateBlock{
		File:      file,
		StartLine: startLine,
		EndLine:   endLine,
		LineCount: lineCount,
	}
}

// TestComputeAggregates_EmptyModuleSet pins the report spec's
// empty-module-set scenario: a graph with zero modules yields every
// aggregate at its zero value — the averages do not divide by zero — and a
// nil graph is handled the same way, mirroring ComputeDelta's nil
// contract.
func TestComputeAggregates_EmptyModuleSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
	}{
		{name: "nil graph", g: nil},
		{name: "zero modules", g: mustGraph(t, metrics.StatusComplete, []metrics.ModuleResult{}, nil, nil)},
	}

	want := Aggregates{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ComputeAggregates(tt.g); got != want {
				t.Errorf("ComputeAggregates(%s): got %+v, want %+v", tt.name, got, want)
			}
		})
	}
}

// TestComputeAggregates_ModuleMeans pins the mean computations for average
// instability and average distance — and, because every case compares the
// full Aggregates struct, the max LCOM4 alongside them. Fixture values are
// exact in binary floating point so the expected means are exact.
func TestComputeAggregates_ModuleMeans(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mods []metrics.ModuleResult
		want Aggregates
	}{
		{
			name: "single module means equal its values",
			mods: []metrics.ModuleResult{mod("pkg/a", 3, 1, 0.25, 0.5, 0.125, 2)},
			want: Aggregates{AverageInstability: 0.25, AverageDistance: 0.125, MaxLCOM4: 2},
		},
		{
			name: "three modules average their values",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 1, 1, 0.25, 0, 0.25, 1),
				mod("pkg/b", 1, 1, 0.5, 0, 0.5, 3),
				mod("pkg/c", 1, 1, 0.75, 0, 0.25, 2),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 1.0 / 3.0, MaxLCOM4: 3},
		},
		{
			name: "0.0 and 1.0 instability extremes average to one half",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 5, 0, 0.0, 0, 0.0, 1),
				mod("pkg/b", 0, 5, 1.0, 0, 0.5, 1),
			},
			// Both modules sit at instability extremes, so they also
			// count toward the spread.
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1, InstabilitySpread: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := mustGraph(t, metrics.StatusComplete, tt.mods, nil, nil)
			if got := ComputeAggregates(g); got != tt.want {
				t.Errorf("ComputeAggregates(%s): got %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

// TestComputeAggregates_MaxLCOM4AndCycleCount pins the max-LCOM4 and
// cycle-count aggregates: the maximum is taken across all modules (0 when
// every module is trivially cohesive), and the cycle count mirrors the
// graph's Cycles field length.
func TestComputeAggregates_MaxLCOM4AndCycleCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mods   []metrics.ModuleResult
		cycles []metrics.Cycle
		want   Aggregates
	}{
		{
			name: "max LCOM4 is the largest module value",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1),
				mod("pkg/b", 1, 1, 0.5, 0, 0.25, 4),
				mod("pkg/c", 1, 1, 0.5, 0, 0.25, 2),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 4},
		},
		{
			name: "trivially cohesive modules leave max LCOM4 at zero",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 1, 1, 0.5, 0, 0.25, 0),
				mod("pkg/b", 1, 1, 0.5, 0, 0.25, 0),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25},
		},
		{
			name:   "cycle count mirrors the Cycles field length",
			mods:   []metrics.ModuleResult{mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1)},
			cycles: []metrics.Cycle{{"pkg/a", "pkg/b"}, {"pkg/b", "pkg/c", "pkg/d"}, {"pkg/e", "pkg/a"}},
			want:   Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1, CycleCount: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := mustGraph(t, metrics.StatusComplete, tt.mods, tt.cycles, nil)
			if got := ComputeAggregates(g); got != tt.want {
				t.Errorf("ComputeAggregates(%s): got %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

// TestComputeAggregates_SpreadCounting pins the report spec's
// spread-counting scenario — modules at instability 0.0, 1.0, and 0.5 give
// a spread of 2 extremes — plus the inclusive bounds (0.001 and 0.999
// count) and values just outside them (0.002 and 0.998 do not).
func TestComputeAggregates_SpreadCounting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mods []metrics.ModuleResult
		want Aggregates
	}{
		{
			name: "spec scenario: 0.0, 1.0, and 0.5 give two extremes",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 5, 0, 0.0, 0, 0.25, 1),
				mod("pkg/b", 0, 5, 1.0, 0, 0.25, 1),
				mod("pkg/c", 1, 1, 0.5, 0, 0.25, 1),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1, InstabilitySpread: 2},
		},
		{
			name: "inclusive bounds: 0.001 and 0.999 count as extremes",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 999, 1, 0.001, 0, 0.25, 1),
				mod("pkg/b", 1, 999, 0.999, 0, 0.25, 1),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1, InstabilitySpread: 2},
		},
		{
			name: "just outside the bounds: 0.002 and 0.998 are not extremes",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 499, 1, 0.002, 0, 0.25, 1),
				mod("pkg/b", 1, 499, 0.998, 0, 0.25, 1),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1},
		},
		{
			name: "every module at an extreme counts",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 5, 0, 0.0, 0, 0.25, 1),
				mod("pkg/b", 0, 5, 1.0, 0, 0.25, 1),
				mod("pkg/c", 9, 0, 0.0, 0, 0.25, 1),
			},
			want: Aggregates{AverageInstability: 1.0 / 3.0, AverageDistance: 0.25, MaxLCOM4: 1, InstabilitySpread: 3},
		},
		{
			name: "mid-band instability values are not extremes",
			mods: []metrics.ModuleResult{
				mod("pkg/a", 1, 1, 0.25, 0, 0.25, 1),
				mod("pkg/b", 1, 1, 0.5, 0, 0.25, 1),
				mod("pkg/c", 1, 1, 0.75, 0, 0.25, 1),
			},
			want: Aggregates{AverageInstability: 0.5, AverageDistance: 0.25, MaxLCOM4: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := mustGraph(t, metrics.StatusComplete, tt.mods, nil, nil)
			if got := ComputeAggregates(g); got != tt.want {
				t.Errorf("ComputeAggregates(%s): got %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

// TestComputeAggregates_DuplicationPercent pins the report spec's
// single-duplicated-block scenario — one 16-line block over 3797 total
// lines computes the percentage that renders as 0.4% — and the
// deduplication contract: a block counts once no matter how many
// duplication groups or modules report it, distinct blocks sum, and zero
// total lines yield 0.0% instead of dividing by zero.
func TestComputeAggregates_DuplicationPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mods []metrics.ModuleResult
		want Aggregates
	}{
		{
			name: "spec scenario: one 16-line block over 3797 total lines renders as 0.4%",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 3797, metrics.Duplication{
					ModulePath: "pkg/a",
					Blocks:     []metrics.DuplicateBlock{dupBlock("a.go", 10, 25, 16)},
					Similarity: 1.0,
				}),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
				DuplicationPercent: 16.0 / 3797.0 * 100,
			},
		},
		{
			name: "the same block reported by two duplication groups counts once",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 1000,
					metrics.Duplication{
						ModulePath: "pkg/a",
						Blocks: []metrics.DuplicateBlock{
							dupBlock("a.go", 10, 25, 16),
							dupBlock("b.go", 40, 55, 16),
						},
						Similarity: 1.0,
					},
					metrics.Duplication{
						ModulePath: "pkg/a",
						Blocks: []metrics.DuplicateBlock{
							// Same file and range as the first
							// group's a.go block: one identity.
							dupBlock("a.go", 10, 25, 16),
							dupBlock("c.go", 5, 20, 16),
						},
						Similarity: 0.95,
					},
				),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
				// a.go once, b.go, c.go: 16 + 16 + 16.
				DuplicationPercent: 48.0 / 1000.0 * 100,
			},
		},
		{
			name: "the same block reported by two modules counts once",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 600, metrics.Duplication{
					ModulePath: "pkg/a",
					Blocks:     []metrics.DuplicateBlock{dupBlock("shared.go", 10, 25, 16)},
					Similarity: 1.0,
				}),
				dupMod("pkg/b", 400, metrics.Duplication{
					ModulePath: "pkg/b",
					Blocks:     []metrics.DuplicateBlock{dupBlock("shared.go", 10, 25, 16)},
					Similarity: 1.0,
				}),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
				// 16 deduplicated lines over 600 + 400 total.
				DuplicationPercent: 16.0 / 1000.0 * 100,
			},
		},
		{
			name: "distinct ranges in the same file are distinct blocks",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 800, metrics.Duplication{
					ModulePath: "pkg/a",
					Blocks: []metrics.DuplicateBlock{
						dupBlock("a.go", 10, 25, 16),
						dupBlock("a.go", 30, 45, 16),
					},
					Similarity: 1.0,
				}),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
				DuplicationPercent: 32.0 / 800.0 * 100,
			},
		},
		{
			name: "conflicting line counts for one identity use the largest",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 1000,
					metrics.Duplication{
						ModulePath: "pkg/a",
						Blocks:     []metrics.DuplicateBlock{dupBlock("a.go", 10, 25, 12)},
						Similarity: 1.0,
					},
					metrics.Duplication{
						ModulePath: "pkg/a",
						Blocks:     []metrics.DuplicateBlock{dupBlock("a.go", 10, 25, 15)},
						Similarity: 1.0,
					},
				),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
				DuplicationPercent: 15.0 / 1000.0 * 100,
			},
		},
		{
			name: "zero total lines yield 0.0 percent without dividing by zero",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 0, metrics.Duplication{
					ModulePath: "pkg/a",
					Blocks:     []metrics.DuplicateBlock{dupBlock("a.go", 10, 25, 16)},
					Similarity: 1.0,
				}),
			},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
			},
		},
		{
			name: "no duplications yield 0.0 percent",
			mods: []metrics.ModuleResult{dupMod("pkg/a", 1000)},
			want: Aggregates{
				AverageInstability: 0.5,
				AverageDistance:    0.25,
				MaxLCOM4:           1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := mustGraph(t, metrics.StatusComplete, tt.mods, nil, nil)
			if got := ComputeAggregates(g); got != tt.want {
				t.Errorf("ComputeAggregates(%s): got %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

// TestComputeAggregates_ModuleOrderIndependence pins the report's
// byte-stability requirement at the engine level: reordering the modules
// array must not change the aggregates. Floating-point addition is not
// associative, so the means are summed in a canonical (ascending) order to
// guarantee this; the fixture deliberately uses non-dyadic values where
// naive slice-order summation could perturb the result.
func TestComputeAggregates_ModuleOrderIndependence(t *testing.T) {
	t.Parallel()

	mods := []metrics.ModuleResult{
		mod("pkg/a", 1, 3, 0.7, 0, 0.1, 2),
		mod("pkg/b", 1, 1, 0.3, 0, 0.3, 5),
		mod("pkg/c", 5, 0, 0.0, 0, 0.7, 1),
		mod("pkg/d", 0, 5, 1.0, 0, 0.5, 3),
		mod("pkg/e", 2, 2, 0.5, 0, 0.9, 4),
	}
	lineTotals := []int{500, 300, 200, 400, 100}
	for i := range mods {
		mods[i].TotalLines = lineTotals[i]
	}
	mods[2].Duplications = []metrics.Duplication{
		{
			ModulePath: "pkg/c",
			Blocks: []metrics.DuplicateBlock{
				dupBlock("c.go", 10, 25, 16),
				dupBlock("c.go", 30, 45, 12),
			},
			Similarity: 1.0,
		},
	}

	reversed := make([]metrics.ModuleResult, len(mods))
	for i, m := range mods {
		reversed[len(mods)-1-i] = m
	}

	forward := mustGraph(t, metrics.StatusComplete, mods, nil, nil)
	backward := mustGraph(t, metrics.StatusComplete, reversed, nil, nil)

	want := ComputeAggregates(forward)
	if got := ComputeAggregates(backward); got != want {
		t.Errorf("ComputeAggregates on reversed module order: got %+v, want %+v", got, want)
	}
}

// wantAllClear is the expected single all-clear recommendation for any
// graph where no rule fires.
var wantAllClear = Recommendation{
	Severity:   SeverityGreen,
	Rule:       RuleAllClear,
	ModulePath: "",
	Message:    "Architecture is healthy — no significant issues detected; keep monitoring",
}

// assertRecommendations fails the test unless got equals want, element
// by element, with per-element diagnostics.
func assertRecommendations(t *testing.T, got, want []Recommendation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d recommendations, want %d: got %+v, want %+v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("recommendation %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

// assertRuleOrder fails the test unless each element of got matches the
// corresponding want's Rule and ModulePath, in order. Ordering tests use
// it to pin sequence while leaving message text to the per-rule
// boundary tests.
func assertRuleOrder(t *testing.T, got, want []Recommendation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d recommendations, want %d: got %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Rule != want[i].Rule || got[i].ModulePath != want[i].ModulePath {
			t.Errorf("recommendation %d: got (%s, %s), want (%s, %s)",
				i, got[i].Rule, got[i].ModulePath, want[i].Rule, want[i].ModulePath)
		}
	}
}

// TestComputeRecommendations_ModuleRuleBoundaries pins the six
// module-scoped rules of the report spec's recommendation table: each
// fires exactly at its threshold — the spec's LCOM-split scenario (a
// module with LCOM4 4 fires the split rule) and stable-protection
// scenario (instability 0.05 or lower fires the protect rule) are the
// "LCOM4 of 4" and "Instability of 0.05" cases — and does not fire just
// inside the bound. Does-not-fire cases expect the all-clear, pinning
// the fallback alongside. Messages are asserted in full: they are
// contract.
func TestComputeRecommendations_ModuleRuleBoundaries(t *testing.T) {
	t.Parallel()

	// abstMod builds a module with an explicit exported-types count,
	// which mod fixes at 1; the abstractness rule needs values past 10.
	abstMod := func(path string, abst float64, exported int) metrics.ModuleResult {
		m := mod(path, 1, 1, 0.5, abst, 0.25, 1)
		m.ExportedTypes = exported
		return m
	}

	tests := []struct {
		name string
		mod  metrics.ModuleResult
		want []Recommendation
	}{
		{
			name: "LCOM4 of 4 fires the split rule",
			mod:  mod("pkg/a", 1, 1, 0.5, 0, 0.25, 4),
			want: []Recommendation{{
				Severity:   SeverityRed,
				Rule:       RuleLCOM4Split,
				ModulePath: "pkg/a",
				Message:    "Split `pkg/a` — LCOM4 of `4` suggests multiple unrelated responsibilities",
			}},
		},
		{
			name: "LCOM4 of 3 does not fire",
			mod:  mod("pkg/a", 1, 1, 0.5, 0, 0.25, 3),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "Distance of 0.5 fires the refactor rule",
			mod:  mod("pkg/a", 1, 1, 0.5, 0, 0.5, 1),
			want: []Recommendation{{
				Severity:   SeverityRed,
				Rule:       RuleDistanceRefactor,
				ModulePath: "pkg/a",
				Message:    "Refactor `pkg/a` — Distance of `0.50` from Main Sequence places it in `normal`",
			}},
		},
		{
			name: "Distance refactor names the module's zone",
			mod: func() metrics.ModuleResult {
				m := mod("pkg/a", 1, 1, 0.5, 0, 0.75, 1)
				m.Zone = metrics.ZoneOfPain
				return m
			}(),
			want: []Recommendation{{
				Severity:   SeverityRed,
				Rule:       RuleDistanceRefactor,
				ModulePath: "pkg/a",
				Message:    "Refactor `pkg/a` — Distance of `0.75` from Main Sequence places it in `zone-of-pain`",
			}},
		},
		{
			name: "Distance of 0.4999 does not fire",
			mod:  mod("pkg/a", 1, 1, 0.5, 0, 0.4999, 1),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "Instability of 0.95 fires the isolate rule",
			mod:  mod("pkg/a", 1, 1, 0.95, 0, 0.25, 1),
			want: []Recommendation{{
				Severity:   SeverityYellow,
				Rule:       RuleInstabilityIsolate,
				ModulePath: "pkg/a",
				Message:    "Isolate `pkg/a` — Instability of `0.95`; depends on everything, nothing depends on it; appropriate for CLI layers",
			}},
		},
		{
			name: "Instability of 0.9499 does not fire the isolate rule",
			mod:  mod("pkg/a", 1, 1, 0.9499, 0, 0.25, 1),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "Instability of 0.05 fires the protect rule",
			mod:  mod("pkg/a", 1, 1, 0.05, 0, 0.25, 1),
			want: []Recommendation{{
				Severity:   SeverityYellow,
				Rule:       RuleInstabilityProtect,
				ModulePath: "pkg/a",
				Message:    "Protect `pkg/a` — Instability of `0.05` makes it maximally stable; changes ripple widely",
			}},
		},
		{
			name: "Instability of 0.0501 does not fire the protect rule",
			mod:  mod("pkg/a", 1, 1, 0.0501, 0, 0.25, 1),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "Ce of 21 fires the decouple rule",
			mod:  mod("pkg/a", 1, 21, 0.5, 0, 0.25, 1),
			want: []Recommendation{{
				Severity:   SeverityYellow,
				Rule:       RuleCeDecouple,
				ModulePath: "pkg/a",
				Message:    "Decouple `pkg/a` — `21` efferent couplings; consider interface extraction",
			}},
		},
		{
			name: "Ce of 20 does not fire the decouple rule",
			mod:  mod("pkg/a", 1, 20, 0.5, 0, 0.25, 1),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "zero abstractness with 11 exported types fires the add-interfaces rule",
			mod:  abstMod("pkg/a", 0, 11),
			want: []Recommendation{{
				Severity:   SeverityYellow,
				Rule:       RuleAbstractnessInterfaces,
				ModulePath: "pkg/a",
				Message:    "Add interfaces to `pkg/a` — 0 abstractness with `11` exported types leaves no room for abstraction",
			}},
		},
		{
			name: "zero abstractness with 10 exported types does not fire",
			mod:  abstMod("pkg/a", 0, 10),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "nonzero abstractness with many exported types does not fire",
			mod:  abstMod("pkg/a", 0.1, 15),
			want: []Recommendation{wantAllClear},
		},
		{
			name: "zero abstractness with zero exported types does not fire",
			mod:  abstMod("pkg/a", 0, 0),
			want: []Recommendation{wantAllClear},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeRecommendations(completeGraph(t, tt.mod))
			assertRecommendations(t, got, tt.want)
		})
	}
}

// TestComputeRecommendations_CycleBreakRule pins the cycle-membership
// rule: every module participating in any circular dependency fires a
// red break recommendation whose message names its representative cycle,
// modules outside every cycle do not fire, and a module in several
// cycles cites its lexicographically smallest containing cycle
// regardless of the order the Cycles array lists them.
func TestComputeRecommendations_CycleBreakRule(t *testing.T) {
	t.Parallel()

	t.Run("every cycle member fires with the cycle's members", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 1),
		}
		g := mustGraph(t, metrics.StatusComplete, mods, []metrics.Cycle{{"pkg/a", "pkg/b"}}, nil)
		want := []Recommendation{
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/a",
				Message:    "Break cycle between `pkg/a`, `pkg/b` — circular dependencies prevent independent testing",
			},
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/b",
				Message:    "Break cycle between `pkg/a`, `pkg/b` — circular dependencies prevent independent testing",
			},
		}
		assertRecommendations(t, ComputeRecommendations(g), want)
	})

	t.Run("a module outside every cycle does not fire", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/c", 1, 1, 0.5, 0, 0.25, 1),
		}
		g := mustGraph(t, metrics.StatusComplete, mods, []metrics.Cycle{{"pkg/a", "pkg/b"}}, nil)
		want := []Recommendation{
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/a",
				Message:    "Break cycle between `pkg/a`, `pkg/b` — circular dependencies prevent independent testing",
			},
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/b",
				Message:    "Break cycle between `pkg/a`, `pkg/b` — circular dependencies prevent independent testing",
			},
		}
		assertRecommendations(t, ComputeRecommendations(g), want)
	})

	t.Run("a module in two cycles cites its smallest containing cycle", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/x", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/y", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/z", 1, 1, 0.5, 0, 0.25, 1),
		}
		// The larger cycle is listed first deliberately: the
		// representative choice must not depend on array order.
		cycles := []metrics.Cycle{{"pkg/x", "pkg/z"}, {"pkg/x", "pkg/y"}}
		g := mustGraph(t, metrics.StatusComplete, mods, cycles, nil)
		want := []Recommendation{
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/x",
				Message:    "Break cycle between `pkg/x`, `pkg/y` — circular dependencies prevent independent testing",
			},
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/y",
				Message:    "Break cycle between `pkg/x`, `pkg/y` — circular dependencies prevent independent testing",
			},
			{
				Severity:   SeverityRed,
				Rule:       RuleCycleBreak,
				ModulePath: "pkg/z",
				Message:    "Break cycle between `pkg/x`, `pkg/z` — circular dependencies prevent independent testing",
			},
		}
		assertRecommendations(t, ComputeRecommendations(g), want)
	})
}

// TestComputeRecommendations_DuplicationRuleBoundaries pins the two
// graph-scoped duplication rules: the red eliminate rule fires at
// exactly 5% overall duplication, the green watch rule fires strictly
// between 0% and 5% and suppresses the all-clear (the spec's
// watch-rule-only scenario), and 0% fires neither — including when
// blocks exist but total lines are zero. Messages cite the deduplicated
// block count, the highest group similarity, and the sorted file list;
// they are asserted in full as contract.
func TestComputeRecommendations_DuplicationRuleBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mods []metrics.ModuleResult
		want []Recommendation
	}{
		{
			name: "exactly 5 percent duplication fires the eliminate rule",
			mods: []metrics.ModuleResult{dupMod("pkg/a", 1000, metrics.Duplication{
				ModulePath: "pkg/a",
				Blocks: []metrics.DuplicateBlock{
					dupBlock("a.go", 10, 34, 25),
					dupBlock("b.go", 40, 64, 25),
				},
				Similarity: 1.0,
			})},
			want: []Recommendation{{
				Severity: SeverityRed,
				Rule:     RuleDuplicationEliminate,
				Message:  "Eliminate duplicate blocks — `2` blocks at `100`% similarity in `a.go`, `b.go`",
			}},
		},
		{
			name: "4 percent duplication fires the watch rule and suppresses the all-clear",
			mods: []metrics.ModuleResult{dupMod("pkg/a", 1000, metrics.Duplication{
				ModulePath: "pkg/a",
				Blocks: []metrics.DuplicateBlock{
					dupBlock("a.go", 10, 34, 20),
					dupBlock("b.go", 40, 64, 20),
				},
				Similarity: 0.95,
			})},
			want: []Recommendation{{
				Severity: SeverityGreen,
				Rule:     RuleDuplicationWatch,
				Message:  "Watch duplication in `a.go`, `b.go` — `2` blocks at `95`% similarity",
			}},
		},
		{
			name: "zero duplication fires neither rule",
			mods: []metrics.ModuleResult{dupMod("pkg/a", 1000)},
			want: []Recommendation{wantAllClear},
		},
		{
			name: "blocks with zero total lines compute 0 percent and fire neither rule",
			mods: []metrics.ModuleResult{dupMod("pkg/a", 0, metrics.Duplication{
				ModulePath: "pkg/a",
				Blocks:     []metrics.DuplicateBlock{dupBlock("a.go", 10, 25, 16)},
				Similarity: 1.0,
			})},
			want: []Recommendation{wantAllClear},
		},
		{
			name: "a block reported by two modules counts once in the message facts",
			mods: []metrics.ModuleResult{
				dupMod("pkg/a", 600, metrics.Duplication{
					ModulePath: "pkg/a",
					Blocks: []metrics.DuplicateBlock{
						dupBlock("shared.go", 10, 25, 30),
						dupBlock("a.go", 10, 34, 30),
					},
					Similarity: 0.9,
				}),
				dupMod("pkg/b", 400, metrics.Duplication{
					ModulePath: "pkg/b",
					Blocks:     []metrics.DuplicateBlock{dupBlock("shared.go", 10, 25, 30)},
					Similarity: 0.8,
				}),
			},
			want: []Recommendation{{
				Severity: SeverityRed,
				Rule:     RuleDuplicationEliminate,
				Message:  "Eliminate duplicate blocks — `2` blocks at `90`% similarity in `a.go`, `shared.go`",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeRecommendations(completeGraph(t, tt.mods...))
			assertRecommendations(t, got, tt.want)
		})
	}
}

// TestComputeRecommendations_AllClearFallback pins the spec's all-clear
// scenario: when no rule of any severity fires — including for the
// empty and nil graphs — exactly one recommendation, the green
// all-clear, is emitted.
func TestComputeRecommendations_AllClearFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		g    *metrics.ModuleGraph
	}{
		{name: "nil graph", g: nil},
		{name: "zero modules", g: completeGraph(t)},
		{name: "clean module", g: completeGraph(t, mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRecommendations(t, ComputeRecommendations(tt.g), []Recommendation{wantAllClear})
		})
	}
}

// TestComputeRecommendations_Ordering pins the ordering contract:
// severity before extremity before path (the spec's severity-precedence
// and deterministic-tiebreak scenarios), extremity descending within a
// severity including numeric comparison across rules, and the
// rule-identifier tiebreak for two rules firing on one module with
// equal extremity.
func TestComputeRecommendations_Ordering(t *testing.T) {
	t.Parallel()

	t.Run("severity precedence: a red cycle hit renders before a yellow decouple hit", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 1),
			mod("pkg/c", 1, 25, 0.5, 0, 0.25, 1),
		}
		g := mustGraph(t, metrics.StatusComplete, mods, []metrics.Cycle{{"pkg/a", "pkg/b"}}, nil)
		want := []Recommendation{
			{Rule: RuleCycleBreak, ModulePath: "pkg/a"},
			{Rule: RuleCycleBreak, ModulePath: "pkg/b"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/c"},
		}
		assertRuleOrder(t, ComputeRecommendations(g), want)
	})

	t.Run("extremity orders within a severity, most extreme first", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 6),
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 4),
			mod("pkg/c", 1, 1, 0.5, 0, 0.25, 5),
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/c"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/b"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})

	t.Run("extremity compares numerically across rules in one severity", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 6), // red, extremity 2
			mod("pkg/b", 1, 1, 0.5, 0, 0.9, 1),  // red, extremity 0.4
			mod("pkg/c", 1, 1, 0.5, 0, 0.5, 1),  // red, extremity 0
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleDistanceRefactor, ModulePath: "pkg/b"},
			{Rule: RuleDistanceRefactor, ModulePath: "pkg/c"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})

	t.Run("equal extremity on the same rule ties break on path ascending", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 4),
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 4),
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/b"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})

	t.Run("equal extremity on one module ties break on rule ascending", func(t *testing.T) {
		t.Parallel()
		// Ce 21 and 11 exported types both sit exactly 1 past their
		// bound, so the two yellow hits tie on severity, extremity,
		// and path; the rule identifier decides.
		m := mod("pkg/a", 1, 21, 0.5, 0, 0.25, 1)
		m.ExportedTypes = 11
		want := []Recommendation{
			{Rule: RuleAbstractnessInterfaces, ModulePath: "pkg/a"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/a"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, m)), want)
	})
}

// TestComputeRecommendations_Cap pins the spec's rule-cap scenario: when
// more than five recommendations fire, at most five render, selected by
// the ordering contract — severity winners first, then the most extreme,
// then path order.
func TestComputeRecommendations_Cap(t *testing.T) {
	t.Parallel()

	t.Run("seven split violations return the five most extreme", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 10),
			mod("pkg/b", 1, 1, 0.5, 0, 0.25, 9),
			mod("pkg/c", 1, 1, 0.5, 0, 0.25, 8),
			mod("pkg/d", 1, 1, 0.5, 0, 0.25, 7),
			mod("pkg/e", 1, 1, 0.5, 0, 0.25, 6),
			mod("pkg/f", 1, 1, 0.5, 0, 0.25, 5),
			mod("pkg/g", 1, 1, 0.5, 0, 0.25, 4),
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/b"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/c"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/d"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/e"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})

	t.Run("equal extremity keeps the five lexicographically first paths", func(t *testing.T) {
		t.Parallel()
		paths := []string{"pkg/a", "pkg/b", "pkg/c", "pkg/d", "pkg/e", "pkg/f", "pkg/g"}
		mods := make([]metrics.ModuleResult, len(paths))
		for i, path := range paths {
			mods[i] = mod(path, 1, 1, 0.5, 0, 0.25, 4)
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/b"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/c"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/d"},
			{Rule: RuleLCOM4Split, ModulePath: "pkg/e"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})

	t.Run("the cap keeps a low-extremity red hit ahead of high-extremity yellow hits", func(t *testing.T) {
		t.Parallel()
		mods := []metrics.ModuleResult{
			mod("pkg/a", 1, 1, 0.5, 0, 0.25, 4),  // red, extremity 0
			mod("pkg/b", 1, 26, 0.5, 0, 0.25, 1), // yellow, extremity 6
			mod("pkg/c", 1, 25, 0.5, 0, 0.25, 1), // yellow, extremity 5
			mod("pkg/d", 1, 24, 0.5, 0, 0.25, 1), // yellow, extremity 4
			mod("pkg/e", 1, 23, 0.5, 0, 0.25, 1), // yellow, extremity 3
			mod("pkg/f", 1, 22, 0.5, 0, 0.25, 1), // yellow, extremity 2
			mod("pkg/g", 1, 21, 0.5, 0, 0.25, 1), // yellow, extremity 1
		}
		want := []Recommendation{
			{Rule: RuleLCOM4Split, ModulePath: "pkg/a"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/b"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/c"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/d"},
			{Rule: RuleCeDecouple, ModulePath: "pkg/e"},
		}
		assertRuleOrder(t, ComputeRecommendations(completeGraph(t, mods...)), want)
	})
}

// TestComputeRecommendations_InputOrderIndependence pins the report's
// byte-stability requirement at the rule-engine level: reordering the
// Modules array and the Cycles array must not change the computed
// recommendations. The fixture fires every rule family through the sort
// and the cap.
func TestComputeRecommendations_InputOrderIndependence(t *testing.T) {
	t.Parallel()

	a := mod("pkg/a", 1, 1, 0.5, 0, 0.25, 7)
	b := mod("pkg/b", 0, 25, 0.97, 0, 0.6, 1)
	c := mod("pkg/c", 9, 0, 0.0, 0, 0.3, 4)
	d := mod("pkg/d", 1, 1, 0.5, 0, 0.25, 1)
	d.ExportedTypes = 12
	e := dupMod("pkg/e", 1000, metrics.Duplication{
		ModulePath: "pkg/e",
		Blocks:     []metrics.DuplicateBlock{dupBlock("e.go", 10, 34, 40)},
		Similarity: 0.9,
	})
	mods := []metrics.ModuleResult{a, b, c, d, e}
	cycles := []metrics.Cycle{{"pkg/a", "pkg/d"}, {"pkg/c", "pkg/e"}}

	forward := mustGraph(t, metrics.StatusComplete, mods, cycles, nil)

	reversed := make([]metrics.ModuleResult, len(mods))
	for i, m := range mods {
		reversed[len(mods)-1-i] = m
	}
	reorderedCycles := []metrics.Cycle{cycles[1], cycles[0]}
	backward := mustGraph(t, metrics.StatusComplete, reversed, reorderedCycles, nil)

	want := ComputeRecommendations(forward)
	if got := ComputeRecommendations(backward); !equalRecommendations(got, want) {
		t.Errorf("ComputeRecommendations on reordered input: got %+v, want %+v", got, want)
	}
}

// equalRecommendations reports whether two recommendation slices are
// element-by-element equal.
func equalRecommendations(a, b []Recommendation) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
