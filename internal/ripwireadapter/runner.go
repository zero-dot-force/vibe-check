package ripwireadapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// defaultMaxOutputBytes bounds the combined stdout the runner accepts from
// a single ripwire invocation. The bound is sized for the largest expected
// verb output (--export=cc.json on a large repository) while keeping a
// misbehaving binary from exhausting analyzer memory. Output at or below
// the bound is returned verbatim; anything larger is an error, never a
// silent truncation.
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
	// maxOutputBytes is the stdout capture bound; it exists as a field so
	// tests can shrink it. Production callers use the package default.
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
//   - stdout exceeding the capture bound returns an error naming the
//     bound — output is never truncated.
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
			r.binary, args, waitErr, previewBytes(stderr))
	}
	return stdout, nil
}

// previewBytes renders a bounded, quoted preview of subprocess stderr for
// error messages.
func previewBytes(b []byte) string {
	const maxPreview = 512
	s := string(b)
	if len(s) > maxPreview {
		s = s[:maxPreview] + "..."
	}
	if s == "" {
		s = "<empty>"
	}
	return fmt.Sprintf("%q", s)
}
