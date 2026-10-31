package ripwireadapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TypeScript abstractness tests — openspec/changes/ripwire-adapter task
// 6.1. They pin the interface-share counts against the hand-computed
// golden (including the namespace-nested interface) and the bias
// warning.

func TestTSAbstractness_InterfaceShareMatchesGolden(t *testing.T) {
	t.Parallel()

	// The recorded arch table's ./src row is the golden source: 4 abstract
	// (Point, Named, Repository, and namespace-nested Shapes.Point) of 5
	// exported types (the four interfaces plus the SplitMethods class).
	archData := readRecording(t, filepath.Join("testdata", "recorded", "typescript", "arch.txt"))
	table, err := parseArchOutput(archData)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	var srcRow *archModuleRow
	for i := range table.Rows {
		if table.Rows[i].Path == "./src" {
			srcRow = &table.Rows[i]
		}
	}
	if srcRow == nil {
		t.Fatal("./src row not found in recorded arch table")
	}
	if srcRow.Abstract != 4 {
		t.Errorf("abstract: got %d, want 4 (incl. nested)", srcRow.Abstract)
	}
	if srcRow.Types != 5 {
		t.Errorf("types: got %d, want 5", srcRow.Types)
	}

	// Cross-check the hand computation against the fixture source: the
	// four interface declarations (one nested inside namespace Shapes)
	// and one class are exactly what the share counts.
	models, err := os.ReadFile(filepath.Join("testdata", "typescript", "src", "models.ts"))
	if err != nil {
		t.Fatalf("read models.ts: %v", err)
	}
	if got := strings.Count(string(models), "export interface "); got != 4 {
		t.Errorf("models.ts interface declarations: got %d, want 4", got)
	}
	if got := strings.Count(string(models), "class "); got != 1 {
		t.Errorf("models.ts class declarations: got %d, want 1", got)
	}

	// The mapping carries the share into the module (pinned further in
	// mapping_test.go); the bias warning is present for TypeScript.
	w := tsAbstractnessBiasWarning()
	if w.Code != "ts-abstractness-bias" {
		t.Errorf("warning code: got %q, want ts-abstractness-bias", w.Code)
	}
	for _, want := range []string{"abstract class", "alias"} {
		if !strings.Contains(w.Message, want) {
			t.Errorf("warning message %q does not mention %q", w.Message, want)
		}
	}
}
