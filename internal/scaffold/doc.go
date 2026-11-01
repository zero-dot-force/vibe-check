// Package scaffold embeds the vibe-check agent and command asset templates
// and the CI workflow template, deploys the agent and command assets into a
// target repository's .opencode/agents/ and .opencode/commands/ directories,
// optionally deploys the CI workflow into .github/workflows/, and registers
// the deployed review agent into the target's Unbound Force policy files
// via `vibe-check init`.
//
// # Role
//
// scaffold is the implementation backing the `vibe-check init` command. It
// carries the canonical, version-controlled asset definitions — the
// divisor-entropy and vibe-check-reporter agents plus the /vibe-check slash
// command — as files embedded into the binary, then writes them into a consuming
// project so OpenCode can auto-discover them. After deployment it registers
// the divisor-entropy reviewer into the project's .uf/ policy files
// (reviewer-capabilities.yaml and review-matrix.yaml) with a byte-preserving
// splice, so review dispatch includes the entropy reviewer. Under
// `vibe-check init --ci` it additionally renders the embedded
// structural-gate caller workflow template with the validated release pin
// pair (tag plus commit SHA) and deploys it into the target's
// .github/workflows/ directory, skip-if-present because the file is user
// property once written. Shipping the assets inside the same binary that
// runs `vibe-check analyze` keeps the deployed assets aligned with the
// metrics engine they drive.
//
// # Layout
//
//   - embed.go embeds assets/agents/*.md into an [embed.FS] exposed by
//     [AgentAssets], assets/commands/*.md into an [embed.FS] exposed by
//     [CommandAssets], and assets/ci/*.yaml into an [embed.FS] exposed by
//     [CIAssets].
//   - assets/agents/ holds the Markdown agent definitions.
//   - assets/commands/ holds the Markdown slash command definitions.
//   - assets/ci/ holds the CI workflow templates deployed by init --ci.
//   - scaffold.go deploys both asset categories via [Run], using a
//     [deployCategory] helper for each category.
//   - ci.go validates the release pin pair, renders the caller workflow
//     template, deploys it skip-if-present, and detects the language
//     manifest at the target root.
//   - register.go holds the registration table and the guarded filesystem
//     orchestration that splices each registered review agent into the .uf/
//     policy files.
//   - splice.go holds the pure, byte-preserving content splices for the
//     policy files.
//   - lines.go holds the line-scanning helpers those splices share.
//
// The embedded assets are the single source of truth: the copies deployed into a
// repository are generated from them rather than hand-authored, which prevents
// the deployed assets from drifting away from the version shipped in the binary.
//
// This package lives under internal/ because scaffolding is a CLI implementation
// detail that external modules must not import.
package scaffold
