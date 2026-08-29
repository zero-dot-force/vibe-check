package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rootCmd creates the root cobra command for vibe-check.
func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vibe-check",
		Short: "Design quality and architectural metrics for Go codebases",
		Long: `vibe-check computes package-level coupling metrics (afferent/efferent coupling,
instability, abstractness, distance from main sequence), cohesion analysis,
and circular dependency detection for Go codebases.`,
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	}

	cmd.AddCommand(analyzeCmd())

	return cmd
}
