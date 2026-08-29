package goadapter

import (
	"context"
	"fmt"
	"sort"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Compile-time interface compliance check.
var _ metrics.Adapter = (*Adapter)(nil)

// Adapter implements [metrics.Adapter] for Go codebases. It uses
// [golang.org/x/tools/go/packages] to load and analyze Go packages,
// computing the full Martin metrics suite plus Go-specific extensions.
type Adapter struct{}

// New creates a new Go language adapter.
func New() *Adapter {
	return &Adapter{}
}

// Language returns "go" as the language identifier.
func (a *Adapter) Language() string {
	return "go"
}

// Capabilities returns all metrics this adapter can compute.
// The Go adapter supports the complete metrics suite.
func (a *Adapter) Capabilities() []metrics.Capability {
	return []metrics.Capability{
		metrics.CapAfferentCoupling,
		metrics.CapEfferentCoupling,
		metrics.CapInstability,
		metrics.CapAbstractness,
		metrics.CapDistance,
		metrics.CapLCOM,
		metrics.CapCircularDeps,
	}
}

// Analyze performs a complete analysis of the Go project at projectPath.
// It loads all packages in the module, computes coupling metrics (Ca, Ce),
// type classification (exported/abstract), LCOM4 cohesion, derived metrics
// (instability, abstractness, distance, zone), and detects circular dependencies.
//
// The returned [metrics.ModuleGraph] conforms to schema version "1.1".
// Status is [metrics.StatusComplete] when all packages load without errors,
// or [metrics.StatusPartial] when some packages have errors but analysis
// can still proceed.
//
// Returns an error if:
//   - projectPath fails validation
//   - the context is cancelled or expired
//   - no Go packages are found in the project
//   - the module path cannot be determined
func (a *Adapter) Analyze(ctx context.Context, projectPath string) (*metrics.ModuleGraph, error) {
	// Step 1: Validate project path.
	if err := metrics.ValidateProjectPath(projectPath); err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	// Step 2: Check context before expensive operations.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	// Step 3: Load and resolve packages.
	pkgs, _, imports, warnings, err := resolvePackages(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("analyze: no Go packages found in %s", projectPath)
	}

	// Build set of module-internal package paths for cycle detection.
	modulePkgs := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		modulePkgs[pkg.PkgPath] = true
	}

	// Step 4: Compute metrics for each package.
	var modules []metrics.ModuleResult
	for _, pkg := range pkgs {
		// Skip packages with errors for detailed analysis but include
		// them in coupling counts (they're still in the import graph).
		if len(pkg.Errors) > 0 {
			continue
		}

		// Raw metrics.
		ca := countCa(pkg.PkgPath, imports)
		ce := countCe(pkg)
		exportedTypes, abstractTypes := countTypes(pkg)
		lcom := computeLCOM4(pkg)

		// Derived metrics via metrics.Compute* functions.
		instability := metrics.ComputeInstability(ca, ce)
		abstractness := metrics.ComputeAbstractness(abstractTypes, exportedTypes)
		distance := metrics.ComputeDistance(abstractness, instability)
		zone := metrics.ComputeZone(abstractness, instability, distance)

		// Go-specific extensions.
		extensions := computeExtensions(pkg)

		modules = append(modules, metrics.ModuleResult{
			Module: metrics.Module{
				Path:          pkg.PkgPath,
				Name:          pkg.Name,
				Ca:            ca,
				Ce:            ce,
				ExportedTypes: exportedTypes,
				AbstractTypes: abstractTypes,
			},
			Instability:  instability,
			Abstractness: abstractness,
			Distance:     distance,
			LCOM:         lcom,
			Zone:         zone,
			Extensions:   extensions,
		})
	}

	// Step 5: Detect circular dependencies.
	cycles := detectCycles(imports, modulePkgs)

	// Step 6: Sort modules by Path for deterministic output.
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	// Ensure non-nil slices per ModuleGraph contract.
	if warnings == nil {
		warnings = []metrics.Warning{}
	}
	if modules == nil {
		modules = []metrics.ModuleResult{}
	}

	// Step 7: Determine status.
	status := metrics.StatusComplete
	if len(warnings) > 0 {
		status = metrics.StatusPartial
	}

	return &metrics.ModuleGraph{
		SchemaVersion: metrics.SchemaVersionCurrent,
		Language:      "go",
		Modules:       modules,
		Cycles:        cycles,
		Warnings:      warnings,
		Status:        status,
	}, nil
}
