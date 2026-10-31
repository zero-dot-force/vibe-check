package ripwireadapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// warningZeroCouplingGuard is the warning code pinned for the zero-coupling
// guard.
const warningZeroCouplingGuard = "zero-coupling-guard"

// zeroCouplingGuard defends against the silent-zero failure mode of a
// misconfigured ripwire crawl. When ripwire is pointed at the wrong root,
// when a Python project's pip src-layout defeats dependency resolution, or
// when the wrong language is crawled, the crawl still succeeds and emits a
// structurally valid --deps XML; the resulting ModuleGraph, however, is all
// zeros: no dependency edges, every module fully disconnected. That graph
// looks valid while carrying no information, so the adapter must refuse to
// present it as a complete result (openspec/changes/ripwire-adapter
// design.md D9; the spec's zero-coupling result guard requirement).
//
// The guard is a pure predicate over the two numbers that graph assembly
// supplies: edgeCount is the number of dependency edges parsed from the
// --deps XML, and sourceFileCount is the number of source files in the
// crawl scope (see countSourceFiles). It fires exactly when edgeCount == 0
// and sourceFileCount > 0, returning a single global metrics.Warning (empty
// Module) with code "zero-coupling-guard" whose message names both counts.
// It returns nil in every other case: an empty corpus legitimately yields
// an edgeless graph, and any non-zero edge count proves the crawl resolved
// real dependencies.
//
// Wiring point: nothing calls the guard yet. Task 6.2 (Analyze graph
// assembly) must call it with the real parsed numbers and, when it returns
// a non-empty slice, append the warnings to the ModuleGraph and force the
// graph status to metrics.StatusPartial instead of metrics.StatusComplete.
// The design deliberately chose warning + partial over a hard error so
// users still receive the graph with the caveat attached.
func zeroCouplingGuard(edgeCount, sourceFileCount int) []metrics.Warning {
	// Complement of the fire condition (edgeCount == 0 && sourceFileCount > 0):
	// resolved edges, an empty corpus, or a nonsensical negative count all
	// leave the guard silent.
	if edgeCount != 0 || sourceFileCount <= 0 {
		return nil
	}
	return []metrics.Warning{{
		Code: warningZeroCouplingGuard,
		Message: fmt.Sprintf(
			"crawl reported %d dependency edges across %d source files in scope; "+
				"a non-empty corpus with zero edges usually indicates a misconfigured "+
				"crawl (wrong root, pip src-layout resolution failure, or wrong language)",
			edgeCount, sourceFileCount,
		),
	}}
}

// countSourceFiles walks root recursively and returns the number of files
// whose extension belongs to the adapter's language set
// (denylistSourceExtensions in denylist.go: .py, .ts, .tsx, .js, .jsx).
// It supplies zeroCouplingGuard's sourceFileCount input, which is what
// distinguishes "nothing was in scope" from "a real corpus yielded no
// edges".
//
// The walk mirrors the denylist pre-check's error posture: a missing or
// non-directory root fails closed as a wrapped error, because a misresolved
// crawl root must not silently return zero and disengage the guard, while
// unreadable subtrees are skipped because one locked directory in a user
// tree must not abort the analysis. The walk intentionally does not apply
// the crawl denylist itself: over-counting can only make the guard fire
// more often, which is the safe direction for a safety net that emits a
// caveat rather than an error.
func countSourceFiles(root string) (int, error) {
	info, err := os.Stat(root)
	if err != nil {
		return 0, fmt.Errorf("zero-coupling guard: stat %s: %w", root, err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("zero-coupling guard: %s: %w", root, errNotDirectory)
	}

	count := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable subtrees, mirroring the denylist pre-check:
			// one locked directory must not abort the analysis, and the
			// crawl itself surfaces real read failures.
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if denylistSourceExtensions[strings.ToLower(filepath.Ext(d.Name()))] {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("zero-coupling guard: walk %s: %w", root, err)
	}
	return count, nil
}
