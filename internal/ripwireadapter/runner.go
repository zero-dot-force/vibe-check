package ripwireadapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

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

// ripwireRunner executes the resolved ripwire binary with the adapter's
// subprocess hygiene: a sanitized environment (no host credentials), a
// caller-provided working directory (the crawl root), bounded stdout and
// stderr capture, and context cancellation that terminates the subprocess.
type ripwireRunner struct {
	// binary is the PATH-resolved ripwire executable.
	binary string
	// maxOutputBytes is the per-stream capture bound applied to both the
	// stdout and the stderr pipe; it exists as a field so tests can shrink
	// it. Production callers use the package default.
	maxOutputBytes int64
}

// newRipwireRunner builds a runner for the given binary with the default
// output bound.
func newRipwireRunner(binary string) *ripwireRunner {
	return &ripwireRunner{binary: binary, maxOutputBytes: defaultMaxOutputBytes}
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
//     an error naming the bound and the stream — output is never truncated.
func (r *ripwireRunner) run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	// Fast path: refuse to start at all when the context is already dead.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("ripwire runner: %w", err)
	}

	cmd := exec.CommandContext(ctx, r.binary, args...)
	cmd.Dir = dir
	cmd.Env = ripwireEnv()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ripwire runner: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("ripwire runner: stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ripwire runner: start %s: %w", r.binary, err)
	}

	// Both pipes are read concurrently and bounded; the first reader to
	// cross the bound kills the process so the sibling reader unblocks
	// and Wait returns instead of deadlocking on a full pipe.
	var killOnce sync.Once
	kill := func() { _ = cmd.Process.Kill() }
	readBounded := func(pipe io.Reader, what string) ([]byte, bool, error) {
		buf := make([]byte, 0, 64*1024)
		chunk := make([]byte, 64*1024)
		for {
			n, readErr := pipe.Read(chunk)
			buf = append(buf, chunk[:n]...)
			if int64(len(buf)) > r.maxOutputBytes {
				killOnce.Do(kill)
				return buf, true, nil
			}
			if readErr == io.EOF {
				return buf, false, nil
			}
			if readErr != nil {
				return buf, false, fmt.Errorf("ripwire runner: reading %s: %w", what, readErr)
			}
		}
	}

	var stdout, stderr []byte
	var stdoutOversized, stderrOversized bool
	var stdoutErr, stderrErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdout, stdoutOversized, stdoutErr = readBounded(stdoutPipe, "stdout")
	}()
	go func() {
		defer wg.Done()
		stderr, stderrOversized, stderrErr = readBounded(stderrPipe, "stderr")
	}()
	wg.Wait()

	waitErr := cmd.Wait()

	// Cancellation wins over every other failure mode so callers can
	// reliably classify it with errors.Is(err, context.Canceled).
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("ripwire runner: %w (wait: %v)", ctxErr, waitErr)
	}
	if stdoutErr != nil {
		return nil, stdoutErr
	}
	if stderrErr != nil {
		return nil, stderrErr
	}
	// Oversize outranks the exit status: crossing the bound kills the
	// process ourselves, so the wait error is just our own kill signal.
	if stdoutOversized || stderrOversized {
		return nil, fmt.Errorf("%w: %s %v produced more than %d bytes of %s",
			errOutputOversized, r.binary, args, r.maxOutputBytes,
			map[bool]string{true: "stdout", false: "stderr"}[stdoutOversized])
	}
	if waitErr != nil {
		return nil, fmt.Errorf("ripwire runner: %s %v: %w; stderr: %s",
			r.binary, args, waitErr, previewOutput(string(stderr), maxStderrPreviewBytes))
	}
	return stdout, nil
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
