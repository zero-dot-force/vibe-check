package coverratchet

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBaseline_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage-floors.json")
	want := Baseline{
		Version: BaselineVersion,
		Floors: map[string]float64{
			"cmd/vibe-check":     92.7,
			"internal/goadapter": 89.3,
			"internal/scaffold":  87.9,
			"metrics":            82.9,
			"metrics/report":     99.4,
		},
	}
	if err := WriteBaseline(path, want); err != nil {
		t.Fatalf("WriteBaseline error: %v", err)
	}
	got, err := ReadBaseline(path)
	if err != nil {
		t.Fatalf("ReadBaseline error: %v", err)
	}
	if got.Version != want.Version {
		t.Errorf("Version = %d, want %d", got.Version, want.Version)
	}
	if !reflect.DeepEqual(got.Floors, want.Floors) {
		t.Errorf("Floors = %v, want %v", got.Floors, want.Floors)
	}
}

func TestWriteBaseline_DeterministicSerialization(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage-floors.json")
	b := Baseline{
		Version: BaselineVersion,
		Floors: map[string]float64{
			"metrics":            82.9,
			"cmd/vibe-check":     92.7,
			"internal/goadapter": 89.3,
		},
	}
	if err := WriteBaseline(path, b); err != nil {
		t.Fatalf("WriteBaseline error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	want := `{
  "version": 1,
  "floors": {
    "cmd/vibe-check": 92.7,
    "internal/goadapter": 89.3,
    "metrics": 82.9
  }
}
`
	if string(data) != want {
		t.Errorf("serialized baseline = %q, want %q", data, want)
	}
	if data[len(data)-1] != '\n' {
		t.Errorf("serialized baseline lacks trailing newline")
	}
}

func TestWriteBaseline_ByteIdenticalRewrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage-floors.json")
	b := Baseline{
		Version: BaselineVersion,
		Floors: map[string]float64{
			"metrics":        82.9,
			"metrics/report": 99.4,
		},
	}
	if err := WriteBaseline(path, b); err != nil {
		t.Fatalf("WriteBaseline error: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	reread, err := ReadBaseline(path)
	if err != nil {
		t.Fatalf("ReadBaseline error: %v", err)
	}
	if err := WriteBaseline(path, reread); err != nil {
		t.Fatalf("WriteBaseline rewrite error: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("rewrite produced a diff: first %q, second %q", first, second)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("staged temp file remains after write: %v", err)
	}
}

func TestWriteBaseline_RenameFailureCleansStagedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage-floors.json")
	// A directory at the target path makes the atomic rename fail, which
	// exercises the staged-file cleanup on the error path.
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("Mkdir error: %v", err)
	}
	b := Baseline{Version: BaselineVersion, Floors: map[string]float64{"metrics": 82.9}}
	if err := WriteBaseline(path, b); err == nil {
		t.Fatal("WriteBaseline onto a directory = nil error, want error")
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("staged temp file remains after failed rename: %v", err)
	}
}

func TestReadBaseline_Rejections(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "unknown version",
			content: `{"version": 2, "floors": {"metrics": 82.9}}`,
		},
		{
			name:    "missing version",
			content: `{"floors": {"metrics": 82.9}}`,
		},
		{
			name:    "malformed JSON",
			content: `{"version": 1, "floors": {`,
		},
		{
			name:    "floor at zero disables gate",
			content: `{"version": 1, "floors": {"metrics": 0}}`,
		},
		{
			name:    "negative floor",
			content: `{"version": 1, "floors": {"metrics": -1}}`,
		},
		{
			name:    "floor above 100",
			content: `{"version": 1, "floors": {"metrics": 150}}`,
		},
		{
			name:    "two-decimal floor",
			content: `{"version": 1, "floors": {"metrics": 92.75}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "floors.json")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("WriteFile error: %v", err)
			}
			if _, err := ReadBaseline(path); err == nil {
				t.Errorf("ReadBaseline(%q) = nil error, want error", tt.content)
			}
		})
	}
}

func TestReadBaseline_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	if _, err := ReadBaseline(path); err == nil {
		t.Errorf("ReadBaseline on missing file = nil error, want error")
	}
}

func TestReadBaseline_ValidFloorDomain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "floors.json")
	content := `{"version": 1, "floors": {"metrics": 0.1, "metrics/report": 100}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	if _, err := ReadBaseline(path); err != nil {
		t.Errorf("ReadBaseline on domain-valid floors error: %v", err)
	}
}

func TestWriteBaseline_RejectsInvalidFloors(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name   string
		floors map[string]float64
	}{
		{name: "floor below range", floors: map[string]float64{"metrics": 0}},
		{name: "floor above range", floors: map[string]float64{"metrics": 100.1}},
		{name: "two-decimal floor", floors: map[string]float64{"metrics": 82.94}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "floors.json")
			b := Baseline{Version: BaselineVersion, Floors: tt.floors}
			if err := WriteBaseline(path, b); err == nil {
				t.Errorf("WriteBaseline(%v) = nil error, want error", tt.floors)
			}
		})
	}
}
