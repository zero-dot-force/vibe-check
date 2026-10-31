package ripwireadapter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Crawl-scope tests — openspec/changes/ripwire-adapter task 3.1.
//
// They pin the root-resolution rules of design decision D9 (Python
// src-layout vs flat, TypeScript manifest directories) and the argument
// shape the runner receives for a resolved scope: positional "." at the
// crawl root, one --exclude flag per pruned subtree.

// writeFile creates name (with parent directories) under dir, failing the
// test on error.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("writeFile: mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: write %s: %v", path, err)
	}
}

func TestResolvePythonCrawlRoot_SrcLayoutFixture(t *testing.T) {
	t.Parallel()

	project := filepath.Join("testdata", "python")
	want := filepath.Join("testdata", "python", "src")

	got, err := resolvePythonCrawlRoot(project)
	if err != nil {
		t.Fatalf("resolvePythonCrawlRoot(%q): %v", project, err)
	}
	if got != want {
		t.Errorf("resolvePythonCrawlRoot(%q) = %q, want %q", project, got, want)
	}
}

// TestResolvePythonCrawlRoot_SrcLayoutAvoidsPipFailureMode guards the
// failure mode that motivated src-layout detection: crawling a src-layout
// project from its repository root silently yields zero coupling because
// the import graph lives beneath src/. The full coupling assertion against
// the recorded deps.xml (health files="7" with includes) lands with the
// functional pipeline tests (task 8.1); here the check is only that
// resolution steers the crawl into the package tree.
func TestResolvePythonCrawlRoot_SrcLayoutAvoidsPipFailureMode(t *testing.T) {
	t.Parallel()

	project := filepath.Join("testdata", "python")

	root, err := resolvePythonCrawlRoot(project)
	if err != nil {
		t.Fatalf("resolvePythonCrawlRoot(%q): %v", project, err)
	}
	if root == project {
		t.Fatalf("src-layout fixture resolved to the project root %q; crawling from there yields zero coupling (the pip failure mode)", project)
	}
	if !containsPythonPackage(root) {
		t.Errorf("resolved root %q contains no Python package, so crawling it could not yield coupling", root)
	}
}

func TestResolvePythonCrawlRoot_FlatLayout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, f := range []struct{ name, content string }{
		{"main.py", "import helper\n"},
		{"helper.py", "ANSWER = 42\n"},
		{"pyproject.toml", "[project]\nname = \"flat-demo\"\n"},
	} {
		writeFile(t, dir, f.name, f.content)
	}

	got, err := resolvePythonCrawlRoot(dir)
	if err != nil {
		t.Fatalf("resolvePythonCrawlRoot(%q): %v", dir, err)
	}
	if got != dir {
		t.Errorf("resolvePythonCrawlRoot(%q) = %q, want %q (a flat project resolves to itself)", dir, got, dir)
	}
}

// TestResolvePythonCrawlRoot_SrcWithoutPackagesIsFlat covers the edge
// where src/ exists but holds no __init__.py anywhere beneath it: it is
// not a package root, so the project is treated as flat.
func TestResolvePythonCrawlRoot_SrcWithoutPackagesIsFlat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("src", "script.py"), "print('hi')\n")
	writeFile(t, dir, filepath.Join("src", "lib", "helper.py"), "ANSWER = 42\n")

	got, err := resolvePythonCrawlRoot(dir)
	if err != nil {
		t.Fatalf("resolvePythonCrawlRoot(%q): %v", dir, err)
	}
	if got != dir {
		t.Errorf("resolvePythonCrawlRoot(%q) = %q, want %q (src without packages is flat)", dir, got, dir)
	}
}

func TestResolveTypeScriptCrawlRoot_Fixture(t *testing.T) {
	t.Parallel()

	project := filepath.Join("testdata", "typescript")

	got, err := resolveTypeScriptCrawlRoot(project)
	if err != nil {
		t.Fatalf("resolveTypeScriptCrawlRoot(%q): %v", project, err)
	}
	if got != project {
		t.Errorf("resolveTypeScriptCrawlRoot(%q) = %q, want %q", project, got, project)
	}
}

