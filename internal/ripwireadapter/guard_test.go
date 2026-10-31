package ripwireadapter

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Zero-coupling guard tests — openspec/changes/ripwire-adapter task 3.3.
//
// They pin the guard's decision table: zero dependency edges over a
// non-empty source corpus fires a single "zero-coupling-guard" warning
// naming both counts; every other combination stays silent, because an
// empty corpus legitimately yields an edgeless graph and any resolved edges
// prove the crawl worked. The counter tests pin the recursive source-file
// census that feeds the guard's sourceFileCount input.

func TestZeroCouplingGuard_DecisionTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		edgeCount       int
		sourceFileCount int
		wantWarning     bool
	}{
		{
			name:            "zero edges over source files fires",
			edgeCount:       0,
			sourceFileCount: 5,
			wantWarning:     true,
		},
		{
			name:            "zero edges over empty corpus is legitimate",
			edgeCount:       0,
			sourceFileCount: 0,
			wantWarning:     false,
		},
		{
			name:            "resolved edges over source files stays silent",
			edgeCount:       3,
			sourceFileCount: 5,
			wantWarning:     false,
		},
		{
			name:            "resolved edges without sources stays silent",
			edgeCount:       3,
			sourceFileCount: 0,
			wantWarning:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			warnings := zeroCouplingGuard(tc.edgeCount, tc.sourceFileCount)

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
			if w.Code != "zero-coupling-guard" {
				t.Errorf("code: got %q, want %q", w.Code, "zero-coupling-guard")
			}
			for _, count := range []string{
				strconv.Itoa(tc.edgeCount),
				strconv.Itoa(tc.sourceFileCount),
			} {
				if !strings.Contains(w.Message, count) {
					t.Errorf("message %q does not name count %s", w.Message, count)
				}
			}
			if w.Module != "" {
				t.Errorf("module: got %q, want empty (global warning)", w.Module)
			}
		})
	}
}

func TestCountSourceFiles_SyntheticTree(t *testing.T) {
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

	// 2 .py (one nested, to pin the recursive walk) + 1 .ts + 1 README = 3.
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "top.py"), "x = 1\n")
	writeFile(t, filepath.Join(root, "pkg", "nested.py"), "x = 1\n")
	writeFile(t, filepath.Join(root, "pkg", "mod.ts"), "export {}\n")
	writeFile(t, filepath.Join(root, "README.md"), "# corpus\n")

	got, err := countSourceFiles(root)
	if err != nil {
		t.Fatalf("countSourceFiles: %v", err)
	}
	if got != 3 {
		t.Errorf("source files: got %d, want 3", got)
	}
}

func TestCountSourceFiles_InvalidRoots(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		root string
	}{
		{"nonexistent path", filepath.Join(t.TempDir(), "does-not-exist")},
		{"file instead of directory", func() string {
			path := filepath.Join(t.TempDir(), "afile.py")
			if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
				panic(err)
			}
			return path
		}()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := countSourceFiles(tc.root)
			if err == nil {
				t.Fatalf("got count %d and no error, want error", got)
			}
			if got != 0 {
				t.Errorf("got count %d alongside error, want 0", got)
			}
		})
	}
}
