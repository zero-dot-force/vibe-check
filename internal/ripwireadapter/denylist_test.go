package ripwireadapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Denylist pre-check tests — openspec/changes/archive/2026-11-24-ripwire-adapter task 3.2,
// extended by openspec/changes/archive/2026-11-25-bump-ripwire-pin task 4.1.
//
// They pin the mirror of ripwire's crawl denylist: source-bearing
// directories named build/dist/out/target/vendor/third_party/captures
// produce a single "denylist-pruned" warning listing them; clean trees
// produce none. The warning lists the union of the pre-check and ripwire's
// own <pruned_packages> disclosure (design D5).

func TestDetectDenylistPruned_PythonFixture(t *testing.T) {
	t.Parallel()

	warnings, err := detectDenylistPruned(filepath.Join("testdata", "python", "src"), nil, nil, 0)
	if err != nil {
		t.Fatalf("detectDenylistPruned: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d (%v), want exactly 1", len(warnings), warnings)
	}
	w := warnings[0]
	if w.Code != "denylist-pruned" {
		t.Errorf("code: got %q, want %q", w.Code, "denylist-pruned")
	}
	if !strings.Contains(w.Message, "fixturepy/operations/build") {
		t.Errorf("message %q does not list the fixture's operations/build directory", w.Message)
	}
	if w.Module != "" {
		t.Errorf("module: got %q, want empty (global warning)", w.Module)
	}
}

func TestDetectDenylistPruned_TypeScriptFixture(t *testing.T) {
	t.Parallel()

	warnings, err := detectDenylistPruned(filepath.Join("testdata", "typescript"), nil, nil, 0)
	if err != nil {
		t.Fatalf("detectDenylistPruned: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: got %d (%v), want none", len(warnings), warnings)
	}
}

func TestDetectDenylistPruned_SyntheticTrees(t *testing.T) {
	t.Parallel()

	writeFile := func(t *testing.T, path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	cases := []struct {
		name        string
		build       func(t *testing.T, dir string)
		wantWarning bool
		wantListed  []string
	}{
		{
			name: "non-source build dir is invisible",
			build: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "build", "README"), "docs only")
			},
			wantWarning: false,
		},
		{
			name: "source-bearing vendor and captures both listed sorted",
			build: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "captures", "x.ts"), "export {}\n")
				writeFile(t, filepath.Join(dir, "vendor", "lib.py"), "x = 1\n")
			},
			wantWarning: true,
			wantListed:  []string{"captures", "vendor"},
		},
		{
			name: "nested source below denylist dir counts as source-bearing",
			build: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "out", "nested", "deep", "m.py"), "x = 1\n")
			},
			wantWarning: true,
			wantListed:  []string{"out"},
		},
		{
			name: "denylist name with wrong case is not pruned",
			build: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "Build", "a.py"), "x = 1\n")
			},
			wantWarning: false,
		},
		{
			name: "clean tree emits nothing",
			build: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "pkg", "a.py"), "x = 1\n")
			},
			wantWarning: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.build(t, dir)

			warnings, err := detectDenylistPruned(dir, nil, nil, 0)
			if err != nil {
				t.Fatalf("detectDenylistPruned: %v", err)
			}
			if !tc.wantWarning {
				if len(warnings) != 0 {
					t.Fatalf("warnings: got %d (%v), want none", len(warnings), warnings)
				}
				return
			}
			if len(warnings) != 1 {
				t.Fatalf("warnings: got %d (%v), want exactly 1", len(warnings), warnings)
			}
			w := warnings[0]
			if w.Code != "denylist-pruned" {
				t.Errorf("code: got %q, want denylist-pruned", w.Code)
			}
			for _, listed := range tc.wantListed {
				if !strings.Contains(w.Message, listed) {
					t.Errorf("message %q does not list %q", w.Message, listed)
				}
			}
		})
	}
}

func TestDetectDenylistPruned_InvalidRoots(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		root string
	}{
		{"nonexistent path", filepath.Join(t.TempDir(), "does-not-exist")},
		{"file instead of directory", func() string {
			path := filepath.Join(t.TempDir(), "afile")
			if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
				panic(err)
			}
			return path
		}()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			warnings, err := detectDenylistPruned(tc.root, nil, nil, 0)
			if err == nil {
				t.Fatalf("got warnings %v and no error, want error", warnings)
			}
			if len(warnings) != 0 {
				t.Errorf("got warnings %v alongside error, want none", warnings)
			}
		})
	}
}

