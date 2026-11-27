package ripwireadapter

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// defaultMaxOutputBytes bounds how much output the runner accepts from a
// single ripwire invocation on EACH stream: stdout and stderr are captured
// through independent bounded readers, so one invocation can deliver at
// most twice this bound. It is sized for the largest expected verb output
// (--export=cc.json on a large repository) while keeping a misbehaving
// binary from exhausting analyzer memory. Output at or below the bound is
// returned verbatim; anything larger is an error, never a silent
// truncation.
const defaultMaxOutputBytes = 16 << 20 // 16 MiB

// errOutputOversized is returned (wrapped) when a ripwire invocation
// produces more output than the runner's bound allows.
var errOutputOversized = errors.New("ripwire output exceeds capture bound")

// defaultWaitDelay bounds how long the runner waits, once ripwire has exited
// or the context is done, for ripwire's output pipes to close. A process that
// ripwire starts inherits both pipes, so without this bound one that
// outlives ripwire would keep the run from ever returning. When the delay
// elapses, os/exec closes the pipes itself (cmd.WaitDelay, golang/go#23019).
// Two seconds is ample for an exiting ripwire to flush and small next to an
// analysis.
const defaultWaitDelay = 2 * time.Second

// ripwireRunner executes the resolved ripwire binary with the adapter's
// subprocess hygiene: a sanitized environment (no host credentials), a
// caller-provided working directory (the crawl root), bounded stdout and
// stderr capture, context cancellation that terminates the subprocess, and
// a bounded wait for the output pipes so every run returns in bounded time.
type ripwireRunner struct {
	// binary is the PATH-resolved ripwire executable.
	binary string
	// maxOutputBytes is the per-stream capture bound applied to both the
	// stdout and the stderr pipe; it exists as a field so tests can shrink
	// it. Production callers use the package default.
	maxOutputBytes int64
	// waitDelay is applied as cmd.WaitDelay: how long the runner waits for
	// the output pipes to close after ripwire exits or the context is done.
	// It exists as a field so tests can shrink it. A zero or negative value
	// falls back to the package default at the use site, so a runner built
	// without newRipwireRunner keeps the bounded wait.
	waitDelay time.Duration
}

// newRipwireRunner builds a runner for the given binary with the default
// output bound and wait delay.
func newRipwireRunner(binary string) *ripwireRunner {
	return &ripwireRunner{binary: binary, maxOutputBytes: defaultMaxOutputBytes, waitDelay: defaultWaitDelay}
}

// boundedWriter captures one output stream of a ripwire run, up to limit
// bytes. A write that would take the capture past limit is not buffered: the
// writer marks itself oversized, calls onOversize (which kills the process),
// and fails the write. os/exec then stops copying that stream and closes its
// pipe, so a process still writing to it gets EPIPE instead of blocking.
//
// os/exec calls Write from a single copying goroutine per stream, and Cmd.Wait
// returns only after that goroutine finishes, so the fields need no locking
// when read after Wait.
type boundedWriter struct {
	buf        []byte
	limit      int64
	oversized  bool
	onOversize func()
}

