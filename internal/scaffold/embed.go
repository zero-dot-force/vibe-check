package scaffold

import "embed"

//go:embed assets/agents/*.md
var agentAssetsFS embed.FS

//go:embed assets/commands/*.md
var commandAssetsFS embed.FS

//go:embed assets/ci/*.yaml
var ciAssetsFS embed.FS

// AgentAssets returns the embedded filesystem containing the agent asset
// templates that vibe-check init deploys into a target repository's
// .opencode/agents/ directory.
func AgentAssets() embed.FS {
	return agentAssetsFS
}

// CommandAssets returns the embedded filesystem containing the command
// asset templates that vibe-check init deploys into a target repository's
// .opencode/commands/ directory.
func CommandAssets() embed.FS {
	return commandAssetsFS
}

// CIAssets returns the embedded filesystem containing the CI workflow
// templates that vibe-check init --ci deploys into a target repository's
// .github/workflows/ directory. Unlike the agent and command filesystems,
// CI templates have no .opencode/ mirror semantics (design D11): they are
// excluded from the categories table, assetPaths, and assetContent.
func CIAssets() embed.FS {
	return ciAssetsFS
}