func TestResolveTypeScriptCrawlRoot_ManifestAtRoot(t *testing.T) {
	t.Parallel()

	for _, manifest := range tsManifestNames {
		t.Run(manifest, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeFile(t, dir, manifest, "{}\n")

			got, err := resolveTypeScriptCrawlRoot(dir)
			if err != nil {
				t.Fatalf("resolveTypeScriptCrawlRoot(%q): %v", dir, err)
			}
			if got != dir {
				t.Errorf("resolveTypeScriptCrawlRoot(%q) = %q, want %q", dir, got, dir)
			}
		})
	}
}

// TestResolveTypeScriptCrawlRoot_NearestManifestWins covers the
// subdirectory search: with no manifest at the project path itself, the
// shallowest manifest directory becomes the crawl root.
func TestResolveTypeScriptCrawlRoot_NearestManifestWins(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("packages", "web", "package.json"), "{\"name\":\"web\"}\n")
	writeFile(t, dir, filepath.Join("apps", "cli", "index.ts"), "console.log('hi');\n")
	writeFile(t, dir, filepath.Join("apps", "cli", "tsconfig.json"), "{}\n")

	got, err := resolveTypeScriptCrawlRoot(dir)
	if err != nil {
		t.Fatalf("resolveTypeScriptCrawlRoot(%q): %v", dir, err)
	}
	want := filepath.Join(dir, "apps", "cli")
	if got != want {
		t.Errorf("resolveTypeScriptCrawlRoot(%q) = %q, want %q (shallowest manifest directory)", dir, got, want)
	}
}

// TestResolveTypeScriptCrawlRoot_SortedTieBreak covers the determinism
// rule: with manifests at equal depth, the sorted visit order decides,
// here "alpha" over "beta".
func TestResolveTypeScriptCrawlRoot_SortedTieBreak(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("beta", "package.json"), "{\"name\":\"beta\"}\n")
	writeFile(t, dir, filepath.Join("alpha", "package.json"), "{\"name\":\"alpha\"}\n")

	got, err := resolveTypeScriptCrawlRoot(dir)
	if err != nil {
		t.Fatalf("resolveTypeScriptCrawlRoot(%q): %v", dir, err)
	}
	want := filepath.Join(dir, "alpha")
	if got != want {
		t.Errorf("resolveTypeScriptCrawlRoot(%q) = %q, want %q (sorted visit order breaks ties)", dir, got, want)
	}
}

// TestResolveTypeScriptCrawlRoot_IgnoresDependencyAndHiddenDirs covers the
// search exclusions: manifests inside node_modules or hidden directories
// belong to dependencies and caches, not project roots, so a tree whose
// only manifests live there fails resolution.
func TestResolveTypeScriptCrawlRoot_IgnoresDependencyAndHiddenDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("node_modules", "left-pad", "package.json"), "{\"name\":\"left-pad\"}\n")
	writeFile(t, dir, filepath.Join(".config", "package.json"), "{\"name\":\"cache\"}\n")

	if got, err := resolveTypeScriptCrawlRoot(dir); err == nil {
		t.Errorf("resolveTypeScriptCrawlRoot(%q) = %q, want error (dependency and hidden-dir manifests are not roots)", dir, got)
	}
}

func TestResolveCrawlRoot_InvalidPaths(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "does-not-exist")
	file := filepath.Join(t.TempDir(), "a-file.txt")
	if err := os.WriteFile(file, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}

	cases := []struct {
		name    string
		resolve func(string) (string, error)
		path    string
	}{
		{"python missing path", resolvePythonCrawlRoot, missing},
		{"typescript missing path", resolveTypeScriptCrawlRoot, missing},
		{"python file path", resolvePythonCrawlRoot, file},
		{"typescript file path", resolveTypeScriptCrawlRoot, file},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root, err := tc.resolve(tc.path)
			if err == nil {
				t.Fatalf("%s: resolved %q, want error", tc.name, root)
			}
		})
	}
}

// TestCrawlScope_RipwireArgs pins the argument shape the runner will
// receive: the positional crawl root is "." with the scope's root as the
// runner's working directory (the capture convention), and each exclude is
// a separate --exclude flag in the order given.
func TestCrawlScope_RipwireArgs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		scope crawlScope
		want  []string
	}{
		{
			name:  "root only",
			scope: newCrawlScope("testdata/python/src"),
			want:  []string{"."},
		},
		{
			name:  "excludes in given order",
			scope: newCrawlScope("testdata/python/src", "a", "b"),
			want:  []string{".", "--exclude=a", "--exclude=b"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.scope.ripwireArgs()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ripwireArgs() = %v, want %v", got, tc.want)
			}
		})
	}
}
