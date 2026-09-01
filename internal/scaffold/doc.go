// Package scaffold embeds the vibe-check agent asset templates and deploys them
// into a target repository's .opencode/agents/ directory via `vibe-check init`.
//
// # Role
//
// scaffold is the implementation backing the `vibe-check init` command. It
// carries the canonical, version-controlled agent definitions (currently the
// divisor-entropy Review Council agent) as files embedded into the binary, then
// writes them into a consuming project so the Review Council can auto-discover
// them. Shipping the assets inside the same binary that runs `vibe-check
// analyze` keeps the deployed agent aligned with the metrics engine it drives.
//
// # Layout
//
//   - embed.go embeds assets/agents/*.md into an [embed.FS] exposed by [Assets].
//   - assets/agents/ holds the Markdown agent definitions that are the single
//     source of truth for what `vibe-check init` deploys.
//
// The embedded asset is the single source of truth: the copy deployed into a
// repository is generated from it rather than hand-authored, which prevents the
// deployed agent from drifting away from the version shipped in the binary.
//
// This package lives under internal/ because scaffolding is a CLI implementation
// detail that external modules must not import.
package scaffold
