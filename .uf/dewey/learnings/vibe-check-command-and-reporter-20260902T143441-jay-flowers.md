---
tag: vibe-check-command-and-reporter
author: jay-flowers
category: pattern
created_at: 2026-09-02T14:34:41Z
identity: vibe-check-command-and-reporter-20260902T143441-jay-flowers
tier: draft
---

When extending the vibe-check scaffold system to deploy multiple asset categories (agents and commands), the key architectural decision was extracting a deployCategory helper function that encapsulates the per-category deployment logic (glob, ensureDir, file iteration, prefix). The Run() function calls deployCategory for each category and merges Result slices. Category-prefixed Result paths (e.g., "agents/divisor-entropy.md", "commands/vibe-check.md") disambiguate entries from different categories. The refactoring preserved all existing security properties (symlink safety, containment checks, path validation) by reusing ensureDir and verifyContained for both deployment targets. Two separate //go:embed directives with separate embed.FS variables are required because Go's embed directive does not support multiple glob patterns in one directive for different directories.
