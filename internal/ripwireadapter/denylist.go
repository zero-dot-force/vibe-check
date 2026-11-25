package ripwireadapter

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// crawlDenylistNames are the directory names ripwire prunes from a crawl
// (validated against the pinned ripwire release). The adapter's pre-check
// mirrors this set so users understand why source code they can see
// contributes no metrics; ripwire's own <pruned_packages> disclosure is
// unioned in (openspec/changes/archive/2026-11-25-bump-ripwire-pin design D5) so an upstream
// denylist change the list has not caught up with still surfaces for real
// packages. A name is a candidate, not a verdict: the pinned release indexes
// a build/ directory whose files git tracks, so the pre-check drops any
// candidate ripwire reported a module at or beneath.
var crawlDenylistNames = []string{
	"build",
	"dist",
	"out",
	"target",
	"vendor",
	"third_party",
	"captures",
}

// denylistSourceExtensions are the file extensions recognized as source
// files for the languages this adapter analyzes. The set is the adapter's
// shared definition of a source file, used by the denylist pre-check (a
// directory holding none of these extensions is invisible to ripwire's
// pruner anyway — nothing to prune — so it does not warrant a warning),
// by guard.go's source census, and by parse_cc.go's totalLines filter.
var denylistSourceExtensions = map[string]bool{
	".py":  true,
	".ts":  true,
	".tsx": true,
	".js":  true,
	".jsx": true,
}

// warningDenylistPruned is the warning code the goldens pin for the denylist
// pre-check.
const warningDenylistPruned = "denylist-pruned"

// detectDenylistPruned reports the source directories ripwire prunes from
// the crawl: the union of the denylist-named, source-bearing directories
// found by walking the crawl root that ripwire did not index (indexedDirs:
// the --arch module paths, crawl-root-relative and slash-separated; a
// candidate with a module at or beneath it was crawled, so it is neither
// reported nor skipped, and nested candidates under it are still checked)
// and the package paths ripwire itself
// lists under <pruned_packages> (ripwirePruned, crawl-root-relative,
// slash-separated; ripwireCount is ripwire's exact count=, which may
// exceed the listed paths because ripwire caps the listing). It returns at
// most one Warning (code "denylist-pruned") whose message lists every
// directory once, root-relative, slash-separated, and sorted, and states
// ripwire's total when its listing was partial; with no directory from
// either source it returns no warnings. A warning is purely informational —
// ripwire prunes these directories regardless, and the analysis proceeds.
//
// Source-bearing means the directory subtree contains at least one file
// with a language extension (recursive: a build/ dir whose sources live one
// level down still gets pruned and still deserves the warning). Unreadable
// subtrees are skipped rather than failing the analysis — one locked
// directory in a user tree must not abort a crawl — while every other walk
// error fails closed as a wrapped error.
func detectDenylistPruned(crawlRoot string, indexedDirs, ripwirePruned []string, ripwireCount int) ([]metrics.Warning, error) {
	info, err := os.Stat(crawlRoot)
	if err != nil {
		return nil, fmt.Errorf("denylist pre-check: stat %s: %w", crawlRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("denylist pre-check: %s: %w", crawlRoot, errNotDirectory)
	}

	deny := make(map[string]bool, len(crawlDenylistNames))
	for _, name := range crawlDenylistNames {
		deny[name] = true
	}

	var pruned []string
	err = filepath.WalkDir(crawlRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable subtrees; the crawl itself will surface
			// any real failure.
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if !deny[base] {
			return nil
		}
		if path == crawlRoot {
			// The crawl root itself being denylist-named is a user
			// scoping choice, not a pruning surprise.
			return nil
		}
		rel := filepath.ToSlash(relativeToRoot(crawlRoot, path))
		if indexedAtOrBeneath(rel, indexedDirs) {
			// ripwire crawled it (a git-tracked build/): not pruned.
			return nil
		}
		if dirContainsSource(path) {
			pruned = append(pruned, rel)
			// Do not descend: everything beneath is pruned with it,
			// and nested denylist names under a pruned dir are
			// noise.
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("denylist pre-check: walk %s: %w", crawlRoot, err)
	}
	return denylistPrunedWarning(pruned, ripwirePruned, ripwireCount), nil
}

// denylistPrunedWarning builds the denylist-pruned warning from the
// pre-check's directories and ripwire's disclosure: the sorted,
// de-duplicated union, plus ripwire's total when it listed fewer paths than
// it counted. Both inputs empty yields no warning.
func denylistPrunedWarning(preCheck, ripwirePruned []string, ripwireCount int) []metrics.Warning {
	seen := make(map[string]bool, len(preCheck)+len(ripwirePruned))
	union := make([]string, 0, len(preCheck)+len(ripwirePruned))
	for _, dir := range append(append([]string(nil), preCheck...), ripwirePruned...) {
		if !seen[dir] {
			seen[dir] = true
			union = append(union, dir)
		}
	}
	if len(union) == 0 {
		return nil
	}
	sort.Strings(union)

	message := fmt.Sprintf(
		"ripwire prunes source-bearing denylist directories from the crawl: %s",
		strings.Join(union, ", "),
	)
	if ripwireCount > len(ripwirePruned) {
		message += fmt.Sprintf(
			"; ripwire reports %d pruned packages in total and lists only %d (run ripwire --skipped for all)",
			ripwireCount, len(ripwirePruned),
		)
	}
	return []metrics.Warning{{Code: warningDenylistPruned, Message: message}}
}

// indexedAtOrBeneath reports whether any indexed module path equals dir or
// lies beneath it. Paths are compared slash-separated with any leading "./"
// removed, since the mapping accepts both row path forms.
func indexedAtOrBeneath(dir string, indexedDirs []string) bool {
	for _, p := range indexedDirs {
		p = strings.TrimPrefix(p, "./")
		if p == dir || strings.HasPrefix(p, dir+"/") {
			return true
		}
	}
	return false
}

// errNotDirectory reports a crawl root that exists but is not a directory.
var errNotDirectory = errors.New("not a directory")

// dirContainsSource reports whether the directory tree rooted at dir holds
// at least one file with a language extension recognized by the denylist
// pre-check.
func dirContainsSource(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || found {
			return fs.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		if denylistSourceExtensions[strings.ToLower(filepath.Ext(d.Name()))] {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// relativeToRoot renders path relative to root when possible, falling back
// to the base name for cross-device oddities.
func relativeToRoot(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return filepath.Base(path)
}
