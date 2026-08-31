package goadapter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// errTotalLoadFailure indicates that packages were found in the target
// directory but none could be type-checked (every package has load/type
// errors or nil type information). It is returned — wrapped with actionable
// remediation via %w — instead of a graph of all-zeroed modules, per the
// go-adapter total-load-failure scenario. Emitting all-zero metrics as if
// real, when nothing could be analyzed, would violate Metric Fidelity.
// Callers may test for it with errors.Is.
var errTotalLoadFailure = errors.New("total load failure")

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

// Capabilities returns all universal metrics this adapter can compute.
// The Go adapter supports the complete metrics suite. Go-specific extension
// capabilities are reported separately by [Adapter.ExtensionCapabilities].
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

// ExtensionCapabilities returns the optional Go-specific extension capabilities
// this adapter populates in each ModuleResult's Extensions map. These are
// declared separately from the universal [Adapter.Capabilities] because they
// are language-specific and namespaced under the "go." prefix. The returned
// identifiers double as the Extensions map keys.
func (a *Adapter) ExtensionCapabilities() []string {
	return []string{CapInterfaceWidth, CapInterfaceProximity}
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
//   - no package can be type-checked (total load failure)
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
	pkgs, imports, warnings, err := resolvePackages(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("analyze: no Go packages found in %s — ensure the path is a Go module directory containing .go files and a go.mod", projectPath)
	}

	// Per the go-adapter total-load-failure scenario: packages were found, but
	// if NONE could be type-checked (every package has load/type errors or nil
	// type information) the load failed entirely. Returning a graph of
	// all-zeroed modules would present fabricated metrics as if real, violating
	// Metric Fidelity — return an error instead. The partial case (at least one
	// package type-checks) is handled below by emitting zeroed ModuleResults
	// plus warnings for the individual errored packages.
	if !anyPackageTypeChecked(pkgs) {
		return nil, fmt.Errorf("analyze: %w: none of the %d package(s) in %s could be type-checked — run 'go build ./...' to see the underlying errors, then ensure all dependencies are available (e.g. run 'go mod download')", errTotalLoadFailure, len(pkgs), projectPath)
	}

	// Build set of module-internal package paths for cycle detection.
	modulePkgs := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		modulePkgs[pkg.PkgPath] = true
	}

	// Step 4: Compute metrics for each package.
	//
	// Per the go-adapter partial-build scenario, packages that fail to load or
	// type-check are NOT dropped. We emit a zeroed ModuleResult for them
	// (ExportedTypes=0, AbstractTypes=0, LCOM=0) so they still appear in the
	// graph, while resolvePackages records a corresponding warning. Coupling
	// (Ca/Ce) is still computed from the import graph, which is safe even when
	// type-checking is incomplete.
	var modules []metrics.ModuleResult
	for _, pkg := range pkgs {
		// Coupling metrics are always safe to compute from the import graph.
		ca := countCa(pkg.PkgPath, imports)
		ce := countCe(pkg)

		// Type classification and LCOM require complete type information.
		// Packages with load/type errors or nil types get zeroed values; a
		// warning for them has already been recorded in resolvePackages.
		var exportedTypes, abstractTypes int
		var lcom metrics.LCOM
		var extensions map[string]any
		if len(pkg.Errors) == 0 && pkg.Types != nil {
			exportedTypes, abstractTypes = countTypes(pkg)
			lcom = computeLCOM4(pkg)
			extensions = computeExtensions(pkg)
		}

		// Derived metrics via metrics.Compute* functions.
		instability := metrics.ComputeInstability(ca, ce)
		abstractness := metrics.ComputeAbstractness(abstractTypes, exportedTypes)
		distance := metrics.ComputeDistance(abstractness, instability)
		zone := metrics.ComputeZone(abstractness, instability, distance)

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
