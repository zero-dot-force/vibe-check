// Package main provides the vibe-check CLI tool for computing design quality
// and architectural metrics for Go codebases.
//
// Usage:
//
//	vibe-check analyze [path]    Analyze Go packages and compute coupling metrics
//	vibe-check --version         Print version information
//
// The analyze command produces JSON output conforming to the ModuleGraph schema
// (version 1.1) and supports CI gate flags for threshold enforcement.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// version, commit, and date are set at build time via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// exitCodeError wraps an error with a specific process exit code.
// This allows RunE functions to communicate exit codes to main()
// without calling os.Exit directly, preserving deferred cleanup.
type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string { return e.err.Error() }

func (e *exitCodeError) Unwrap() error { return e.err }

func main() {
	os.Exit(run())
}

// run executes the root command and returns the process exit code. It is
// separated from main so the exit-code mapping is unit-testable; main itself
// only calls os.Exit.
func run() int {
	return exitCode(rootCmd().Execute(), os.Stderr)
}

// exitCode maps a command error to a process exit code and returns the
// corresponding integer:
//
//	nil error       → 0 (success)
//	*exitCodeError  → its carried code (e.g. 1 for policy failures)
//	any other error → 2 (tool failure), with the error written to stderr
//
// Errors carried by *exitCodeError are already reported to stderr by the
// command layer, so they are not re-printed here.
func exitCode(err error, stderr io.Writer) int {
	if err == nil {
		return 0
	}
	var ece *exitCodeError
	if errors.As(err, &ece) {
		return ece.code
	}
	_, _ = fmt.Fprintln(stderr, err)
	return 2
}
