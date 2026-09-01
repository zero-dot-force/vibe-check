package scaffold

import "embed"

//go:embed assets/agents/*.md
var assetsFS embed.FS

// Assets returns the embedded filesystem containing the agent asset
// templates that vibe-check init deploys into a target repository.
func Assets() embed.FS {
	return assetsFS
}
