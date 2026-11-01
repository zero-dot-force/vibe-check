---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-31T19:30:51Z
identity: documentation-pipeline-20261031T193051-jay-flowers
tier: draft
---

Vibe-Check's documentation gates (the AGENTS.md documentation gate and the constitution's Website Documentation Sync) enumerate CHANGELOG.md, AGENTS.md, README.md, and issue filing — but not the CLI's embedded cobra help text (rootCmd/analyzeCmd Short/Long and flag help), a user-facing docs surface with no gate owner. Behavior-expanding changes can land with repo docs and issue tracking fully satisfied while the command-layer help goes stale: the ripwire-adapter change (PR #75, opsx/ripwire-adapter) made `vibe-check analyze` auto-detect Python/TypeScript and route to ripwire-backed adapters, updated README/CHANGELOG/AGENTS per docs issue #70, yet analyzeCmd still reads "Analyze Go packages and compute coupling metrics" and rootCmd "for Go codebases" because the diff touched only RunAnalyze's routing block. Documentation-impact checklists for CLI behavior changes must include the command layer's Short/Long/flag help as a fourth surface; the Scribe owns CLI help content, but only once the Curator triages the gap.
