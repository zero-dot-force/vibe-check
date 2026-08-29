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
	if err := rootCmd().Execute(); err != nil {
		var ece *exitCodeError
		if errors.As(err, &ece) {
			os.Exit(ece.code)
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
