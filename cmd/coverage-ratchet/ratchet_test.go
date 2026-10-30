package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testProfile describes three packages with deterministic statement counts:
// internal/coverratchet 2/5 = 40.0%, cmd/coverage-ratchet 2/2 = 100.0%, and
// metrics 3/5 = 60.0%.
const testProfile = `mode: set
github.com/zero-dot-force/vibe-check/internal/coverratchet/profile.go:10.20,12.3 2 1
github.com/zero-dot-force/vibe-check/internal/coverratchet/profile.go:15.10,17.3 3 0
github.com/zero-dot-force/vibe-check/cmd/coverage-ratchet/ratchet.go:5.2,7.1 2 1
github.com/zero-dot-force/vibe-check/metrics/values.go:8.2,10.1 3 1
github.com/zero-dot-force/vibe-check/metrics/values.go:12.2,14.1 2 0
`

// testProfileWithNewPkg adds cmd/newpkg at 2/4 = 50.0%, below the 80.0 entry
// bar, so an update run must refuse to register it.
const testProfileWithNewPkg = testProfile + `github.com/zero-dot-force/vibe-check/cmd/newpkg/new.go:1.2,3.1 2 1
github.com/zero-dot-force/vibe-check/cmd/newpkg/new.go:5.2,7.1 2 0
`

const testFloorsPass = `{
  "version": 1,
  "floors": {
    "cmd/coverage-ratchet": 100.0,
    "internal/coverratchet": 40.0,
    "metrics": 60.0
  }
}
`

const testFloorsBelowFloor = `{
  "version": 1,
  "floors": {
    "cmd/coverage-ratchet": 100.0,
    "internal/coverratchet": 80.0,
    "metrics": 60.0
  }
}
`

const testFloorsUnregistered = `{
  "version": 1,
  "floors": {
    "cmd/coverage-ratchet": 100.0,
    "internal/coverratchet": 40.0
  }
}
`

const testFloorsStale = `{
  "version": 1,
  "floors": {
    "cmd/coverage-ratchet": 100.0,
    "internal/coverratchet": 40.0,
    "internal/gone": 50.0,
    "metrics": 60.0
  }
}
`

const testFloorsRaiseTarget = `{
  "version": 1,
  "floors": {
    "cmd/coverage-ratchet": 100.0,
    "internal/coverratchet": 20.0,
    "metrics": 60.0
  }
}
`

// writeFixture writes content to dir/name and returns the path.
func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// mustReadFile reads path and returns its content, failing the test on error.
func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestRatchetFlagParsing verifies that all four flags are registered with the
// documented defaults and accept explicit values.
func TestRatchetFlagParsing(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Defaults on a fresh command.
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		t.Fatalf("GetString profile: %v", err)
	}
	if profile != "coverage.out" {
		t.Errorf("default profile: got %q, want %q", profile, "coverage.out")
	}
	floors, err := cmd.Flags().GetString("floors")
	if err != nil {
		t.Fatalf("GetString floors: %v", err)
	}
	if floors != "coverage-floors.json" {
		t.Errorf("default floors: got %q, want %q", floors, "coverage-floors.json")
	}
	update, err := cmd.Flags().GetBool("update")
	if err != nil {
		t.Fatalf("GetBool update: %v", err)
	}
	if update {
		t.Error("default update: got true, want false")
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		t.Fatalf("GetBool json: %v", err)
	}
	if jsonOut {
		t.Error("default json: got true, want false")
	}

	// Explicit values parse without error.
	if err := cmd.Flags().Parse([]string{
		"--profile", "p.out",
		"--floors", "f.json",
		"--update",
		"--json",
	}); err != nil {
		t.Fatalf("flag parsing failed: %v", err)
	}
	profile, err = cmd.Flags().GetString("profile")
	if err != nil {
		t.Fatalf("GetString profile: %v", err)
	}
	if profile != "p.out" {
		t.Errorf("profile: got %q, want %q", profile, "p.out")
	}
	floors, err = cmd.Flags().GetString("floors")
	if err != nil {
		t.Fatalf("GetString floors: %v", err)
	}
	if floors != "f.json" {
		t.Errorf("floors: got %q, want %q", floors, "f.json")
	}
	update, err = cmd.Flags().GetBool("update")
	if err != nil {
		t.Fatalf("GetBool update: %v", err)
	}
	if !update {
		t.Error("update: got false, want true")
	}
	jsonOut, err = cmd.Flags().GetBool("json")
	if err != nil {
		t.Fatalf("GetBool json: %v", err)
	}
	if !jsonOut {
		t.Error("json: got false, want true")
	}
}

