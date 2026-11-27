package ripwireadapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Subprocess-runner tests — openspec/changes/archive/2026-11-24-ripwire-adapter task 2.3.
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

// TestBoundedWriter_Write pins the capture-bound boundary at the Write
// level: a write that lands exactly on the limit succeeds (the runner
// promises output at or below the bound is returned verbatim), only a
// write past it fails, and a failing write marks the writer oversized,
// invokes onOversize once, and buffers nothing past the bound.
func TestBoundedWriter_Write(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		limit         int64
		writes        [][]byte
		wantLastN     int
		wantLastErr   error
		wantOversized bool
		wantBuf       string
	}{
		{
			name:      "write under the limit is buffered verbatim",
			limit:     8,
			writes:    [][]byte{[]byte("abc")},
			wantLastN: 3,
			wantBuf:   "abc",
		},
		{
			name:      "writes landing exactly on the limit succeed",
			limit:     8,
			writes:    [][]byte{[]byte("abc"), []byte("defgh")},
			wantLastN: 5,
			wantBuf:   "abcdefgh",
		},
		{
			name:          "write past the limit fails without buffering",
			limit:         8,
			writes:        [][]byte{[]byte("abc"), []byte("defghijklm")},
			wantLastN:     0,
			wantLastErr:   errOutputOversized,
			wantOversized: true,
			wantBuf:       "abc",
		},
		{
			name:          "write to a buffer already at the limit fails",
			limit:         8,
			writes:        [][]byte{[]byte("abcdefgh"), []byte("x")},
			wantLastN:     0,
			wantLastErr:   errOutputOversized,
			wantOversized: true,
			wantBuf:       "abcdefgh",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kills := 0
			w := &boundedWriter{limit: tc.limit, onOversize: func() { kills++ }}
			var lastN int
			var lastErr error
			for _, p := range tc.writes {
				lastN, lastErr = w.Write(p)
			}
			if tc.wantLastErr != nil {
				if !errors.Is(lastErr, tc.wantLastErr) {
					t.Fatalf("Write: got error %v, want %v", lastErr, tc.wantLastErr)
				}
			} else if lastErr != nil {
				t.Fatalf("Write: got error %v, want nil", lastErr)
			}
			if lastN != tc.wantLastN {
				t.Errorf("Write return: got %d, want %d", lastN, tc.wantLastN)
			}
			if w.oversized != tc.wantOversized {
				t.Errorf("oversized flag: got %v, want %v", w.oversized, tc.wantOversized)
			}
			if string(w.buf) != tc.wantBuf {
				t.Errorf("buffered bytes: got %q, want %q", w.buf, tc.wantBuf)
			}
			wantKills := 0
			if tc.wantLastErr != nil {
				wantKills = 1
			}
			if kills != wantKills {
				t.Errorf("onOversize calls: got %d, want %d", kills, wantKills)
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

// runDeadline bounds how long a runner test waits for run to return. It sits
// far above the runner's wait delay plus race-detector slack, and far below
// the package timeout, so a hung run fails its own test quickly.
const runDeadline = 20 * time.Second

// testWaitDelay is the shrunk wait delay for pipe-hang regression tests
// (design D2): far above process-spawn jitter and pipe teardown, far below
// the production default, so each test spends milliseconds waiting for the
// mechanism under test rather than the full production delay.
const testWaitDelay = 200 * time.Millisecond

// maxReturnTime bounds how long a pipe-hang regression test's run may take
// to return once its mechanism (cancellation or the wait delay) has fired:
// the shrunk test delay plus generous scheduling and race-detector slack,
// far below runDeadline, so a regression that inflates the delay fails the
// timing assertion instead of passing inside the hang deadline.
const maxReturnTime = 5 * time.Second

// runWithin calls runner.run in a goroutine and fails the test if it has not
// returned within runDeadline, so a hang (the failure mode of the
// runner-pipe-hang change) is reported by name instead of consuming the
// package timeout.
func runWithin(ctx context.Context, t *testing.T, runner *ripwireRunner, args ...string) ([]byte, error) {
	t.Helper()

	type result struct {
		stdout []byte
		err    error
	}
	done := make(chan result, 1)
	go func() {
		stdout, err := runner.run(ctx, ".", args...)
		done <- result{stdout, err}
	}()
	select {
	case r := <-done:
		return r.stdout, r.err
	case <-time.After(runDeadline):
		t.Fatalf("run %v did not return within %v: hung on a pipe held open by a child process", args, runDeadline)
		return nil, nil
	}
}

// installRipwireScript installs a /bin/sh script named ripwire on PATH with
// the given body and returns the directory it lives in, for tests that
// need a scratch file next to the script.
func installRipwireScript(t *testing.T, body string) string {
	t.Helper()

	dir := t.TempDir()
	bin := filepath.Join(dir, "ripwire")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("installRipwireScript: write %s: %v", bin, err)
	}
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("installRipwireScript: chmod %s: %v", bin, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

// killChildOnCleanup registers a cleanup that kills the process whose PID a
// script wrote to pidFile, so a pipe-holding child a test started never
// outlives the test. A missing or unreadable file means the child never
// started, and there is nothing to kill.
func killChildOnCleanup(t *testing.T, pidFile string) {
	t.Helper()

	t.Cleanup(func() {
		raw, err := os.ReadFile(pidFile)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
		if err != nil {
			return
		}
		if p, err := os.FindProcess(pid); err == nil {
			_ = p.Kill()
		}
	})
}

// installFloodScript installs a probe executable named ripwire on PATH that
// writes blob to the named stream ("stdout" or "stderr") and nothing to the
// other stream, so the runner's oversized-capture error can be driven per
// stream. The writer is a backgrounded child the script waits on, so the
// child (not ripwire itself) holds both pipes: killing the script leaves it
// running. The blob is expected to exceed the OS pipe buffer, so the child
// can never drain it unread and exit on its own; only closing the crossed
// pipe stops it.
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
		// 3>&1 keeps a copy of the stdout pipe open in the child: a bare
		// >&2 would replace it, and the child would hold only stderr.
		redirect = "3>&1 >&2"
	}
	script := "#!/bin/sh\ncat " + shellSingleQuote(blobFile) + " " + redirect + " &\nwait\n"
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
	// The blob is far larger than any OS pipe buffer, so the backgrounded
	// writer can never finish on its own: the run returns only if crossing
	// the bound also releases the child's hold on the pipes.
	blob := bytes.Repeat([]byte("x"), 1<<20)

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
			// exactly this test. The constructor's bounded pipe wait still
			// applies.
			runner := newRipwireRunner(bin)
			runner.maxOutputBytes = bound

			stdout, err := runWithin(t.Context(), t, runner)
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

// TestRunner_CancelWithPipeHoldingChild pins the cancellation path when a
// child of ripwire holds both output pipes: killing ripwire alone never
// delivers EOF, so the run must still return, with an error wrapping
// context.Canceled and no stdout, instead of waiting for the child.
func TestRunner_CancelWithPipeHoldingChild(t *testing.T) {
	// Not parallel: installs PATH state.
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	installRipwireScript(t, "sleep 60 &\necho $! > "+shellSingleQuote(pidFile)+"\nwait\n")
	killChildOnCleanup(t, pidFile)

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("script not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)
	runner.waitDelay = testWaitDelay

	// Cancel only once the child exists, so the run is in flight (not the
	// pre-start fast path) and the child already holds the pipes.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() {
		deadline := time.Now().Add(runDeadline)
		for time.Now().Before(deadline) {
			if raw, err := os.ReadFile(pidFile); err == nil && len(strings.TrimSpace(string(raw))) > 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
	}()

	start := time.Now()
	stdout, err := runWithin(ctx, t, runner)
	if stdout != nil {
		t.Fatalf("cancelled run: got stdout %d bytes, want nil", len(stdout))
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled run: got error %v, want one wrapping context.Canceled", err)
	}
	if elapsed := time.Since(start); elapsed > maxReturnTime {
		t.Errorf("cancelled run: returned after %v, want within %v (wait delay plus scheduling slack)", elapsed, maxReturnTime)
	}
}

// TestRunner_SuccessfulExitWithAbandonedPipes pins the fail-closed rule for a
// ripwire that exits 0 while a child it started still holds its output
// pipes: the output may be incomplete, so the run must return an error
// wrapping exec.ErrWaitDelay and no stdout, and must not wait for the child.
func TestRunner_SuccessfulExitWithAbandonedPipes(t *testing.T) {
	// Not parallel: installs PATH state.
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	installRipwireScript(t, "sleep 60 &\necho $! > "+shellSingleQuote(pidFile)+"\necho partial\n")
	killChildOnCleanup(t, pidFile)

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("script not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)
	runner.waitDelay = testWaitDelay

	start := time.Now()
	stdout, err := runWithin(t.Context(), t, runner)
	if stdout != nil {
		t.Fatalf("abandoned pipes: got stdout %q, want nil (fail closed)", stdout)
	}
	if !errors.Is(err, exec.ErrWaitDelay) {
		t.Fatalf("abandoned pipes: got error %v, want one wrapping exec.ErrWaitDelay", err)
	}
	if elapsed := time.Since(start); elapsed > maxReturnTime {
		t.Errorf("abandoned pipes: returned after %v, want within %v (wait delay plus scheduling slack)", elapsed, maxReturnTime)
	}
}

// TestRunner_ZeroWaitDelayFallsBackToDefault pins the use-site fallback: a
// runner built as a struct literal, bypassing newRipwireRunner, leaves
// waitDelay zero, and its run must still bound the pipe wait at the package
// default rather than hang on a child that holds the output pipes.
func TestRunner_ZeroWaitDelayFallsBackToDefault(t *testing.T) {
	// Not parallel: installs PATH state.
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	installRipwireScript(t, "sleep 60 &\necho $! > "+shellSingleQuote(pidFile)+"\necho partial\n")
	killChildOnCleanup(t, pidFile)

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("script not on PATH: %v", err)
	}
	runner := &ripwireRunner{binary: bin, maxOutputBytes: defaultMaxOutputBytes}

	start := time.Now()
	stdout, err := runWithin(t.Context(), t, runner)
	if stdout != nil {
		t.Fatalf("zero wait delay: got stdout %q, want nil (fail closed)", stdout)
	}
	if !errors.Is(err, exec.ErrWaitDelay) {
		t.Fatalf("zero wait delay: got error %v, want one wrapping exec.ErrWaitDelay (bounded by the default)", err)
	}
	if elapsed := time.Since(start); elapsed < defaultWaitDelay {
		t.Errorf("zero wait delay: returned after %v, want at least the default %v (the fallback was not applied)", elapsed, defaultWaitDelay)
	}
	if elapsed := time.Since(start); elapsed > maxReturnTime {
		t.Errorf("zero wait delay: returned after %v, want within %v (default wait delay plus scheduling slack)", elapsed, maxReturnTime)
	}
}

// installFloodWithHolderScript installs a probe executable named ripwire
// on PATH that floods stdout past any test capture bound while a sibling
// sleeper holds both output pipes open without writing. The flood drives
// the oversize path; the sleeper — whose PID the script writes to the
// returned file — keeps the pipes held past the wait delay even after the
// flood writer dies, so oversize and abandoned-pipe conditions co-occur.
// The sleeper starts and its PID is written before the flood begins, so the
// oversize kill of the script can never pre-empt the PID write: the file
// always appears and cleanup always reaps the sleeper.
func installFloodWithHolderScript(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	blobFile := filepath.Join(dir, "blob")
	if err := os.WriteFile(blobFile, bytes.Repeat([]byte("x"), 1<<20), 0o644); err != nil {
		t.Fatalf("installFloodWithHolderScript: write %s: %v", blobFile, err)
	}
	pidFile := filepath.Join(dir, "child.pid")
	installRipwireScript(t, "sleep 60 &\necho $! > "+shellSingleQuote(pidFile)+"\ncat "+shellSingleQuote(blobFile)+" >&1 &\nwait\n")
	killChildOnCleanup(t, pidFile)
	return pidFile
}

// TestRunner_OversizeWinsOverWaitDelay pins the error-classification order
// when oversize and abandoned pipes co-occur (design D3): the flood crosses
// the capture bound while the sibling sleeper holds the pipes past the wait
// delay, so Wait reports exec.ErrWaitDelay underneath — and the run must
// still classify the failure as oversize.
func TestRunner_OversizeWinsOverWaitDelay(t *testing.T) {
	// Not parallel: installs PATH state.
	installFloodWithHolderScript(t)

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("script not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)
	runner.maxOutputBytes = 8
	runner.waitDelay = testWaitDelay

	stdout, err := runWithin(t.Context(), t, runner)
	if stdout != nil {
		t.Fatalf("oversize with held pipes: got stdout %d bytes, want nil", len(stdout))
	}
	if !errors.Is(err, errOutputOversized) {
		t.Fatalf("oversize with held pipes: got error %v, want errOutputOversized (outranks exec.ErrWaitDelay)", err)
	}
}

// cancelWindowWaitDelay is the wait delay for
// TestRunner_CancellationWinsOverOversize: long enough that a cancel issued
// cancelAfterHolder after the holder starts always lands while Wait is
// still blocked on the held pipes, short enough to stay under maxReturnTime.
const cancelWindowWaitDelay = 2 * time.Second

// cancelAfterHolder is how long TestRunner_CancellationWinsOverOversize
// waits after the holder's PID appears before cancelling, giving the flood
// time to cross the capture bound so both failure modes co-occur.
const cancelAfterHolder = 300 * time.Millisecond

// TestRunner_CancellationWinsOverOversize pins the error-classification
// order when cancellation and oversize co-occur (design D3): the flood
// crosses the capture bound, the sibling sleeper holds the pipes, and the
// context is cancelled mid-flight — the run must classify the failure as
// context.Canceled, and must return within the wait delay plus slack.
func TestRunner_CancellationWinsOverOversize(t *testing.T) {
	// Not parallel: installs PATH state.
	pidFile := installFloodWithHolderScript(t)

	bin, err := exec.LookPath("ripwire")
	if err != nil {
		t.Fatalf("script not on PATH: %v", err)
	}
	runner := newRipwireRunner(bin)
	runner.maxOutputBytes = 8
	// A long wait delay opens a wide window: after the oversize kill, Wait
	// blocks until the delay expires because the sleeper holds the pipes.
	runner.waitDelay = cancelWindowWaitDelay

	// Cancel a fixed interval after the holder exists. The holder's PID is
	// written before the flood starts, so the trigger cannot be destroyed
	// by the oversize kill. The cancel lands inside the wait-delay window
	// whether or not the flood has crossed the bound yet, and the runner
	// checks the context only after Wait returns, so the expected error is
	// context.Canceled on every interleaving — precedence decides it.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() {
		deadline := time.Now().Add(runDeadline)
		for time.Now().Before(deadline) {
			if raw, err := os.ReadFile(pidFile); err == nil && len(strings.TrimSpace(string(raw))) > 0 {
				time.Sleep(cancelAfterHolder)
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
	}()

	start := time.Now()
	stdout, err := runWithin(ctx, t, runner)
	if stdout != nil {
		t.Fatalf("cancelled flood: got stdout %d bytes, want nil", len(stdout))
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled flood: got error %v, want one wrapping context.Canceled (outranks oversize)", err)
	}
	if elapsed := time.Since(start); elapsed > maxReturnTime {
		t.Errorf("cancelled flood: returned after %v, want within %v (wait delay plus scheduling slack)", elapsed, maxReturnTime)
	}
}