// Write implements io.Writer for one captured stream.
func (w *boundedWriter) Write(p []byte) (int, error) {
	if int64(len(w.buf))+int64(len(p)) > w.limit {
		w.oversized = true
		w.onOversize()
		return 0, errOutputOversized
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

// ripwireEnv returns the environment every ripwire subprocess runs with:
// the sanitized baseline (PATH/HOME/LANG) and nothing else, so no host
// credentials can reach the analyzer subprocess.
func ripwireEnv() []string {
	return metrics.SanitizeEnvironment(nil)
}

// run executes the binary with args, with the process working directory
// set to dir (the crawl root; empty inherits the caller's). It returns the
// captured stdout. Failures are returned as errors:
//
//   - a cancelled context returns an error wrapping the context error,
//     with the subprocess terminated (never orphaned);
//   - a nonzero exit returns an error wrapping the exit error plus a
//     bounded stderr preview;
//   - either stream (stdout or stderr) exceeding the capture bound returns
//     an error naming the bound and the stream — output is never truncated;
//   - a successful exit whose output pipes are still held open after the
//     wait delay (a process ripwire started kept them) returns an error
//     wrapping exec.ErrWaitDelay, since the captured output may be
//     incomplete.
//
// Every path returns in bounded time: once ripwire exits or the context is
// done, os/exec closes any pipe still open after the wait delay.
func (r *ripwireRunner) run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	// Fast path: refuse to start at all when the context is already dead.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("ripwire runner: %w", err)
	}

	// A zero or negative waitDelay falls back to the package default, so
	// every construction path — including struct literals that bypass
	// newRipwireRunner — keeps a bounded pipe wait.
	waitDelay := r.waitDelay
	if waitDelay <= 0 {
		waitDelay = defaultWaitDelay
	}
	cmd := exec.CommandContext(ctx, r.binary, args...)
	cmd.Dir = dir
	cmd.Env = ripwireEnv()
	cmd.WaitDelay = waitDelay

	// The first stream to cross the bound kills the process. The kill is
	// shared so it happens once, and os/exec closes the crossed stream's
	// pipe when its writer fails, so nothing blocks writing to it.
	var killOnce sync.Once
	kill := func() { killOnce.Do(func() { _ = cmd.Process.Kill() }) }
	stdout := &boundedWriter{limit: r.maxOutputBytes, onOversize: kill}
	stderr := &boundedWriter{limit: r.maxOutputBytes, onOversize: kill}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ripwire runner: start %s: %w", r.binary, err)
	}
	waitErr := cmd.Wait()

	// Cancellation wins over every other failure mode so callers can
	// reliably classify it with errors.Is(err, context.Canceled).
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("ripwire runner: %w (wait: %v)", ctxErr, waitErr)
	}
	// Oversize outranks the exit status: crossing the bound kills the
	// process ourselves, so the wait error is just our own kill signal.
	if stdout.oversized || stderr.oversized {
		return nil, fmt.Errorf("%w: %s %v produced more than %d bytes of %s",
			errOutputOversized, r.binary, args, r.maxOutputBytes,
			map[bool]string{true: "stdout", false: "stderr"}[stdout.oversized])
	}
	// Abandoned pipes: ripwire exited 0, but something it started still
	// held its output when the wait delay expired, so os/exec closed the
	// pipes under it and the capture may be incomplete. Fail closed.
	if errors.Is(waitErr, exec.ErrWaitDelay) {
		return nil, fmt.Errorf("ripwire runner: %s %v exited but left its output pipes open past %v: %w",
			r.binary, args, waitDelay, waitErr)
	}
	if waitErr != nil {
		return nil, fmt.Errorf("ripwire runner: %s %v: %w; stderr: %s",
			r.binary, args, waitErr, previewOutput(string(stderr.buf), maxStderrPreviewBytes))
	}
	return stdout.buf, nil
}

// maxStderrPreviewBytes bounds the preview of multi-line stderr from a
// failed ripwire run: larger than the version preview because stderr
// diagnostics span several lines, still small enough that a hostile binary
// cannot flood a log through a failed run.
const maxStderrPreviewBytes = 512

// maxVersionPreviewBytes bounds the preview of a single-line --version
// payload: enough to name a recognizable prefix, small enough to keep
// unparseable-version errors readable.
const maxVersionPreviewBytes = 120

// previewOutput renders a bounded, quoted preview of subprocess output for
// error messages. It exists so a hostile or broken binary cannot flood the
// error (and any log it lands in) with arbitrary bytes: the preview is
// trimmed of surrounding whitespace, capped at limit bytes (a trailing
// "..." marks the cut), quoted with %q so control bytes render visibly,
// and an empty or whitespace-only result renders as "<empty>". The byte
// cap can split a multi-byte UTF-8 rune; %q renders the dangling byte as
// an escape sequence, so the split is accepted rather than paying for rune
// alignment on a diagnostic path. Both ripwire subprocess consumers — the
// --version parser and the failed-run stderr diagnostic — share this
// helper so the bounding policy lives in exactly one place.
func previewOutput(s string, limit int) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		trimmed = "<empty>"
	}
	if len(trimmed) > limit {
		trimmed = trimmed[:limit] + "..."
	}
	return fmt.Sprintf("%q", trimmed)
}