// TestVersionString_DevDefault verifies that without ldflags injection the
// version is the dev default: the version global holds "dev", semanticVersion
// reports "dev", and versionString starts with it. The test never mutates the
// version global because parallel tests read it through rootCmd and the
// provenance envelope; it asserts the ambient test-build default instead.
func TestVersionString_DevDefault(t *testing.T) {
	t.Parallel()

	if version != "dev" {
		t.Fatalf("version: got %q, want %q (test builds inject no ldflags)", version, "dev")
	}
	if got := semanticVersion(); got != "dev" {
		t.Errorf("semanticVersion: got %q, want %q", got, "dev")
	}
	if got := versionString(); !strings.HasPrefix(got, "dev") {
		t.Errorf("versionString: got %q, want prefix %q", got, "dev")
	}
}

// TestRunRatchet_ExitCodes exercises every exit-code path of the gate: 0 on
// success, 1 on policy failures (below-floor, unregistered, stale floor), and
// 2 on tool failures (missing profile, malformed baseline). Check-mode runs
// never modify the baseline file.
func TestRunRatchet_ExitCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		profile      string
		floors       string
		wantExitCode int
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "all_pass",
			profile:      testProfile,
			floors:       testFloorsPass,
			wantExitCode: 0,
			wantContains: []string{"summary: 3 packages, 3 pass"},
		},
		{
			name:         "below_floor",
			profile:      testProfile,
			floors:       testFloorsBelowFloor,
			wantExitCode: 1,
			wantContains: []string{"below-floor", "summary: 3 packages, 2 pass, 1 below-floor"},
		},
		{
			name:         "unregistered",
			profile:      testProfile,
			floors:       testFloorsUnregistered,
			wantExitCode: 1,
			wantContains: []string{"unregistered", "summary: 3 packages, 2 pass, 1 unregistered"},
		},
		{
			name:         "stale_floor",
			profile:      testProfile,
			floors:       testFloorsStale,
			wantExitCode: 1,
			wantContains: []string{"stale-floor", "internal/gone", "summary: 4 packages, 3 pass, 1 stale-floor"},
		},
		{
			name:         "missing_profile",
			profile:      "",
			floors:       testFloorsPass,
			wantExitCode: 2,
			wantErr:      true,
		},
		{
			name:         "malformed_baseline",
			profile:      testProfile,
			floors:       "not json",
			wantExitCode: 2,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			profilePath := writeFixture(t, dir, "profile.out", tt.profile)
			if tt.profile == "" {
				profilePath = filepath.Join(dir, "missing.out")
			}
			floorsPath := writeFixture(t, dir, "coverage-floors.json", tt.floors)

			var stdout bytes.Buffer
			opts := RatchetOptions{
				Profile: profilePath,
				Floors:  floorsPath,
				Stdout:  &stdout,
			}
			result, err := RunRatchet(opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RunRatchet error: got %v, wantErr %v", err, tt.wantErr)
			}
			if result.ExitCode != tt.wantExitCode {
				t.Errorf("exit code: got %d, want %d", result.ExitCode, tt.wantExitCode)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(stdout.String(), want) {
					t.Errorf("output missing %q; got:\n%s", want, stdout.String())
				}
			}
			// Check mode never writes the baseline.
			if got := mustReadFile(t, floorsPath); got != tt.floors {
				t.Errorf("baseline modified in check mode: got %q, want %q", got, tt.floors)
			}
		})
	}
}

