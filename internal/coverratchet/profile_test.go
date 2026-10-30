package coverratchet

import (
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestPackageKey_ModulePrefixStripping(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
		wantErr  bool
	}{
		{
			name:     "top-level package",
			fileName: "github.com/zero-dot-force/vibe-check/metrics/values.go",
			want:     "metrics",
		},
		{
			name:     "nested internal package",
			fileName: "github.com/zero-dot-force/vibe-check/internal/scaffold/scaffold.go",
			want:     "internal/scaffold",
		},
		{
			name:     "command package",
			fileName: "github.com/zero-dot-force/vibe-check/cmd/vibe-check/main.go",
			want:     "cmd/vibe-check",
		},
		{
			name:     "foreign module path",
			fileName: "example.com/other/pkg/tool.go",
			wantErr:  true,
		},
		{
			name:     "module path without trailing slash",
			fileName: "github.com/zero-dot-force/vibe-check",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PackageKey(tt.fileName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("PackageKey(%q) = %q, nil; want error", tt.fileName, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("PackageKey(%q) error: %v", tt.fileName, err)
			}
			if got != tt.want {
				t.Errorf("PackageKey(%q) = %q, want %q", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestLoadProfiles_MultiPackage(t *testing.T) {
	profiles, err := LoadProfiles(filepath.Join("testdata", "multipackage.out"))
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("len(profiles) = %d, want 3 (one per source file)", len(profiles))
	}
	var got []string
	for _, p := range profiles {
		key, err := PackageKey(p.FileName)
		if err != nil {
			t.Fatalf("PackageKey(%q) error: %v", p.FileName, err)
		}
		got = append(got, key)
	}
	sort.Strings(got)
	want := []string{"internal/scaffold", "metrics", "metrics"}
	if !slices.Equal(got, want) {
		t.Errorf("package keys = %v, want %v", got, want)
	}
}

func TestLoadProfiles_UnprefixablePath(t *testing.T) {
	profiles, err := LoadProfiles(filepath.Join("testdata", "foreign.out"))
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("len(profiles) = %d, want 1", len(profiles))
	}
	if _, err := PackageKey(profiles[0].FileName); err == nil {
		t.Errorf("PackageKey(%q) = nil error, want error for file outside the module", profiles[0].FileName)
	}
}

func TestLoadProfiles_MissingFile(t *testing.T) {
	if _, err := LoadProfiles(filepath.Join("testdata", "does-not-exist.out")); err == nil {
		t.Errorf("LoadProfiles on missing file = nil error, want error")
	}
}
