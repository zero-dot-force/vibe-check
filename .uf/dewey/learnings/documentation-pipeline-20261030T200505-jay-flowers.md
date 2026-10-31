---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:05:05Z
identity: documentation-pipeline-20261030T200505-jay-flowers
tier: draft
---

When a change alters a project's headline capability (supported languages, architecture-layer composition, roadmap phases), scope the tasks.md documentation task from an inventory of every section asserting facts about that capability — not just the project-structure listing. Adding internal/ripwireadapter/ to vibe-check makes the AGENTS.md Project Overview ("toolkit for Go codebases"), Architecture Layer-2 description (goadapter named as the only Layer-2 adapter), RFC phasing status (Python=P2, TS/JS=P3 future), and the README's opening line and Usage section ("a Go module directory") stale; a docs task scoped to "the project-structure section" plus a usage note leaves those narrative sections internally inconsistent when executed literally. The goreleaser-release-automation design's "Documentation Impact" section is the pattern to copy: enumerate all affected docs before writing the docs task.
