package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// updateGoldens rewrites the committed golden files instead of comparing
// against them. It is the single writer of the goldens under testdata, so a
// regenerated golden is always the deliberate product of one run:
//
//	go test ./cmd/coverage-ratchet/ -update
var updateGoldens = flag.Bool("update", false, "rewrite golden files under testdata instead of comparing")

// goldenFixtureDir is the committed fixture directory shared by the golden and
// determinism tests.
const goldenFixtureDir = "testdata"

// goldenFixture describes one committed profile/baseline pair. The mixed
// fixture is the richest single scenario: it exercises every check-mode
// status — pass, below-floor, unregistered, and stale-floor — including the
// "-" placeholders for absent measurements and floors.
type goldenFixture struct {
	name string
}

// goldenProfilePath returns the committed coverage profile for a fixture.
func goldenProfilePath(name string) string {
	return filepath.Join(goldenFixtureDir, name+".profile.out")
}

// goldenFloorsPath returns the committed floors baseline for a fixture.
func goldenFloorsPath(name string) string {
	return filepath.Join(goldenFixtureDir, name+".floors.json")
}

// goldenTablePath returns the committed table golden for a fixture.
func goldenTablePath(name string) string {
	return filepath.Join(goldenFixtureDir, name+".golden.table.txt")
}

// goldenJSONPath returns the committed JSON golden for a fixture.
func goldenJSONPath(name string) string {
	return filepath.Join(goldenFixtureDir, name+".golden.json")
}

// runGateFixture runs the gate in check mode over a committed fixture pair and
// returns the report written to stdout.
func runGateFixture(t *testing.T, fx goldenFixture, jsonOut bool) string {
	t.Helper()
	var stdout bytes.Buffer
	result, err := RunRatchet(RatchetOptions{
		Profile: goldenProfilePath(fx.name),
		Floors:  goldenFloorsPath(fx.name),
		JSON:    jsonOut,
		Stdout:  &stdout,
	})
	if err != nil {
		t.Fatalf("RunRatchet error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("exit code: got %d, want 1 (the mixed fixture must fail the gate)", result.ExitCode)
	}
	if result.Updated {
		t.Error("Updated: got true, want false in check mode")
	}
	return stdout.String()
}

// generatedAtPattern matches the provenance timestamp value in the JSON
// report. The timestamp is the only volatile byte of the JSON output, so the
// golden pins it as a fixed placeholder.
var generatedAtPattern = regexp.MustCompile(`"generatedAt": "[^"]*"`)

// normalizeTimestamp replaces the provenance generatedAt value with a fixed
// placeholder so JSON output from any run compares equal to the golden.
func normalizeTimestamp(s string) string {
	return generatedAtPattern.ReplaceAllString(s, `"generatedAt": "TIMESTAMP"`)
}

// compareGolden compares got against the committed golden at path, rewriting
// the golden instead when -update is set.
func compareGolden(t *testing.T, path, got string) {
	t.Helper()
	if *updateGoldens {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("update golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (regenerate with: go test ./cmd/coverage-ratchet/ -update)", path, err)
	}
	if got != string(want) {
		t.Errorf("output does not match the committed golden %s:\ngot:\n%s\nwant:\n%s\nIf the change is intentional, regenerate with: go test ./cmd/coverage-ratchet/ -update",
			path, got, want)
	}
}

// TestRunRatchet_TableGolden pins the default table contract: one line per
// package (measured, floor, status) sorted by package path, "-" for absent
// measurements and floors, and a summary line. Any byte drift in the default
// output fails here; this test is the single writer of the table golden under
// -update.
func TestRunRatchet_TableGolden(t *testing.T) {
	t.Parallel()

	got := runGateFixture(t, goldenFixture{name: "mixed"}, false)
	compareGolden(t, goldenTablePath("mixed"), got)

	for _, want := range []string{"unregistered", "below-floor", "stale-floor", "pass"} {
		if !strings.Contains(got, want) {
			t.Errorf("golden table missing status %q:\n%s", want, got)
		}
	}
}

// TestRunRatchet_JSONGolden pins the --json contract: a stable object with
// per-package results (null for absent measurements and floors), the overall
// verdict, and a provenance envelope. The timestamp is excluded from the
// golden via a fixed placeholder; this test is the single writer of the JSON
// golden under -update.
func TestRunRatchet_JSONGolden(t *testing.T) {
	t.Parallel()

	got := normalizeTimestamp(runGateFixture(t, goldenFixture{name: "mixed"}, true))
	compareGolden(t, goldenJSONPath("mixed"), got)

	if !strings.Contains(got, `"verdict": "fail"`) {
		t.Errorf("golden JSON missing fail verdict:\n%s", got)
	}
	if !strings.Contains(got, `"producer": "coverage-ratchet"`) {
		t.Errorf("golden JSON missing producer:\n%s", got)
	}
}

// TestRunRatchet_Determinism asserts the spec's determinism requirement: two
// runs over the same profile and baseline produce byte-identical output —
// strictly for the table, and modulo the provenance timestamp for the JSON
// object.
func TestRunRatchet_Determinism(t *testing.T) {
	t.Parallel()

	fx := goldenFixture{name: "mixed"}

	t.Run("table_byte_identical", func(t *testing.T) {
		t.Parallel()
		first := runGateFixture(t, fx, false)
		second := runGateFixture(t, fx, false)
		if first != second {
			t.Errorf("table output not byte-identical:\nfirst:  %q\nsecond: %q", first, second)
		}
	})

	t.Run("json_identical_modulo_timestamp", func(t *testing.T) {
		t.Parallel()
		first := normalizeTimestamp(runGateFixture(t, fx, true))
		second := normalizeTimestamp(runGateFixture(t, fx, true))
		if first != second {
			t.Errorf("JSON output not identical modulo timestamp:\nfirst:  %q\nsecond: %q", first, second)
		}
	})
}