// TestRunRatchet_Update exercises the tighten-only update mode: floors rise to
// measured coverage, a below-bar new package is refused without any write,
// and a no-change update rewrites the file byte-identically.
func TestRunRatchet_Update(t *testing.T) {
	t.Parallel()

	t.Run("raises_floor", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		profilePath := writeFixture(t, dir, "profile.out", testProfile)
		floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsRaiseTarget)

		var stdout bytes.Buffer
		result, err := RunRatchet(RatchetOptions{
			Profile: profilePath,
			Floors:  floorsPath,
			Update:  true,
			Stdout:  &stdout,
		})
		if err != nil {
			t.Fatalf("RunRatchet error: %v", err)
		}
		if result.ExitCode != 0 {
			t.Errorf("exit code: got %d, want 0", result.ExitCode)
		}
		if !result.Updated {
			t.Error("Updated: got false, want true")
		}
		var baseline struct {
			Version int                `json:"version"`
			Floors  map[string]float64 `json:"floors"`
		}
		if err := json.Unmarshal([]byte(mustReadFile(t, floorsPath)), &baseline); err != nil {
			t.Fatalf("unmarshal updated baseline: %v", err)
		}
		want := map[string]float64{
			"cmd/coverage-ratchet":  100.0,
			"internal/coverratchet": 40.0,
			"metrics":               60.0,
		}
		for pkg, floor := range want {
			if got := baseline.Floors[pkg]; got != floor {
				t.Errorf("floor for %s: got %v, want %v", pkg, got, floor)
			}
		}
		out := stdout.String()
		if !strings.Contains(out, "raised") {
			t.Errorf("output missing %q; got:\n%s", "raised", out)
		}
		if !strings.Contains(out, "baseline updated: "+floorsPath) {
			t.Errorf("output missing %q; got:\n%s", "baseline updated: "+floorsPath, out)
		}
	})

	t.Run("refused_below_bar_leaves_file_untouched", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		profilePath := writeFixture(t, dir, "profile.out", testProfileWithNewPkg)
		floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsPass)

		var stdout bytes.Buffer
		result, err := RunRatchet(RatchetOptions{
			Profile: profilePath,
			Floors:  floorsPath,
			Update:  true,
			Stdout:  &stdout,
		})
		if err != nil {
			t.Fatalf("RunRatchet error: %v", err)
		}
		if result.ExitCode != 1 {
			t.Errorf("exit code: got %d, want 1", result.ExitCode)
		}
		if result.Updated {
			t.Error("Updated: got true, want false")
		}
		if got := mustReadFile(t, floorsPath); got != testFloorsPass {
			t.Errorf("baseline modified on refusal: got %q, want %q", got, testFloorsPass)
		}
		out := stdout.String()
		for _, want := range []string{"-  refused", "baseline not modified: " + floorsPath} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q; got:\n%s", want, out)
			}
		}
	})

	t.Run("no_change_rewrites_byte_identically", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		profilePath := writeFixture(t, dir, "profile.out", testProfile)
		floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsPass)

		run := func() string {
			var stdout bytes.Buffer
			result, err := RunRatchet(RatchetOptions{
				Profile: profilePath,
				Floors:  floorsPath,
				Update:  true,
				Stdout:  &stdout,
			})
			if err != nil {
				t.Fatalf("RunRatchet error: %v", err)
			}
			if result.ExitCode != 0 {
				t.Fatalf("exit code: got %d, want 0", result.ExitCode)
			}
			if !result.Updated {
				t.Fatal("Updated: got false, want true")
			}
			return mustReadFile(t, floorsPath)
		}
		first := run()
		second := run()
		if first != second {
			t.Errorf("rewrite not byte-identical:\nfirst:  %q\nsecond: %q", first, second)
		}
		if !strings.Contains(first, "\"version\": 1") {
			t.Errorf("rewritten baseline missing version field: %q", first)
		}
	})
}

// failingWriter is an io.Writer whose Write always fails, exercising the
// report-write error path.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("writer closed")
}

// TestRunRatchet_ReportWriteFailure verifies the report-write error path: a
// failing writer maps to exit code 2 with the error returned.
func TestRunRatchet_ReportWriteFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	profilePath := writeFixture(t, dir, "profile.out", testProfile)
	floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsPass)

	result, err := RunRatchet(RatchetOptions{
		Profile: profilePath,
		Floors:  floorsPath,
		Stdout:  failingWriter{},
	})
	if err == nil {
		t.Fatal("RunRatchet with failing writer = nil error, want error")
	}
	if result.ExitCode != 2 {
		t.Errorf("exit code: got %d, want 2", result.ExitCode)
	}
}

// TestRunRatchet_UpdateBaselineWriteFailure verifies the update-mode
// baseline-write error path: a directory at the staged .tmp path makes the
// atomic write fail while the baseline itself remains readable, mapping to
// exit code 2 without marking the run as updated.
func TestRunRatchet_UpdateBaselineWriteFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	profilePath := writeFixture(t, dir, "profile.out", testProfile)
	floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsRaiseTarget)
	if err := os.Mkdir(floorsPath+".tmp", 0o755); err != nil {
		t.Fatalf("Mkdir staged path: %v", err)
	}

	var stdout bytes.Buffer
	result, err := RunRatchet(RatchetOptions{
		Profile: profilePath,
		Floors:  floorsPath,
		Update:  true,
		Stdout:  &stdout,
	})
	if err == nil {
		t.Fatal("RunRatchet update with unwritable staged baseline = nil error, want error")
	}
	if result.ExitCode != 2 {
		t.Errorf("exit code: got %d, want 2", result.ExitCode)
	}
	if result.Updated {
		t.Error("Updated: got true, want false")
	}
}

