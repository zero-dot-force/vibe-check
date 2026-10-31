package main

import (
	"os"
	"path/filepath"
	"testing"
)

// detectLanguage tests — openspec/changes/ripwire-adapter task 7.1.
// Table-driven over every marker, the census majority, the tie default,
// and the markerless fallback to the existing Go path.

func writeMarker(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectLanguage_Markers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		marker string
		want   string
	}{
		{"go.mod", "go.mod", "go"},
		{"pyproject.toml", "pyproject.toml", "python"},
		{"setup.py", "setup.py", "python"},
		{"setup.cfg", "setup.cfg", "python"},
		{"tsconfig.json", "tsconfig.json", "typescript"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeMarker(t, dir, tc.marker, "x\n")
			got, err := detectLanguage(dir)
			if err != nil {
				t.Fatalf("detectLanguage: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectLanguage_MarkerPriority(t *testing.T) {
	t.Parallel()

	// A tree with both go.mod and tsconfig.json is a Go module: the
	// marker order, not map iteration, decides.
	dir := t.TempDir()
	writeMarker(t, dir, "go.mod", "module example.com/m\n")
	writeMarker(t, dir, "tsconfig.json", "{}\n")
	got, err := detectLanguage(dir)
	if err != nil {
		t.Fatalf("detectLanguage: %v", err)
	}
	if got != "go" {
		t.Errorf("got %q, want go (go.mod wins over tsconfig.json)", got)
	}
}

func TestDetectLanguage_Census(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files []string
		want  string
	}{
		{
			name:  "python majority",
			files: []string{"a.py", "b.py", "c.py", "x.ts"},
			want:  "python",
		},
		{
			name:  "typescript majority",
			files: []string{"a.ts", "b.tsx", "c.py"},
			want:  "typescript",
		},
		{
			name:  "tie defaults to go",
			files: []string{"a.py", "b.ts"},
			want:  "go",
		},
		{
			name:  "empty tree defaults to go",
			files: nil,
			want:  "go",
		},
		{
			name:  "unrelated files default to go",
			files: []string{"README.md", "Makefile"},
			want:  "go",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tc.files {
				writeMarker(t, dir, f, "x\n")
			}
			got, err := detectLanguage(dir)
			if err != nil {
				t.Fatalf("detectLanguage: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectLanguage_CensusIsRecursive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	nested := filepath.Join(dir, "src", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMarker(t, nested, "mod.py", "x\n")
	writeMarker(t, nested, "other.py", "x\n")
	writeMarker(t, dir, "only.ts", "x\n")

	got, err := detectLanguage(dir)
	if err != nil {
		t.Fatalf("detectLanguage: %v", err)
	}
	if got != "python" {
		t.Errorf("got %q, want python (nested .py beats root .ts)", got)
	}
}

func TestDetectLanguage_Errors(t *testing.T) {
	t.Parallel()

	if _, err := detectLanguage(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("got no error, want error for missing path")
	}

	file := filepath.Join(t.TempDir(), "file.txt")
	writeMarker(t, filepath.Dir(file), "file.txt", "x\n")
	if _, err := detectLanguage(file); err == nil {
		t.Error("got no error, want error for non-directory path")
	}
}
