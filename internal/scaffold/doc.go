// Package scaffold embeds the vibe-check agent and command asset templates and
// deploys them into a target repository's .opencode/agents/ and
// .opencode/commands/ directories via `vibe-check init`.
//
// # Role
//
// scaffold is the implementation backing the `vibe-check init` command. It
// carries the canonical, version-controlled asset definitions — the
// divisor-entropy and vibe-check-reporter agents plus the /vibe-check slash
// command — as files embedded into the binary, then writes them into a consuming
// project so OpenCode can auto-discover them. Shipping the assets inside the
// same binary that runs `vibe-check analyze` keeps the deployed assets aligned
// with the metrics engine they drive.
//
// # Layout
//
//   - embed.go embeds assets/agents/*.md into an [embed.FS] exposed by
//     [AgentAssets] and assets/commands/*.md into an [embed.FS] exposed by
//     [CommandAssets].
//   - assets/agents/ holds the Markdown agent definitions.
//   - assets/commands/ holds the Markdown slash command definitions.
//   - scaffold.go deploys both asset categories via [Run], using a
//     [deployCategory] helper for each category.
//
// The embedded assets are the single source of truth: the copies deployed into a
// repository are generated from them rather than hand-authored, which prevents
// the deployed assets from drifting away from the version shipped in the binary.
//
// This package lives under internal/ because scaffolding is a CLI implementation
// detail that external modules must not import.
package scaffold
