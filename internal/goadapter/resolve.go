package goadapter

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// loadFlags defines the information requested from go/packages.
// NeedName: package name and path.
// NeedImports: import map for coupling analysis.
// NeedTypes: type information for abstractness and LCOM4.
// NeedSyntax: AST for LCOM4 field-access analysis.
// NeedTypesInfo: resolved type info for identifier resolution.
// NeedModule: module metadata for stdlib detection and module path.
const loadFlags = packages.NeedName |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedModule

// packageEnvAllowlist enumerates the environment variables passed through to
// go/packages (and any subprocess it spawns) via metrics.SanitizeEnvironment.
// Only variables required to resolve modules and locate the build cache are
// included. Credential-bearing and command-injection vectors — notably
// GOFLAGS — are deliberately excluded. This is an immutable package-level
// slice, not mutable global state.
var packageEnvAllowlist = []string{
	"GOPATH",
	"GOROOT",
	"GOMODCACHE",
	"GOPROXY",
	"GONOSUMCHECK",
	"GOMOD",
}

// packageEnv builds the environment for the go/packages subprocess. It starts
// from the sanitized allowlist ([metrics.SanitizeEnvironment]) and appends
// "GOTOOLCHAIN=local" as the LAST entry so it wins: on exec a duplicated key is
// resolved last-wins, forcing the local toolchain regardless of the host
// environment. This blocks a target module's go.mod "toolchain" directive from
// triggering a Go toolchain download and execution during subprocess package
// loading. GOTOOLCHAIN is intentionally NOT in packageEnvAllowlist, so no host
// value ever passes through — the appended value is authoritative.
func packageEnv() []string {
	return append(metrics.SanitizeEnvironment(packageEnvAllowlist), "GOTOOLCHAIN=local")
}

// resolvePackages loads Go packages from projectPath using go/packages and
// builds the module-internal import adjacency map used to compute Ca/Ce. It
// filters results to module-internal packages and records warnings for any
// package that fails to load or type-check.
//
// Returns:
//   - pkgs: loaded module-internal packages (may include packages with errors)
//   - imports: adjacency map of module-internal import edges (pkg → its module-internal imports)
//   - warnings: any non-fatal issues encountered during loading
//   - err: fatal error if package loading fails entirely
func resolvePackages(ctx context.Context, projectPath string) (
	pkgs []*packages.Package,
	imports map[string][]string,
	warnings []metrics.Warning,
	err error,
) {
	cfg := &packages.Config{
		Mode:    loadFlags,
		Dir:     projectPath,
		Context: ctx,
		Env:     packageEnv(),
		Tests:   false,
	}

	allPkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("resolve packages: %w — verify the path is a valid Go module directory containing go.mod and .go files", err)
	}

	if len(allPkgs) == 0 {
		return nil, nil, nil, nil
	}

	// Determine module path from the first package with module info.
	modulePath := detectModulePath(allPkgs)
	if modulePath == "" {
		return nil, nil, nil, fmt.Errorf("resolve packages: unable to determine module path — ensure the target directory contains a go.mod (run 'go mod init' if missing)")
	}

	// Filter to module-internal packages and collect warnings for errored packages.
	var internal []*packages.Package
	internalSet := make(map[string]bool)

	for _, pkg := range allPkgs {
		if !strings.HasPrefix(pkg.PkgPath, modulePath) {
			continue
		}
		internalSet[pkg.PkgPath] = true
		internal = append(internal, pkg)

		relPath := relativeModulePath(pkg.PkgPath, modulePath)

		// Collect load/type errors as warnings rather than failing.
		for _, pkgErr := range pkg.Errors {
			warnings = append(warnings, metrics.Warning{
				Code:    "load-error",
				Message: fmt.Sprintf("package %s: %s", relPath, pkgErr.Msg),
				Module:  pkg.PkgPath,
			})
		}

		// A package with no reported errors but nil type information could not
		// be type-checked. Record a warning so its zeroed type metrics (emitted
		// by the adapter per the go-adapter partial-build scenario) are
		// explained to consumers.
		if len(pkg.Errors) == 0 && pkg.Types == nil {
			warnings = append(warnings, metrics.Warning{
				Code:    "load-error",
				Message: fmt.Sprintf("package %s: type information unavailable (type-checking did not complete)", relPath),
				Module:  pkg.PkgPath,
			})
		}
	}

	// Build import adjacency map: for each internal package, record which
	// other internal packages it imports (module-internal edges only).
	imports = make(map[string][]string, len(internal))
	for _, pkg := range internal {
		var internalImports []string
		for impPath := range pkg.Imports {
			if internalSet[impPath] {
				internalImports = append(internalImports, impPath)
			}
		}
		imports[pkg.PkgPath] = internalImports
	}

	return internal, imports, warnings, nil
}

// anyPackageTypeChecked reports whether at least one package in pkgs was
// successfully type-checked — that is, it has no load/type errors and non-nil
// type information. It distinguishes a partial build (at least one usable
// package, so analysis proceeds with warnings for the rest) from a total load
// failure (no usable package, so Analyze must return an error rather than a
// graph of all-zeroed modules). The predicate mirrors the per-package guard in
// Adapter.Analyze that gates type classification and LCOM computation.
func anyPackageTypeChecked(pkgs []*packages.Package) bool {
	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 && pkg.Types != nil {
			return true
		}
	}
	return false
}

// countCe counts the efferent coupling for a package: the number of distinct
// imports including standard library, third-party, and module-internal packages.
// Per Robert C. Martin's definition, Ce counts all outgoing dependencies
// regardless of their origin. Stdlib and third-party packages are excluded
// from the module list but still contribute to Ce counts.
func countCe(pkg *packages.Package) int {
	return len(pkg.Imports)
}

// countCa counts the afferent coupling for a package: the number of
// module-internal packages that import it.
func countCa(pkgPath string, imports map[string][]string) int {
	count := 0
	for _, deps := range imports {
		for _, dep := range deps {
			if dep == pkgPath {
				count++
				break
			}
		}
	}
	return count
}

// detectModulePath extracts the module path from loaded packages.
// It uses the Module field from the first package that has one.
func detectModulePath(pkgs []*packages.Package) string {
	for _, pkg := range pkgs {
		if pkg.Module != nil {
			return pkg.Module.Path
		}
	}
	return ""
}

// relativeModulePath returns the package path relative to the module root.
// For example, "example.com/foo/bar/baz" with module "example.com/foo" returns "bar/baz".
// If the package is the module root, returns ".".
func relativeModulePath(pkgPath, modulePath string) string {
	rel := strings.TrimPrefix(pkgPath, modulePath)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return "."
	}
	return rel
}
