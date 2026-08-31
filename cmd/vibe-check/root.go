package main

import (
	"fmt"
	"runtime/debug"

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
		Version: versionString(),
	}

	cmd.AddCommand(analyzeCmd())

	return cmd
}

// versionString builds the --version output value in the form
// "<version> (commit <hash>, built <date>)". When the ldflags-injected version
// is empty or the default "dev" (e.g., the binary was installed via
// `go install github.com/zero-dot-force/vibe-check/cmd/vibe-check@vX`), it falls
// back to build information embedded by the Go toolchain: the main module
// version and the vcs.revision / vcs.time build settings. This ensures
// --version reports meaningful data even without explicit ldflags.
func versionString() string {
	v, c, d := version, commit, date
	if v == "" || v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if mv := info.Main.Version; mv != "" && mv != "(devel)" {
				v = mv
			}
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					if s.Value != "" {
						c = s.Value
					}
				case "vcs.time":
					if s.Value != "" {
						d = s.Value
					}
				}
			}
		}
	}
	return fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}
