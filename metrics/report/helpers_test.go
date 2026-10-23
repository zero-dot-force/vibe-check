package report

import (
	"encoding/json"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// mod builds a metrics.ModuleResult with explicit metric values. It mirrors
// the unexported helper of the same name in the metrics package's own tests:
// the report engine's tests cannot reach across the package boundary to
// reuse it, so the fixture builder travels with the tests that consume it.
// Fixtures set the metric fields directly to exercise rule and grade
// boundaries; they are NOT required to satisfy the D = |A + I - 1| identity,
// which metrics.Validate does not enforce — Validate checks only that ratio
// metrics fall in [0.0, 1.0] and that counts are non-negative.
func mod(path string, ca, ce int, inst, abst, dist float64, lcom int) metrics.ModuleResult {
	return metrics.ModuleResult{
		Module: metrics.Module{
			Path:          path,
			Name:          path,
			Ca:            ca,
			Ce:            ce,
			ExportedTypes: 1,
			AbstractTypes: 0,
		},
		Instability:  metrics.Instability(inst),
		Abstractness: metrics.Abstractness(abst),
		Distance:     metrics.Distance(dist),
		LCOM:         metrics.LCOM(lcom),
		Zone:         metrics.ZoneNormal,
	}
}

// mustGraph builds a schema-valid metrics.ModuleGraph fixture and fails the
// test if it does not pass metrics.Validate, mirroring the metrics package's
// own mustGraph so every report fixture is validated at construction.
func mustGraph(t *testing.T, status metrics.Status, mods []metrics.ModuleResult, cycles []metrics.Cycle, warnings []metrics.Warning) *metrics.ModuleGraph {
	t.Helper()
	if mods == nil {
		mods = []metrics.ModuleResult{}
	}
	if cycles == nil {
		cycles = []metrics.Cycle{}
	}
	if warnings == nil {
		warnings = []metrics.Warning{}
	}
	g := &metrics.ModuleGraph{
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
	if err := metrics.Validate(data); err != nil {
		t.Fatalf("fixture failed schema validation: %v", err)
	}
	return g
}

// completeGraph builds a schema-valid, complete-status graph from the given
// modules with no cycles and no warnings, mirroring the metrics package's
// own completeGraph.
func completeGraph(t *testing.T, mods ...metrics.ModuleResult) *metrics.ModuleGraph {
	t.Helper()
	return mustGraph(t, metrics.StatusComplete, mods, nil, nil)
}
