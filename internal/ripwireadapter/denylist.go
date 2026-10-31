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

// crawlDenylistNames are the directory names ripwire hard-prunes from every
// crawl (validated against ripwire 0.6.9 in task 1.2: a fixture
// operations/build directory is absent from all recorded outputs). The
// adapter's pre-check mirrors this set so users understand why source code
// they can see contributes no metrics.
var crawlDenylistNames = []string{
	"build",
	"dist",
	"out",
	"target",
	"vendor",
	"third_party",
	"captures",
}

// denylistSourceExtensions are the file extensions that make a directory
// "source-bearing" for the denylist pre-check. The set covers the languages
// this adapter analyzes; a directory holding none of them is invisible to
// ripwire's pruner anyway (nothing to prune), so it does not warrant a
// warning.
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

// detectDenylistPruned walks the crawl root and reports the denylist-named,
// source-bearing directories that ripwire will prune from the crawl. It
// returns at most one Warning (code "denylist-pruned") whose message lists
// every matched directory, root-relative and sorted; a clean tree yields no
// warnings. A warning is purely informational — ripwire prunes these
// directories regardless, and the analysis proceeds.
//
// Source-bearing means the directory subtree contains at least one file
// with a language extension (recursive: a build/ dir whose sources live one
// level down still gets pruned and still deserves the warning). Unreadable
// subtrees are skipped rather than failing the analysis — one locked
// directory in a user tree must not abort a crawl — while every other walk
// error fails closed as a wrapped error.
func detectDenylistPruned(crawlRoot string) ([]metrics.Warning, error) {
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
		if dirContainsSource(path) {
			pruned = append(pruned, relativeToRoot(crawlRoot, path))
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
	if len(pruned) == 0 {
		return nil, nil
	}
	sort.Strings(pruned)

	warning := metrics.Warning{
		Code: warningDenylistPruned,
		Message: fmt.Sprintf(
			"ripwire prunes source-bearing denylist directories from the crawl: %s",
			strings.Join(pruned, ", "),
		),
	}
	return []metrics.Warning{warning}, nil
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
