package ripwireadapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// crawlScope is the resolved scope of a ripwire crawl: the package-root
// directory the crawl starts from, plus the subtrees pruned from it.
//
// The adapter never crawls from a repository root that merely contains
// nested projects; the root is always a package root resolved by
// resolvePythonCrawlRoot or resolveTypeScriptCrawlRoot. Excludes exist so a
// user pointing the crawl at a larger tree can scope it down.
type crawlScope struct {
	// root is the directory the crawl starts from, in whatever form
	// projectPath was given. It is passed to ripwireRunner.run as the
	// working directory, mirroring the capture convention where recordings
	// ran from python/src and typescript/ directly.
	root string

	// excludes are root-relative subtree paths pruned from the crawl via
	// ripwire's --exclude flag, in the order given. Callers must pass
	// non-empty relative paths; validating user input is the CLI wiring's
	// job, not the scope's.
	excludes []string
}

// newCrawlScope returns a crawl scope starting at root with the given
// root-relative exclude paths. The exclude slice is copied so that later
// mutation by the caller cannot change the scope.
func newCrawlScope(root string, excludes ...string) crawlScope {
	return crawlScope{
		root:     root,
		excludes: append([]string(nil), excludes...),
	}
}

// ripwireArgs returns the ripwire command-line arguments that realize the
// scope. The positional crawl root is always "." because scope.root is
// passed to the runner as the working directory (the capture convention);
// each exclude path becomes one --exclude flag, in the order given. This is
// pure string formatting so it stays unit-testable independent of the
// runner.
func (s crawlScope) ripwireArgs() []string {
	args := make([]string, 0, 1+len(s.excludes))
	args = append(args, ".")
	for _, exclude := range s.excludes {
		args = append(args, "--exclude="+exclude)
	}
	return args
}

// resolvePythonCrawlRoot resolves the crawl root for a Python project at
// projectPath, per ripwire-adapter design decision D9:
//
//   - src layout: when <projectPath>/src exists and contains at least one
//     __init__.py somewhere beneath it, the crawl root is <projectPath>/src.
//     Crawling a src-layout project from the repository root instead is the
//     pip failure mode — the import graph lives beneath src/, so the crawl
//     reports zero coupling.
//   - flat layout: otherwise the crawl root is projectPath itself.
//
// The returned root keeps the form projectPath was given. It returns an
// error when projectPath does not exist or is not a directory.
func resolvePythonCrawlRoot(projectPath string) (string, error) {
	if err := checkProjectDir(projectPath); err != nil {
		return "", fmt.Errorf("resolving python crawl root: %w", err)
	}
	src := filepath.Join(projectPath, "src")
	if info, err := os.Stat(src); err == nil && info.IsDir() && containsPythonPackage(src) {
		return src, nil
	}
	return projectPath, nil
}

// resolveTypeScriptCrawlRoot resolves the crawl root for a TypeScript
// project at projectPath, per ripwire-adapter design decision D9: the
// crawl root is the directory holding tsconfig.json or package.json.
//
// When neither file sits directly in projectPath, its subdirectories are
// searched breadth-first up to tsManifestSearchDepth levels for the nearest
// directory holding a manifest; the shallowest match wins, and within a
// level directories are visited in sorted order so the result is
// deterministic. Hidden directories and node_modules are not searched: the
// manifests inside them belong to dependencies and caches, not project
// roots. It returns an error when no manifest is found within that radius,
// or when projectPath does not exist or is not a directory.
func resolveTypeScriptCrawlRoot(projectPath string) (string, error) {
	if err := checkProjectDir(projectPath); err != nil {
		return "", fmt.Errorf("resolving typescript crawl root: %w", err)
	}
	if hasTypeScriptManifest(projectPath) {
		return projectPath, nil
	}
	if root, ok := nearestTypeScriptManifestDir(projectPath); ok {
		return root, nil
	}
	return "", fmt.Errorf("resolving typescript crawl root: no %s or %s in %q or within %d subdirectory levels of it",
		tsManifestNames[0], tsManifestNames[1], projectPath, tsManifestSearchDepth)
}

// checkProjectDir verifies that projectPath exists and is a directory, so
// it can serve as a crawl-root candidate.
func checkProjectDir(projectPath string) error {
	info, err := os.Stat(projectPath)
	if err != nil {
		// The underlying error already names the path; callers add the
		// resolver context when wrapping.
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", projectPath)
	}
	return nil
}

// containsPythonPackage reports whether dir contains at least one
// __init__.py anywhere beneath it, i.e. whether dir is or encloses a
// Python package root. The walk terminates at the first match; when none
// exists it costs a single pass over the tree. Unreadable entries are
// skipped rather than failing the check: a crawl of an unreadable tree
// fails in ripwire itself and surfaces the underlying error there.
func containsPythonPackage(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "__init__.py" {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// tsManifestNames are the files whose presence marks a directory as a
// TypeScript package root.
var tsManifestNames = []string{"package.json", "tsconfig.json"}

// tsManifestSearchDepth is how many directory levels below the project
// path the TypeScript manifest search reaches when neither manifest file
// sits directly in the project path.
const tsManifestSearchDepth = 2

// hasTypeScriptManifest reports whether dir directly holds a TypeScript
// manifest file. A directory named like a manifest does not count.
func hasTypeScriptManifest(dir string) bool {
	for _, name := range tsManifestNames {
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

// nearestTypeScriptManifestDir searches the subdirectories of root,
// breadth-first up to tsManifestSearchDepth levels, for the nearest
// directory that directly holds a TypeScript manifest. os.ReadDir returns
// entries sorted by name, so the shallowest match — ties broken by sorted
// visit order — is deterministic.
func nearestTypeScriptManifestDir(root string) (string, bool) {
	level := []string{root}
	for range tsManifestSearchDepth {
		var next []string
		for _, dir := range level {
			entries, err := os.ReadDir(dir)
			if err != nil {
				// Unreadable directories cannot hold a searched root.
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() || skipManifestSearchDir(entry.Name()) {
					continue
				}
				candidate := filepath.Join(dir, entry.Name())
				if hasTypeScriptManifest(candidate) {
					return candidate, true
				}
				next = append(next, candidate)
			}
		}
		level = next
	}
	return "", false
}

// skipManifestSearchDir reports whether a subdirectory must not be
// searched for a TypeScript package root: hidden directories (version
// control, caches) and node_modules (dependency trees whose package.json
// files do not mark a project root).
func skipManifestSearchDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules"
}
