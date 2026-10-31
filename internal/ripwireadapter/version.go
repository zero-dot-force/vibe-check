package ripwireadapter

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// The ripwire surface this adapter is validated against. The pin is
// fail-closed: any binary whose reported version or commit does not match
// these constants is refused at Analyze time, because every parsing,
// mapping, and golden decision in this package was made against this
// exact output shape.
const (
	// pinnedRipwireVersion is the validated ripwire release version.
	pinnedRipwireVersion = "0.6.9"
	// pinnedRipwireCommit is the validated commit prefix (built_from=).
	pinnedRipwireCommit = "bedc0cd"
)

// ripwireInstallURL names the install location carried by the absent-binary
// guidance. It is also asserted by the CLI's absent-binary path.
const ripwireInstallURL = "https://github.com/redhat-et/ripwire"

// maxVersionOutputBytes bounds the --version output the adapter is willing
// to read. Output at or below this bound is parsed; anything larger is
// treated as unparseable rather than truncated, so a misbehaving binary on
// PATH can never feed the pin check a partial line.
const maxVersionOutputBytes = 4096

// Version-pin sentinel errors. They allow callers (including the CLI's
// absent-binary path) to classify failures with errors.Is while the wrapped
// messages carry the full guidance context.
var (
	// ErrRipwireAbsent reports that no ripwire binary was found on PATH.
	ErrRipwireAbsent = errors.New("ripwire binary not found on PATH")
	// ErrRipwireVersionMismatch reports a binary whose version or commit
	// does not match the pinned validated pair.
	ErrRipwireVersionMismatch = errors.New("ripwire version mismatch")
	// ErrRipwireVersionUnparseable reports version output that does not
	// match the expected "ripwire <version> (... built_from=<commit>)"
	// shape, is empty, or exceeds the size bound.
	ErrRipwireVersionUnparseable = errors.New("ripwire version output unparseable")
)

// ripwireVersionInfo is the (version, commit) pair parsed from a ripwire
// --version line.
type ripwireVersionInfo struct {
	version string
	commit  string
}

// matchesPin reports whether the parsed info equals the pinned version and
// carries a commit that extends the pinned commit prefix.
func (v ripwireVersionInfo) matchesPin() bool {
	return v.version == pinnedRipwireVersion && strings.HasPrefix(v.commit, pinnedRipwireCommit)
}

// absentBinaryError constructs the shared absent-binary guidance error. The
// CLI's absent-binary path (exit code 2) surfaces this message verbatim, so
// it must always name ripwire, the install location, and the validated pair.
func absentBinaryError(lookErr error) error {
	return fmt.Errorf(
		"%w: %v; install the validated ripwire %s (built from commit %s) from %s and ensure it is on PATH",
		ErrRipwireAbsent, lookErr, pinnedRipwireVersion, pinnedRipwireCommit, ripwireInstallURL,
	)
}

// checkRipwirePin resolves the ripwire binary on PATH and validates its
// reported version and commit against the pinned pair. It runs at Analyze
// time (never at construction) so an adapter instance never captures stale
// PATH state. The --version subprocess runs through the shared runner with
// a sanitized environment (no host credentials) and a tight output bound.
func checkRipwirePin(ctx context.Context) error {
	_, _, err := validatedRipwire(ctx)
	return err
}

// validatedRipwire resolves the ripwire binary on PATH, validates its
// reported version and commit against the pinned pair, and returns both
// the binary path and the validated version info. Analyze uses the
// returned pair for its subprocess runner and the source-tool provenance
// fields (design D11): since the pin fails closed, only the validated
// (version, commit) pair can ever appear in output.
func validatedRipwire(ctx context.Context) (string, ripwireVersionInfo, error) {
	bin, err := exec.LookPath("ripwire")
	if err != nil {
		return "", ripwireVersionInfo{}, absentBinaryError(err)
	}

	runner := &ripwireRunner{binary: bin, maxOutputBytes: maxVersionOutputBytes}
	stdout, err := runner.run(ctx, "", "--version")
	if err != nil {
		return "", ripwireVersionInfo{}, fmt.Errorf("ripwire version check: %w", err)
	}

	info, err := parseRipwireVersion(stdout)
	if err != nil {
		return "", ripwireVersionInfo{}, err
	}
	if !info.matchesPin() {
		return "", ripwireVersionInfo{}, fmt.Errorf(
			"%w: pinned ripwire %s @ %s, found %s @ %s",
			ErrRipwireVersionMismatch, pinnedRipwireVersion, pinnedRipwireCommit, info.version, info.commit,
		)
	}
	return bin, info, nil
}

// parseRipwireVersion parses a ripwire --version line of the validated
// shape:
//
//	ripwire 0.6.9 (Release, AppleClang 21.0.0.21000101, emit=std::print, built_from=bedc0cdcf)
//
// The version is the second whitespace-separated token; the commit is the
// value of the trailing built_from= attribute. Compiler and emit fields in
// the parenthesized tail are tolerated but ignored. Anything else — empty
// output, missing tokens, a missing built_from=, or output exceeding the
// size bound — fails closed as unparseable.
func parseRipwireVersion(output []byte) (ripwireVersionInfo, error) {
	if len(output) > maxVersionOutputBytes {
		return ripwireVersionInfo{}, fmt.Errorf(
			"%w: version output exceeds %d bytes", ErrRipwireVersionUnparseable, maxVersionOutputBytes,
		)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return ripwireVersionInfo{}, fmt.Errorf(
			"%w: %s", ErrRipwireVersionUnparseable, previewVersionOutput(output),
		)
	}
	info := ripwireVersionInfo{version: fields[1]}
	for _, field := range fields[2:] {
		if commit, ok := strings.CutPrefix(field, "built_from="); ok {
			// The built_from attribute closes the parenthesized tail,
			// so its token carries a trailing ')' in the recorded
			// output shape.
			info.commit = strings.TrimRight(commit, ")")
		}
	}
	if info.commit == "" {
		return ripwireVersionInfo{}, fmt.Errorf(
			"%w: no built_from= commit in %s", ErrRipwireVersionUnparseable, previewVersionOutput(output),
		)
	}
	return info, nil
}

// previewVersionOutput renders a bounded preview of unparseable version
// output for error messages, so a hostile or corrupt binary cannot flood
// the error (and any log it lands in) with arbitrary bytes.
func previewVersionOutput(output []byte) string {
	const maxPreview = 120
	s := strings.TrimSpace(string(output))
	if len(s) > maxPreview {
		s = s[:maxPreview] + "..."
	}
	if s == "" {
		s = "<empty>"
	}
	return fmt.Sprintf("%q", s)
}
