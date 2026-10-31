package ripwireadapter

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Subprocess-runner tests — openspec/changes/ripwire-adapter task 2.3.
//
// They pin the runner's hygiene contract: sanitized environment (no host
// credentials reach the subprocess), working directory at the crawl root,
// bounded capture (oversized output errors, never truncation), and context
// cancellation that terminates the subprocess without orphaning it.

// installProbeScript installs a tiny probe executable named ripwire on PATH
// that reports whether GITHUB_TOKEN is visible in its environment and
// echoes its working directory — proof by observation for the env and
// cmd.Dir contracts.
func installProbeScript(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	bin := filepath.Join(dir, "ripwire")
	script := "#!/bin/sh\n" +
		"if [ -n \"${GITHUB_TOKEN:-}\" ]; then echo TOKEN-PRESENT; else echo TOKEN-ABSENT; fi\n" +
		"pwd\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("installProbeScript: write %s: %v", bin, err)
	}
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("installProbeScript: chmod %s: %v", bin, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunner_EnvironmentAndWorkingDirectory(t *testing.T) {
	// Not parallel: installs PATH state and sets a host credential.
	installProbeScript(t)
	t.Setenv("GITHUB_TOKEN", "definitely-a-secret")

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("probe not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)

	crawlRoot := filepath.Join("testdata", "python", "src")
	wantDir, err := filepath.Abs(crawlRoot)
	if err != nil {
		t.Fatalf("resolve crawl root: %v", err)
	}
	wantDir, err = filepath.EvalSymlinks(wantDir)
	if err != nil {
		t.Fatalf("resolve crawl root symlinks: %v", err)
	}

	stdout, err := runner.run(t.Context(), crawlRoot)
	if err != nil {
		t.Fatalf("run probe: %v", err)
	}
	got := strings.SplitN(strings.TrimSpace(string(stdout)), "\n", 2)
	if len(got) != 2 {
		t.Fatalf("probe output: got %q, want token line and pwd line", stdout)
	}
	if got[0] != "TOKEN-ABSENT" {
		t.Errorf("GITHUB_TOKEN reached the subprocess: probe reported %q", got[0])
	}
	gotDir, err := filepath.EvalSymlinks(strings.TrimSpace(got[1]))
	if err != nil {
		t.Fatalf("resolve probe pwd: %v", err)
	}
	if gotDir != wantDir {
		t.Errorf("working directory: got %q, want %q", gotDir, wantDir)
	}
}

func TestRunner_CancelledContext(t *testing.T) {
	// Not parallel: installs PATH state.
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "python"), "valid")

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("fake not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	stdout, err := runner.run(ctx, ".", "--version")
	if stdout != nil {
		t.Fatalf("cancelled context: got stdout %d bytes, want nil", len(stdout))
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context: got error %v, want one wrapping context.Canceled", err)
	}

	// No orphaned process: a subsequent healthy run through the same
	// runner still completes.
	stdout, err = runner.run(t.Context(), ".", "--version")
	if err != nil {
		t.Fatalf("run after cancellation: %v", err)
	}
	want := readRecording(t, filepath.Join("testdata", "recorded", "python", "version.txt"))
	if string(stdout) != string(want) {
		t.Errorf("run after cancellation: stdout got %q, want recorded version bytes", stdout)
	}
}

func TestRunner_OversizedOutput(t *testing.T) {
	// Not parallel: installs PATH state.
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "python"), "valid")

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("fake not on PATH: %v", err)
	}
	// Shrink the bound below the 91-byte version line: the field exists
	// for exactly this test.
	runner := &ripwireRunner{binary: bin, maxOutputBytes: 8}

	stdout, err := runner.run(t.Context(), ".", "--version")
	if stdout != nil {
		t.Fatalf("oversized output: got stdout %d bytes, want nil (never truncated)", len(stdout))
	}
	if !errors.Is(err, errOutputOversized) {
		t.Fatalf("oversized output: got error %v, want errOutputOversized", err)
	}
	if !strings.Contains(err.Error(), "8") {
		t.Errorf("oversized output: error %q does not name the bound", err)
	}
}

func TestRunner_NonzeroExit(t *testing.T) {
	// Not parallel: installs PATH state.
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "python"), "valid")

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("fake not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)

	stdout, err := runner.run(t.Context(), ".", "--version", "--bogus-flag")
	if stdout != nil {
		t.Fatalf("nonzero exit: got stdout %d bytes, want nil", len(stdout))
	}
	if err == nil {
		t.Fatal("nonzero exit: got nil error")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("nonzero exit: got error %v, want one wrapping exec.ExitError", err)
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("nonzero exit: error %q does not carry the stderr preview", err)
	}
}

func TestRunner_ReplaysRecordings(t *testing.T) {
	// Not parallel: installs PATH state.
	for _, language := range []string{"python", "typescript"} {
		t.Run(language, func(t *testing.T) {
			recordings := filepath.Join("testdata", "recorded", language)
			installFakeRipwireWrapper(t, recordings, "valid")

			bin, err := exec.LookPath("ripwire")
			if err != nil {
				t.Fatalf("fake not on PATH: %v", err)
			}
			runner := newRipwireRunner(bin)

			for _, tc := range []struct {
				args      []string
				recording string
			}{
				{[]string{"--version"}, "version.txt"},
				{[]string{"--arch=some-rules.txt"}, "arch.txt"},
				{[]string{"--deps", "--limit=1000"}, "deps.xml"},
				{[]string{"--export=cc.json"}, "cc.json"},
			} {
				stdout, err := runner.run(t.Context(), ".", tc.args...)
				if err != nil {
					t.Fatalf("run %v: %v", tc.args, err)
				}
				want := readRecording(t, filepath.Join(recordings, tc.recording))
				if string(stdout) != string(want) {
					t.Errorf("run %v: stdout got %d bytes, want %d recorded bytes",
						tc.args, len(stdout), len(want))
				}
			}
		})
	}
}
