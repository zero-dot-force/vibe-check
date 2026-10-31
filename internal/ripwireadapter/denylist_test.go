package ripwireadapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Denylist pre-check tests — openspec/changes/ripwire-adapter task 3.2.
//
// They pin the mirror of ripwire's crawl denylist: source-bearing
// directories named build/dist/out/target/vendor/third_party/captures
// produce a single "denylist-pruned" warning listing them; clean trees
// produce none.

func TestDetectDenylistPruned_PythonFixture(t *testing.T) {
	t.Parallel()

	warnings, err := detectDenylistPruned(filepath.Join("testdata", "python", "src"))
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
	if !strings.Contains(w.Message, filepath.Join("fixturepy", "operations", "build")) {
		t.Errorf("message %q does not list the fixture's operations/build directory", w.Message)
	}
	if w.Module != "" {
		t.Errorf("module: got %q, want empty (global warning)", w.Module)
	}
}

func TestDetectDenylistPruned_TypeScriptFixture(t *testing.T) {
	t.Parallel()

	warnings, err := detectDenylistPruned(filepath.Join("testdata", "typescript"))
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

			warnings, err := detectDenylistPruned(dir)
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
			warnings, err := detectDenylistPruned(tc.root)
			if err == nil {
				t.Fatalf("got warnings %v and no error, want error", warnings)
			}
			if len(warnings) != 0 {
				t.Errorf("got warnings %v alongside error, want none", warnings)
			}
		})
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