// TestDetectDenylistPruned_UnionWithRipwire pins design D5: the warning
// lists the sorted, de-duplicated union of both sources and discloses a
// partial ripwire listing.
func TestDetectDenylistPruned_UnionWithRipwire(t *testing.T) {
	t.Parallel()

	pythonRoot := filepath.Join("testdata", "python", "src")
	cases := []struct {
		name          string
		root          string
		ripwirePruned []string
		ripwireCount  int
		want          string // full message; "" means no warning
	}{
		{
			name:          "both sources agree: listed once",
			root:          pythonRoot,
			ripwirePruned: []string{"fixturepy/operations/build"},
			ripwireCount:  1,
			want:          "ripwire prunes source-bearing denylist directories from the crawl: fixturepy/operations/build",
		},
		{
			name:          "ripwire-only package is added and sorted",
			root:          pythonRoot,
			ripwirePruned: []string{"fixturepy/zz/staging"},
			ripwireCount:  1,
			want:          "ripwire prunes source-bearing denylist directories from the crawl: fixturepy/operations/build, fixturepy/zz/staging",
		},
		{
			name:          "partial ripwire listing states the total",
			root:          pythonRoot,
			ripwirePruned: []string{"fixturepy/operations/build"},
			ripwireCount:  9,
			want: "ripwire prunes source-bearing denylist directories from the crawl: fixturepy/operations/build" +
				"; ripwire reports 9 pruned packages in total and lists only 1 (run ripwire --skipped for all)",
		},
		{
			name:          "ripwire-only on a clean tree",
			root:          filepath.Join("testdata", "typescript"),
			ripwirePruned: []string{"pkg/dist"},
			ripwireCount:  1,
			want:          "ripwire prunes source-bearing denylist directories from the crawl: pkg/dist",
		},
		{
			name: "neither source",
			root: filepath.Join("testdata", "typescript"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			warnings, err := detectDenylistPruned(tc.root, nil, tc.ripwirePruned, tc.ripwireCount)
			if err != nil {
				t.Fatalf("detectDenylistPruned: %v", err)
			}
			if tc.want == "" {
				if len(warnings) != 0 {
					t.Fatalf("warnings: got %v, want none", warnings)
				}
				return
			}
			if len(warnings) != 1 || warnings[0].Code != warningDenylistPruned || warnings[0].Message != tc.want {
				t.Errorf("warnings: got %+v, want one %s warning with message %q", warnings, warningDenylistPruned, tc.want)
			}
		})
	}
}

// TestDetectDenylistPruned_PlainDirectoryFromPreCheck pins the
// pre-check's role for directories ripwire does not row on --arch: a
// package-less dist/ (ripwire's why="pruned-dir" class) is listed even
// when ripwire's <pruned_packages> is empty.
func TestDetectDenylistPruned_PlainDirectoryFromPreCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pkg", "dist"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "dist", "bundle.py"), []byte("x = 1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	warnings, err := detectDenylistPruned(dir, nil, nil, 0)
	if err != nil {
		t.Fatalf("detectDenylistPruned: %v", err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0].Message, "pkg/dist") {
		t.Errorf("warnings: got %+v, want one listing pkg/dist", warnings)
	}
}

// TestDenylistNamesMatchRipwire pins the mirrored name set so it cannot
// drift from the validated surface without a test failing.
func TestDenylistNamesMatchRipwire(t *testing.T) {
	t.Parallel()

	want := []string{"build", "dist", "out", "target", "vendor", "third_party", "captures"}
	if len(crawlDenylistNames) != len(want) {
		t.Fatalf("denylist names: got %v, want %v", crawlDenylistNames, want)
	}
	for i, name := range want {
		if crawlDenylistNames[i] != name {
			t.Errorf("denylist names[%d]: got %q, want %q", i, crawlDenylistNames[i], name)
		}
	}
}

// TestDetectDenylistPruned_IndexedCandidateNotReported pins the pinned
// release's tracked-build/ behavior: ripwire indexes a build/ directory
// whose files git tracks, so a denylist-named directory with an --arch
// module at or beneath it is not reported, while a nested candidate ripwire
// did not index still is.
func TestDetectDenylistPruned_IndexedCandidateNotReported(t *testing.T) {
	t.Parallel()

	pythonRoot := filepath.Join("testdata", "python", "src")
	dir := t.TempDir()
	for _, f := range []string{"pkg/build/sub/a.py", "pkg/build/dist/b.py"} {
		path := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("x = 1\n"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	cases := []struct {
		name    string
		root    string
		indexed []string
		want    string // full message; "" means no warning
	}{
		{
			name:    "indexed row equals the candidate",
			root:    pythonRoot,
			indexed: []string{"fixturepy", "fixturepy/operations", "fixturepy/operations/build"},
		},
		{
			name:    "dot-slash row form",
			root:    pythonRoot,
			indexed: []string{"./fixturepy/operations/build"},
		},
		{
			name:    "sibling prefix is not beneath",
			root:    pythonRoot,
			indexed: []string{"fixturepy/operations/buildx"},
			want:    "ripwire prunes source-bearing denylist directories from the crawl: fixturepy/operations/build",
		},
		{
			name:    "indexed beneath the candidate; nested candidate still reported",
			root:    dir,
			indexed: []string{"pkg/build/sub"},
			want:    "ripwire prunes source-bearing denylist directories from the crawl: pkg/build/dist",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			warnings, err := detectDenylistPruned(tc.root, tc.indexed, nil, 0)
			if err != nil {
				t.Fatalf("detectDenylistPruned: %v", err)
			}
			if tc.want == "" {
				if len(warnings) != 0 {
					t.Fatalf("warnings: got %v, want none", warnings)
				}
				return
			}
			if len(warnings) != 1 || warnings[0].Message != tc.want {
				t.Errorf("warnings: got %+v, want one with message %q", warnings, tc.want)
			}
		})
	}
}
