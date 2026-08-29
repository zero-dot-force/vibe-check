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
	"fmt"
	"os"
)

// version, commit, and date are set at build time via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
