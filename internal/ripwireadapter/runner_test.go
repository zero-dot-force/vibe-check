package ripwireadapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

// TestPreviewOutput pins the shared preview helper's contract: bounded
// output (cap bytes plus a trailing "..."), %q quoting, the "<empty>"
// fallback, the whitespace-trim policy, and the accepted UTF-8 rune split
// at the cap. Both the --version parser and the failed-run stderr
// diagnostic render through this helper, so the contract lives here once.
func TestPreviewOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		limit int
		want  string
	}{
		{name: "short input passes through quoted", input: "unknown flag", limit: maxStderrPreviewBytes, want: `"unknown flag"`},
		{name: "input over the cap renders cap bytes then ellipsis", input: strings.Repeat("a", 200), limit: maxVersionPreviewBytes, want: `"` + strings.Repeat("a", maxVersionPreviewBytes) + `..."`},
		{name: "empty input renders the empty placeholder", input: "", limit: maxStderrPreviewBytes, want: `"<empty>"`},
		{name: "whitespace-only input renders the empty placeholder", input: " \n\t ", limit: maxStderrPreviewBytes, want: `"<empty>"`},
		{name: "surrounding whitespace is trimmed before quoting", input: "  ok  ", limit: maxStderrPreviewBytes, want: `"ok"`},
		{name: "cap may split a multi-byte rune, rendered as an escape", input: "abécd", limit: 3, want: "\"ab\\xc3...\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := previewOutput(tc.input, tc.limit); got != tc.want {
				t.Errorf("previewOutput(%q, %d) = %s, want %s", tc.input, tc.limit, got, tc.want)
			}
		})
	}
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

// installFloodScript installs a probe executable named ripwire on PATH that
// writes blob to the named stream ("stdout" or "stderr") and nothing to the
// other stream, so the runner's oversized-capture error can be driven per
// stream. The blob is expected to exceed the OS pipe buffer, which exercises
// the concurrent drain: the first bounded reader to cross the bound kills the
// process so the sibling reader unblocks.
func installFloodScript(t *testing.T, stream string, blob []byte) {
	t.Helper()

	dir := t.TempDir()
	bin := filepath.Join(dir, "ripwire")
	blobFile := filepath.Join(dir, "blob")
	if err := os.WriteFile(blobFile, blob, 0o644); err != nil {
		t.Fatalf("installFloodScript: write %s: %v", blobFile, err)
	}
	redirect := ">&1"
	if stream == "stderr" {
		redirect = ">&2"
	}
	script := "#!/bin/sh\ncat \"" + blobFile + "\" " + redirect + "\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("installFloodScript: write %s: %v", bin, err)
	}
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("installFloodScript: chmod %s: %v", bin, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestRunner_OversizedOutput pins the oversized-capture error for BOTH
// streams: the error wraps errOutputOversized, names the crossed bound
// ("more than N bytes"), names the stream that crossed it ("bytes of
// stdout" / "bytes of stderr"), and the stdout capture is nil rather than
// truncated. The label assertion also pins the label-selection expression
// against a stream-swap regression.
func TestRunner_OversizedOutput(t *testing.T) {
	// Not parallel: installs PATH state.
	const bound = 8
	// The blob exceeds the OS pipe buffer, so the writer blocks until the
	// bounded reader crosses the bound and kills the process.
	blob := bytes.Repeat([]byte("x"), 128<<10)

	cases := []struct {
		name      string
		stream    string
		wantLabel string
	}{
		{name: "stdout crossing the bound", stream: "stdout", wantLabel: "bytes of stdout"},
		{name: "stderr crossing the bound", stream: "stderr", wantLabel: "bytes of stderr"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installFloodScript(t, tc.stream, blob)

			bin, err := exec.LookPath("ripwire")
			if err != nil {
				t.Fatalf("flood script not on PATH: %v", err)
			}
			// Shrink the bound below the blob size: the field exists for
			// exactly this test.
			runner := &ripwireRunner{binary: bin, maxOutputBytes: bound}

			stdout, err := runner.run(t.Context(), ".")
			if stdout != nil {
				t.Fatalf("oversized output: got stdout %d bytes, want nil (never truncated)", len(stdout))
			}
			if !errors.Is(err, errOutputOversized) {
				t.Fatalf("oversized output: got error %v, want errOutputOversized", err)
			}
			if !strings.Contains(err.Error(), tc.wantLabel) {
				t.Errorf("oversized output: error %q does not name %q", err, tc.wantLabel)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("more than %d bytes", bound)) {
				t.Errorf("oversized output: error %q does not name the bound", err)
			}
		})
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
