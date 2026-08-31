package main

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

// TestExitCode_Mapping verifies the error → process-exit-code mapping used by
// run(): success is 0, an *exitCodeError passes its code through unchanged, and
// any other error falls back to 2 with the message written to stderr.
func TestExitCode_Mapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  bool // whether the error should be written to stderr
	}{
		{
			name:     "success",
			err:      nil,
			wantCode: 0,
			wantMsg:  false,
		},
		{
			name:     "policy_failure_passthrough",
			err:      &exitCodeError{code: 1, err: errors.New("threshold violations detected")},
			wantCode: 1,
			wantMsg:  false, // command layer already reported it
		},
		{
			name:     "tool_failure_passthrough",
			err:      &exitCodeError{code: 2, err: errors.New("analyze failed")},
			wantCode: 2,
			wantMsg:  false,
		},
		{
			name:     "fallback_plain_error",
			err:      errors.New("boom"),
			wantCode: 2,
			wantMsg:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer
			got := exitCode(tt.err, &stderr)
			if got != tt.wantCode {
				t.Errorf("exitCode: got %d, want %d", got, tt.wantCode)
			}
			if tt.wantMsg && stderr.Len() == 0 {
				t.Error("expected error written to stderr, got none")
			}
			if !tt.wantMsg && stderr.Len() != 0 {
				t.Errorf("expected no stderr output, got %q", stderr.String())
			}
		})
	}
}

// TestRun_VersionExitsZero exercises run() end-to-end for the success path via
// the --version flag, which prints and returns a nil error (exit 0).
func TestRun_VersionExitsZero(t *testing.T) {
	// Not parallel: mutates the global os.Args, which run() reads via cobra.
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"vibe-check", "--version"}

	if code := run(); code != 0 {
		t.Errorf("run() with --version: got %d, want 0", code)
	}
}