// TestRunRatchet_JSON verifies the --json report contract: per-package results
// sorted by package path, the overall verdict, null for absent measurements
// and floors, and a provenance envelope with producer, version, and an RFC
// 3339 UTC timestamp. It covers both check mode and update mode: a refused
// package renders floor:null with the update refused, and a raised row carries
// the post-update floor with updated:true.
func TestRunRatchet_JSON(t *testing.T) {
	t.Parallel()

	runJSON := func(t *testing.T, profile, floors string, update bool) (jsonReport, *bytes.Buffer) {
		t.Helper()
		dir := t.TempDir()
		profilePath := writeFixture(t, dir, "profile.out", profile)
		floorsPath := writeFixture(t, dir, "coverage-floors.json", floors)

		var stdout bytes.Buffer
		result, err := RunRatchet(RatchetOptions{
			Profile: profilePath,
			Floors:  floorsPath,
			Update:  update,
			JSON:    true,
			Stdout:  &stdout,
		})
		if err != nil {
			t.Fatalf("RunRatchet error: %v", err)
		}
		if result.ExitCode != 0 && result.ExitCode != 1 {
			t.Fatalf("exit code: got %d, want 0 or 1", result.ExitCode)
		}
		var report jsonReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("unmarshal report: %v\noutput:\n%s", err, stdout.String())
		}
		return report, &stdout
	}

	t.Run("pass_report_shape", func(t *testing.T) {
		t.Parallel()
		report, _ := runJSON(t, testProfile, testFloorsPass, false)

		if report.Verdict != "pass" {
			t.Errorf("verdict: got %q, want %q", report.Verdict, "pass")
		}
		if report.Updated {
			t.Error("updated: got true, want false")
		}
		wantOrder := []string{"cmd/coverage-ratchet", "internal/coverratchet", "metrics"}
		if len(report.Packages) != len(wantOrder) {
			t.Fatalf("packages: got %d entries, want %d", len(report.Packages), len(wantOrder))
		}
		for i, pkg := range wantOrder {
			if report.Packages[i].Package != pkg {
				t.Errorf("packages[%d]: got %q, want %q", i, report.Packages[i].Package, pkg)
			}
		}
		p := report.Packages[1] // internal/coverratchet: 40.0 measured, 40.0 floor
		if p.Measured == nil || *p.Measured != 40 {
			t.Errorf("measured: got %v, want 40", p.Measured)
		}
		if p.Floor == nil || *p.Floor != 40 {
			t.Errorf("floor: got %v, want 40", p.Floor)
		}
		if p.Status != "pass" {
			t.Errorf("status: got %q, want %q", p.Status, "pass")
		}
	})

	t.Run("stale_floor_null_measured", func(t *testing.T) {
		t.Parallel()
		report, _ := runJSON(t, testProfile, testFloorsStale, false)

		if report.Verdict != "fail" {
			t.Errorf("verdict: got %q, want %q", report.Verdict, "fail")
		}
		var gone *jsonPackage
		for i := range report.Packages {
			if report.Packages[i].Package == "internal/gone" {
				gone = &report.Packages[i]
			}
		}
		if gone == nil {
			t.Fatal("internal/gone missing from packages")
		}
		if gone.Measured != nil {
			t.Errorf("stale floor measured: got %v, want null", *gone.Measured)
		}
		if gone.Floor == nil || *gone.Floor != 50 {
			t.Errorf("stale floor floor: got %v, want 50", gone.Floor)
		}
		if gone.Status != "stale-floor" {
			t.Errorf("stale floor status: got %q, want %q", gone.Status, "stale-floor")
		}
	})

	t.Run("unregistered_null_floor", func(t *testing.T) {
		t.Parallel()
		report, _ := runJSON(t, testProfile, testFloorsUnregistered, false)

		var metrics *jsonPackage
		for i := range report.Packages {
			if report.Packages[i].Package == "metrics" {
				metrics = &report.Packages[i]
			}
		}
		if metrics == nil {
			t.Fatal("metrics missing from packages")
		}
		if metrics.Floor != nil {
			t.Errorf("unregistered floor: got %v, want null", *metrics.Floor)
		}
		if metrics.Measured == nil || *metrics.Measured != 60 {
			t.Errorf("unregistered measured: got %v, want 60", metrics.Measured)
		}
		if metrics.Status != "unregistered" {
			t.Errorf("unregistered status: got %q, want %q", metrics.Status, "unregistered")
		}
	})

	t.Run("update_mode_refused_null_floor", func(t *testing.T) {
		t.Parallel()
		report, _ := runJSON(t, testProfileWithNewPkg, testFloorsPass, true)

		if report.Verdict != "fail" {
			t.Errorf("verdict: got %q, want %q", report.Verdict, "fail")
		}
		if report.Updated {
			t.Error("updated: got true, want false")
		}
		var newpkg *jsonPackage
		for i := range report.Packages {
			if report.Packages[i].Package == "cmd/newpkg" {
				newpkg = &report.Packages[i]
			}
		}
		if newpkg == nil {
			t.Fatal("cmd/newpkg missing from packages")
		}
		if newpkg.Floor != nil {
			t.Errorf("refused floor: got %v, want null", *newpkg.Floor)
		}
		if newpkg.Measured == nil || *newpkg.Measured != 50 {
			t.Errorf("refused measured: got %v, want 50", newpkg.Measured)
		}
		if newpkg.Status != "refused" {
			t.Errorf("refused status: got %q, want %q", newpkg.Status, "refused")
		}
	})

	t.Run("update_mode_raises_floor_to_measured", func(t *testing.T) {
		t.Parallel()
		report, _ := runJSON(t, testProfile, testFloorsRaiseTarget, true)

		if report.Verdict != "pass" {
			t.Errorf("verdict: got %q, want %q", report.Verdict, "pass")
		}
		if !report.Updated {
			t.Error("updated: got false, want true")
		}
		var engine *jsonPackage
		for i := range report.Packages {
			if report.Packages[i].Package == "internal/coverratchet" {
				engine = &report.Packages[i]
			}
		}
		if engine == nil {
			t.Fatal("internal/coverratchet missing from packages")
		}
		if engine.Status != "raised" {
			t.Errorf("status: got %q, want %q", engine.Status, "raised")
		}
		if engine.Floor == nil || *engine.Floor != 40 {
			t.Errorf("raised floor: got %v, want 40 (the post-update floor)", engine.Floor)
		}
		if engine.Measured == nil || *engine.Measured != 40 {
			t.Errorf("raised measured: got %v, want 40", engine.Measured)
		}
	})

	t.Run("provenance_envelope", func(t *testing.T) {
		t.Parallel()
		report, stdout := runJSON(t, testProfile, testFloorsPass, false)

		if report.Provenance == nil {
			t.Fatal("provenance: got nil, want envelope")
		}
		if report.Provenance.Producer != producerName {
			t.Errorf("producer: got %q, want %q", report.Provenance.Producer, producerName)
		}
		if report.Provenance.Version == "" {
			t.Error("version: got empty, want non-empty")
		}
		ts, err := time.Parse(time.RFC3339, report.Provenance.GeneratedAt)
		if err != nil {
			t.Errorf("generatedAt not RFC 3339: %v", err)
		} else if ts.Location() != time.UTC {
			t.Errorf("generatedAt not UTC: got %v", ts.Location())
		}
		if !strings.Contains(stdout.String(), "\"provenance\"") {
			t.Errorf("output missing provenance key; got:\n%s", stdout.String())
		}
	})
}

// TestRootCmd_Execute_BelowFloor exercises the full cobra command path for a
// policy failure: the table goes to stdout, cobra reports the summary error to
// stderr, and the error maps to exit code 1.
func TestRootCmd_Execute_BelowFloor(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	profilePath := writeFixture(t, dir, "profile.out", testProfile)
	floorsPath := writeFixture(t, dir, "coverage-floors.json", testFloorsBelowFloor)

	cmd := rootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--profile", profilePath, "--floors", floorsPath})

	err := cmd.Execute()
	var stderrOut bytes.Buffer
	code := exitCode(err, &stderrOut)
	if code != 1 {
		t.Errorf("exit code: got %d, want 1 (err: %v)", code, err)
	}
	if !strings.Contains(stdout.String(), "below-floor") {
		t.Errorf("stdout missing below-floor; got:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Error: coverage floor violations detected") {
		t.Errorf("stderr missing summary error; got:\n%s", stderr.String())
	}
	if got := mustReadFile(t, floorsPath); got != testFloorsBelowFloor {
		t.Errorf("baseline modified in check mode: got %q, want %q", got, testFloorsBelowFloor)
	}
}
