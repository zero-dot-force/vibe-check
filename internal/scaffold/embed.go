package scaffold

import "embed"

//go:embed assets/agents/*.md
var agentAssetsFS embed.FS

//go:embed assets/commands/*.md
var commandAssetsFS embed.FS

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
